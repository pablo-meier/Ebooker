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
	"time"
)

// Ebooker is the provider of the service, and maintains its internal resources
// in this struct.
type Ebooker struct {
	logger *logging.LogMaster
	data   *DataHandle
	oauth  *oauth1.OAuth1
	tf     *TweetFetcher
}

type Bot struct {
	username string
	sources  []string
	gen      *Generator
	token    *oauth1.Token
	sched    *Schedule

	ebooker *Ebooker
}

const DEFAULT_USER = "SrPablo"

const applicationKey = "MxIkjx9eCC3j1JC8kTig"
const applicationSecret = "IgOkwoh5m7AS4LplszxcPaF881vjvZYZNCAvvUz1x0"

// Starts the service
func main() {

	//	var silent bool
	var debug, timestamps bool
	var port string
	//	flag.BoolVar(&silent, "silent", true, "Generate only the tweets, without other status information.")
	flag.BoolVar(&debug, "debug", false, "Print debugging information.")
	flag.BoolVar(&timestamps, "timestamps", false, "Print log/debug with timestamps.")
	flag.StringVar(&port, "port", "8998", "Port to run the server on.")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	// Silent default to false, since there isn't really an aesthetic need to do so
	logger := logging.GetLogMaster(false, debug, timestamps)
	dh := getDataHandle("./ebooker_tweets.db", &logger)
	defer dh.Cleanup()
	oauth1 := oauth1.CreateOAuth1(&logger, applicationKey, applicationSecret)
	tf := getTweetFetcher(&logger, &oauth1)

	logger.StatusWrite("Welcome to EBOOKER -- let's make some nonsense ^_^\n")
	logger.StatusWrite("Registering Ebooker RPC..\n")
	eb := Ebooker{&logger, &dh, &oauth1, &tf}
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

	eb.logger.StatusWrite("Creating a generator...\n", user)
	gen, err := eb.createSeededGenerator(&args.Gen)
	if err != nil {
		*out = "fail"
		eb.logger.DebugWrite("Generator creation failed. Error: %v\n", err)
		return err
	}

	schedule := cronParse(args.Sched.Cron)
	bot := Bot{user, args.Gen.Users, gen, token, &schedule, eb}
	*out = "The next tweet will arrive at: " + schedule.next().String()
	eb.logger.StatusWrite("Bot created! %s\n", *out)
	go bot.Run()
	return nil
}

// Runs perpetually, forever tweeting
func (b *Bot) Run() {
	b.ebooker.logger.StatusWrite("Bot %s ordered to run! Away we go!\n", b.username)

	c := b.sched.tickingChannel()
	for _ = range c {
		if b.sched.shouldKill() {
			b.ebooker.logger.StatusWrite("Bot %s received killing order! Dying...\n", b.username)
			break
		}
		b.ebooker.logger.StatusWrite("At %v bot %s received the order to tweet.\n", time.Now(), b.username)

		// Update Sources
		newSources := b.ebooker.fetchNewSources(b.sources, b.token)

		// Add new seeds to generator.
		for _, val := range newSources {
			b.gen.AddSeeds(val)
		}

		// fire off the new tweet
		message := b.gen.GenerateText()
		b.ebooker.logger.StatusWrite("Sending \"%s\"\n", message)
		b.ebooker.tf.sendTweet(message, b.token)
		b.ebooker.logger.StatusWrite("Success! Next tweet due after %v\n", b.sched.next().String())
	}
}

func (b *Bot) Kill() {
	b.sched.kill()
}

func (eb *Ebooker) createSeededGenerator(args *defs.GenParams) (*Generator, error) {

	token := eb.getAccessToken(DEFAULT_USER)
	sourcestrings := eb.fetchNewSources(args.Users, token)

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

func (eb *Ebooker) fetchNewSources(userlist []string, token *oauth1.Token) []string {

	var sourcestrings []string
	for _, username := range userlist {
		// get tweets from persistent storage
		eb.logger.StatusWrite("Reading from persistent storage for %s...\n", username)
		oldTweets := eb.data.GetTweetsFromStorage(username)

		var newTweets Tweets
		if len(oldTweets) == 0 {
			eb.logger.StatusWrite("Found no tweets for %s, doing a deep dive to retrieve their history.\n", username)
			newTweets = eb.tf.DeepDive(username, token)
		} else {
			eb.logger.StatusWrite("Found %d tweets for %s.\n", len(oldTweets), username)
			newest := oldTweets[len(oldTweets)-1]
			newTweets = eb.tf.GetRecentTimeline(username, &newest, token)
		}

		// update the persistent storage
		eb.logger.StatusWrite("Inserting %d new tweets into persistent storage.\n", len(newTweets))
		eb.data.InsertFreshTweets(username, newTweets)

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
