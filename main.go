/*
Command-line client that uses the janky ebooker package to generate tweets
easily. See --help for all the settings.
*/
package main

import (
    "ebooker"

    "errors"
    "flag"
    "log"
    "math/rand"
    "net"
    "net/rpc"
    "net/http"
    "time"
)

// Starts the service
func main() {

    var silent, debug, timestamps bool
    flag.BoolVar(&silent, "silent", true, "Generate only the tweets, without other status information.")
    flag.BoolVar(&debug, "debug", false, "Print debugging information.")
    flag.BoolVar(&timestamps, "timestamps", false, "Print log/debug with timestamps.")
    flag.Parse()

    rand.Seed(time.Now().UnixNano())

    logger := ebooker.GetLogMaster(silent, debug, timestamps)
    dh := ebooker.GetDataHandle("./ebooker_tweets.db", &logger)
    defer dh.Cleanup()

    ebRequest := EbookerRequest{ &logger, &dh }
    rpc.Register(&ebRequest)
    rpc.HandleHTTP()
    l, e := net.Listen("tcp", ":1234")
    if e != nil {
        log.Fatal("listen error:", e)
    }
    http.Serve(l, nil)
}


type EbookerRequest struct {
    logger *ebooker.LogMaster
    dh     *ebooker.DataHandle
}



type Tweets []string
type GenerateTweetsArgs struct {
    users []string
    numTweets int
    reps bool
    prefixLen int
}

func (eb *EbookerRequest) GenerateTweets(args *GenerateTweetsArgs, out *Tweets) error {

    var sourcestrings []string
    for _, username := range args.users {
        // get tweets from persistent storage
        eb.logger.StatusWrite("Reading from persistent storage for %s...\n", username)
        oldTweets := eb.dh.GetTweetsFromStorage(username)

        tf := ebooker.GetTweetFetcher(eb.logger, eb.dh)

        var newTweets ebooker.Tweets
        if len(oldTweets) == 0 {
            eb.logger.StatusWrite("Found no tweets for %s, doing a deep dive to retrieve their history.\n", username)
            newTweets = tf.DeepDive(username)
        } else {
            eb.logger.StatusWrite("Found %d tweets for %s.\n", len(oldTweets), username)
            newest := oldTweets[len(oldTweets) - 1]
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
        *out = Tweets{}
        return errors.New("No text for users in list. Either unauthorized, or they don't exist")
    }

	// fetch or create a Generator
	gen := ebooker.CreateGenerator(args.prefixLen, 140, eb.logger)
	if args.reps {
		gen.CanonicalizeSources()
	}

	// Seed the Generator
	for _, str := range sourcestrings {
		gen.AddSeeds(str)
	}

    eb.logger.StatusWrite("Outputting nonsense tweets for \"%v\":\n", args.users)

    tweets := make(Tweets, args.numTweets)
    for i := 0; i < args.numTweets; i++ {
        tweets = append(tweets, gen.GenerateText())
    }
    *out = tweets
    return nil
}


func copyFrom(dst *[]string, src *ebooker.Tweets) {
	for _, str := range *src {
		*dst = append(*dst, str.Text)
	}
}
