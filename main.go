package main

import (
    "ebooker"

    "flag"
    "fmt"
)

func main() {

    var username string
    var prefixLen, numTweets int
    var silent bool

    flag.StringVar(&username, "user", "laurelita", "Twitter user to base output off of.")
    flag.IntVar(&prefixLen, "prefixLength", 1, "length of prefix")
    flag.IntVar(&numTweets, "numTweets", 50, "the number of tweets to produce")
    flag.BoolVar(&silent, "silent", true, "Generate only the tweets, without other status information.")

    flag.Parse()

    // get tweets from persistent storage
    statusMsg("Reading from persistent storage...\n", silent)
    oldTweets := ebooker.GetTweetsFromStorage(username)

    var newTweets []ebooker.TweetData
    if len(oldTweets) == 0 {
        statusMsg(fmt.Sprintf("Found no tweets for %s, doing a deep dive to retrieve their history.\n", username), silent)
        newTweets = ebooker.DeepDive(username)
    } else {
        statusMsg(fmt.Sprintf("Found %d tweets for %s.\n", len(oldTweets), username), silent)
        newTweets = ebooker.GetTimelineFromRequest(username, &(oldTweets[0]))
    }

    // update the persistent storage
    statusMsg(fmt.Sprintf("Inserting %d new tweets into persistent storage.\n", len(newTweets)), silent)
    ebooker.InsertFreshTweets(newTweets)

    // fetch Generator from datastore
    gen := ebooker.CreateGenerator(prefixLen, 140)

    // Seed the Generator
    for i := range oldTweets {
        gen.AddSeeds(oldTweets[i].Text)
    }
    for i := range newTweets {
        gen.AddSeeds(newTweets[i].Text)
    }

    // Generate some faux tweets. Print them!
    for i := 0; i < numTweets; i++ {
        statusMsg(fmt.Sprintf("For %s:\n", username), silent)
        fmt.Printf("-------------\n%s\n", gen.GenerateText())
    }
}

func statusMsg(s string, silent bool) {
    if !silent {
        fmt.Printf(s)
    }
}
