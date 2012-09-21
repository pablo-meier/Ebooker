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
    "time"
)

func main() {

    // Parse flags, initialize data...
    var username string
    var prefixLen, numTweets int
    var silent, debug, reps, timestamps bool

    flag.StringVar(&username, "user", "laurelita", "Twitter user to base output off of.")
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


    // get tweets from persistent storage
    logger.StatusWrite("Reading from persistent storage...\n")
    oldTweets := dh.GetTweetsFromStorage(username)

    tf := ebooker.GetTweetFetcher(&logger)

    var newTweets []ebooker.TweetData
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

    // fetch or create a Generator
    gen := ebooker.CreateGenerator(prefixLen, 140, &logger)
    if reps {
        gen.CanonicalizeSources()
    }

    // Seed the Generator
    for _, tweet := range oldTweets {
        gen.AddSeeds(tweet.Text)
    }
    for _, tweet := range newTweets {
        gen.AddSeeds(tweet.Text)
    }

    // Generate some faux tweets. Print them!
    logger.StatusWrite("Outputting nonsense tweets for %s:\n", username)
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

