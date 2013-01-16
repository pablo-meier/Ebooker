/*
Command that launches a service that responds to requests with ebooking
capabilities.
*/
package main

import (
	"ebooker/defs"
	"ebooker/logging"
	"ebooker/oauth1"

	"errors"
	"flag"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"time"
)

// Ebooker is the provider of the service, and maintains its internal resources
// in this struct.
type Ebooker struct {
	bots map[string]*Bot

	logger *logging.LogMaster
	data   *DataHandle
	oauth  *oauth1.OAuth1
	tf     *TweetFetcher
}

const DEFAULT_USER = "SrPablo"

// Starts the service
func main() {
	var debug, timestamps, silent bool
	var port, keyFile string
	flag.BoolVar(&silent, "silent", false, "Generate only the tweets, without other status information.")
	flag.BoolVar(&debug, "debug", false, "Print debugging information.")
	flag.BoolVar(&timestamps, "timestamps", false, "Print log/debug with timestamps.")
	flag.StringVar(&port, "port", "8998", "Port to run the server on.")
	flag.StringVar(&keyFile, "keyfile", "keys.txt", "File containing the application keys assigned to you by Twitter.")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	// Silent default to false, since there isn't really an aesthetic need to do so
	logger := logging.GetLogMaster(silent, debug, timestamps)
	dh := getDataHandle("./ebooker_tweets.db", &logger)
	defer dh.Cleanup()
	applicationKey, applicationSecret := oauth1.ParseFromFile(keyFile)
	oauth1 := oauth1.CreateOAuth1(&logger, applicationKey, applicationSecret)
	tf := getTweetFetcher(&logger, &oauth1)
	bots := make(map[string]*Bot)

	logger.StatusWrite("Welcome to EBOOKER -- let's make some nonsense ^_^\n")
	logger.StatusWrite("Registering Ebooker RPC...\n")

	eb := Ebooker{bots, &logger, &dh, &oauth1, &tf}
	rpc.Register(&eb)
	rpc.HandleHTTP()

	logger.StatusWrite("Starting up on port %s\n", port)
	l, e := net.Listen("tcp", ":"+port)
	if e != nil {
		logger.StatusWrite("Listen error: %v.\nTerminating...", e)
		os.Exit(1)
	}
	http.Serve(l, nil)
}

// GenerateTweets is the core service: given a set of arguments (namely the
// Twitter user(s) in question), generate a bunch of Markovian Tweets.
func (eb *Ebooker) GenerateTweets(args *defs.GenParams, out *defs.Tweets) error {
	gen, err := eb.createSeededGenerator(args)
	if err != nil {
		*out = defs.Tweets{}
		return err
	}

	eb.logger.StatusWrite("Outputting nonsense tweets for \"%v\":\n", args.Users)
	tweets := make(defs.Tweets, args.NumTweets)
	for i := 0; i < args.NumTweets; i++ {
		tweets[i] = gen.GenerateText()
	}
	*out = tweets
	return nil
}

// NewBot takes parameters needed to create a self-tweeting, perpetual bot,
// sets it running on the server.
func (eb *Ebooker) NewBot(args *defs.NewBotParams, out *string) error {

	user := args.Auth.User
	eb.logger.StatusWrite("Creating a new bot for %v\n", user)
	token, exists := eb.data.getUserAccessToken(user)
	if !exists {
		eb.logger.StatusWrite("%v does not have credentials in the database. Adding...\n", user)
		token := &oauth1.Token{args.Auth.Token, args.Auth.TokenSecret}
		eb.data.insertUserAccessToken(user, token)
	}

	eb.logger.StatusWrite("Creating a generator...\n")
	gen, err := eb.createSeededGenerator(&args.Gen)
	if err != nil {
		*out = "fail"
		eb.logger.DebugWrite("Generator creation failed. Error: %v\n", err)
		return err
	}

	schedule := cronParse(args.Sched.Cron)
	bot := Bot{user, args.Gen.Users, gen, token, &schedule,
		eb.logger, eb.data, eb.oauth, eb.tf}

	eb.bots[user] = &bot
	*out = "The next tweet will arrive at: " + schedule.next().String()
	eb.logger.StatusWrite("Bot created! %s\n", *out)
	go bot.Run()
	return nil
}

// Lists the bots this Ebooker server is running.
func (eb *Ebooker) ListBots(_ string, out *[]string) error {

	for k, v := range eb.bots {
		botstring := k + ":" + strings.Join(v.sources, ",")
		*out = append(*out, botstring)
	}

	return nil
}

// Cancels this bot, preventing it from tweeting.
func (eb *Ebooker) CancelBot(name string, out *string) error {

	bot, exists := eb.bots[name]
	if !exists {
		*out = ""
		return errors.New("No bot found for that name.")
	}

	bot.Kill()
	*out = name + " now inactive. You can always start it up again later ^_^"
	return nil
}

// Cancels this bot, preventing it from tweeting.
func (eb *Ebooker) DeleteBot(name string, out *string) error {

	bot, exists := eb.bots[name]
	if !exists {
		*out = ""
		return errors.New("No bot found for that name.")
	}

	bot.Kill()
	delete(eb.bots, name)
	*out = name + " gone!"
	return nil
}

func (eb *Ebooker) createSeededGenerator(args *defs.GenParams) (*Generator, error) {

	appToken := eb.oauth.MakeToken()
	sourcestrings := fetchNewSources(args.Users, appToken, eb.data, eb.logger, eb.tf)

	if len(sourcestrings) == 0 {
		eb.logger.StatusWrite("Can't write nonsense tweets, as we don't have a corpus!\n")
		noTextError := errors.New("No text for users in list. Either unauthorized, or they don't exist")
		return nil, noTextError
	}

	// fetch or create a Generator
	gen := CreateGenerator(args.PrefixLen, 140, eb.logger)
	if args.Reps {
		gen.CanonicalizeSources()
	}

	// Seed the Generator
	for _, str := range sourcestrings {
		gen.AddSeeds(str)
	}

	return gen, nil
}

// fetchNewSources will check the Twitter API for new tweets by 'sources,' using the
// authentication from 'token.' Note that we'd like this to be a member function of some
// struct interface "ResourceHolder," but Goobuntu + GBus Wifi are so craptacularly out
// of sync that I have to do this one offline, and can't look up whether we could even
// attach an implementation to an interface in Go, or what that would look like.
//
// Feel my first 'rants' email coming along...
func fetchNewSources(userlist []string, token *oauth1.Token, data *DataHandle, logger *logging.LogMaster, tf *TweetFetcher) []string {

	var sourcestrings []string
	for _, username := range userlist {
		// get tweets from persistent storage
		logger.StatusWrite("Reading from persistent storage for %s...\n", username)
		oldTweets := data.GetTweetsFromStorage(username)

		var newTweets Tweets
		if len(oldTweets) == 0 {
			logger.StatusWrite("Found no tweets for %s, doing a deep dive to retrieve their history.\n", username)
			newTweets = tf.DeepDive(username, token)
		} else {
			logger.StatusWrite("Found %d tweets for %s.\n", len(oldTweets), username)
			newest := oldTweets[len(oldTweets)-1]
			newTweets = tf.GetRecentTimeline(username, &newest, token)
		}

		// update the persistent storage
		logger.StatusWrite("Inserting %d new tweets into persistent storage.\n", len(newTweets))
		data.InsertFreshTweets(username, newTweets)

		copyFrom(&sourcestrings, &oldTweets)
		copyFrom(&sourcestrings, &newTweets)
	}
	return sourcestrings
}

func copyFrom(dst *[]string, src *Tweets) {
	for _, str := range *src {
		*dst = append(*dst, str.Text)
	}
}

// Returns the access token we have in storage for the user. If the user doesn't
// exist, we return an error.
func (eb *Ebooker) getAccessToken(user string) *oauth1.Token {
	accessToken, exists := eb.data.getUserAccessToken(user)

	if !exists {
		eb.logger.StatusWrite("Access token for %v not present! Beginning OAuth...\n", user)
		requestToken := eb.oauth.ObtainRequestToken()
		token := eb.oauth.ObtainAccessToken(requestToken)

		eb.data.insertUserAccessToken(user, token)
		accessToken = token
	}
	return accessToken
}
