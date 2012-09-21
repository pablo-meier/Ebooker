/*
Command-line client that uses the janky ebooker package to generate tweets
easily. See --help for all the settings.
*/
package main

import (
    "ebooker"

    "flag"
    "fmt"
    "math/rand"
    "strings"
    "time"
)

func main() {

    // Parse flags, initialize data...
    var userlist string
    var prefixLen, numTweets int
    var silent, debug, reps, timestamps bool

    flag.StringVar(&userlist, "users", "laurelita", "Comma-separated list of twitter users to base output off of.")
    flag.IntVar(&prefixLen, "prefixLength", 1, "length of prefix")
    flag.IntVar(&numTweets, "numTweets", 50, "the number of tweets to produce")
    flag.BoolVar(&silent, "silent", true, "Generate only the tweets, without other status information.")
    flag.BoolVar(&reps, "representations", false, "Treat various representations (e.g. \"Its/it's/IT'S\") as the same in generation.")
    flag.BoolVar(&debug, "debug", false, "Print debugging information.")
    flag.BoolVar(&timestamps, "timestamps", false, "Print log/debug with timestamps.")


    flag.Parse()
    rand.Seed(time.Now().UnixNano())

    logger := ebooker.GetLogMaster(silent, debug, timestamps)
    dh := ebooker.GetDataHandle("./ebooker_tweets.db", &logger)
    defer dh.Cleanup()

    users := strings.Split(userlist, ",")
    var sourcestrings[]string

    for _, username := range users {
        // get tweets from persistent storage
        logger.StatusWrite("Reading from persistent storage for %s...\n", username)
        oldTweets := dh.GetTweetsFromStorage(username)

        tf := ebooker.GetTweetFetcher(&logger)

        var newTweets ebooker.Tweets
        if len(oldTweets) == 0 {
            logger.StatusWrite("Found no tweets for %s, doing a deep dive to retrieve their history.\n", username)
            newTweets = tf.DeepDive(username)
        } else {
            logger.StatusWrite("Found %d tweets for %s.\n", len(oldTweets), username)
            newest := oldTweets[len(oldTweets) - 1]
            newTweets = tf.GetTimelineFromRequest(username, &newest)
        }

        // update the persistent storage
        logger.StatusWrite("Inserting %d new tweets into persistent storage.\n", len(newTweets))
        dh.InsertFreshTweets(username, newTweets)

		copyFrom(&sourcestrings, &oldTweets)
		copyFrom(&sourcestrings, &newTweets)
    }

	// fetch or create a Generator
	gen := ebooker.CreateGenerator(prefixLen, 140, &logger)
	if reps {
		gen.CanonicalizeSources()
	}

	// Seed the Generator
	for _, str := range sourcestrings {
		gen.AddSeeds(str)
	}

    // Generate some faux tweets. Print them!
    logger.StatusWrite("Outputting nonsense tweets for \"%v\":\n", userlist)
    var format string
    if silent {
        format = "%s\n"
    } else {
        format = "  %s\n"
    }

    for i := 0; i < numTweets; i++ {
        fmt.Printf(format, gen.GenerateText())
    }
}


func copyFrom(dst *[]string, src *ebooker.Tweets) {
	for _, str := range *src {
		*dst = append(*dst, str.Text)
	}
}
