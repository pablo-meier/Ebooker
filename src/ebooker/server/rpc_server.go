/*
Command that launches a service that responds to requests with ebooking 
capabilities.
*/
package main

import (
	"ebooker/defs"
	"ebooker/ebooks"

	"errors"
	"flag"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

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
	logger := ebooks.GetLogMaster(false, debug, timestamps)
	dh := ebooks.GetDataHandle("./ebooker_tweets.db", &logger)
	defer dh.Cleanup()

	logger.StatusWrite("Welcome to EBOOKER -- let's make some nonsense ^_^\n")
	logger.StatusWrite("Registering Ebooker RPC..\n")
	eb := Ebooker{&logger, &dh}
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

// Ebooker is the provider of the service, and maintains its internal resources
// in this struct.
type Ebooker struct {
	logger *ebooks.LogMaster
	dh     *ebooks.DataHandle
}

// GenerateTweets is the core service: given a set of arguments (namely the
// Twitter user(s) in question), generate a bunch of Markovian Tweets.
func (eb *Ebooker) GenerateTweets(args *defs.GenParams, out *defs.Tweets) error {

	var sourcestrings []string
	for _, username := range args.Users {
		// get tweets from persistent storage
		eb.logger.StatusWrite("Reading from persistent storage for %s...\n", username)
		oldTweets := eb.dh.GetTweetsFromStorage(username)

		tf := ebooks.GetTweetFetcher(eb.logger, eb.dh)

		var newTweets ebooks.Tweets
		if len(oldTweets) == 0 {
			eb.logger.StatusWrite("Found no tweets for %s, doing a deep dive to retrieve their history.\n", username)
			newTweets = tf.DeepDive(username)
		} else {
			eb.logger.StatusWrite("Found %d tweets for %s.\n", len(oldTweets), username)
			newest := oldTweets[len(oldTweets)-1]
			newTweets = tf.GetRecentTimeline(username, &newest)
		}

		// update the persistent storage
		eb.logger.StatusWrite("Inserting %d new tweets into persistent storage.\n", len(newTweets))
		eb.dh.InsertFreshTweets(username, newTweets)

		copyFrom(&sourcestrings, &oldTweets)
		copyFrom(&sourcestrings, &newTweets)
	}

	if len(sourcestrings) == 0 {
		eb.logger.StatusWrite("Can't write nonsense tweets, as we don't have a corpus!\n")
		noTextError := errors.New("No text for users in list. Either unauthorized, or they don't exist")
		*out = defs.Tweets{}
		return noTextError
	}

	// fetch or create a Generator
	gen := ebooks.CreateGenerator(args.PrefixLen, 140, eb.logger)
	if args.Reps {
		gen.CanonicalizeSources()
	}

	// Seed the Generator
	for _, str := range sourcestrings {
		gen.AddSeeds(str)
	}

	eb.logger.StatusWrite("Outputting nonsense tweets for \"%v\":\n", args.Users)

	tweets := make(defs.Tweets, args.NumTweets)
	for i := 0; i < args.NumTweets; i++ {
		tweets[i] = gen.GenerateText()
	}
	*out = tweets
	return nil
}

func copyFrom(dst *[]string, src *ebooks.Tweets) {
	for _, str := range *src {
		*dst = append(*dst, str.Text)
	}
}
