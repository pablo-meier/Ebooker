package main

import (
	"ebooker/logging"
	"launchpad.net/gocheck"
)

// hook up gocheck into the gotest runner.
type StorageSuite struct{}

var _ = gocheck.Suite(&StorageSuite{})

// Simple test case, where we acquire a handle, save some tweets to it, and retrieve them.
func (s StorageSuite) TestStorageFunctionality(c *gocheck.C) {

	dh := getDataHandle("./ebooker_tweets.db", &logging.LogMaster{})
	defer dh.Cleanup()

	pabloTweets := []TweetData{TweetData{398273498291123, "Just got an email whose only contents were \"LOL\". The day is won."},
		TweetData{398273498291124, "@Popehat When I was 8 and asked my dad what his job was, he confused me with \"I'm a transaction cost.\""},
		TweetData{398273498291125, "Cautionary tales on type system design and notation, brought to you by the letter 'Scala'"},
		TweetData{398273498291126, "@laurelita ... I'm pretty sure my key comes with a free invite for a friend ^_- Will investigate!"},
		TweetData{398273498291127, "@jseakle @SkunkFunk927 Sweet, I'm \"sicp\" in LoL. Someon seems to have taken \"SrPablo.\" I play LoL least, but would rather play with you ^_^"},
		TweetData{398273498291129, "@SkunkFunk927 d'you ever play with your sister? She's pretty sick with that Tibbers-bearing girl (BEARING LOLOLOLLOLO). I go Singed or Ryze"}}

	laurenTweets := []TweetData{TweetData{298273498291123, "it's kind of the best that Mitt Romney saying shit like this goes public on #s17."},
		TweetData{298273498291124, "my thoughts on new Ariel Pink, re: high pitchfork rating + \"Symphony of the Nymph\"; was thinking of this the whole time http://i.qkme.me/3qyfwm.jpg "},
		TweetData{298273498291125, "sending good thoughts/mojo to #s17! fight the good fight, and #freemollycrabapple!"},
		TweetData{298273498291126, ".@beatonna 's tweets / about being hunks / have given my heart / a change of mood / and saved some part / of a day I had rued."},
		TweetData{298273498291127, "@SrPablo I'm also fine with sneaking on to your account when you're asleep. #devious"},
		TweetData{298273498291129, "@SrPablo also, JELLY, I WANT TO PLAY DOTA2."}}

	dh.InsertFreshTweets("SrPablo", pabloTweets)
	dh.InsertFreshTweets("laurelita", laurenTweets)

	pabloTweetBacks := dh.GetTweetsFromStorage("SrPablo")
	laurenTweetBacks := dh.GetTweetsFromStorage("laurelita")

	ensureTweetsExist(pabloTweets, pabloTweetBacks, c)
	ensureTweetsExist(laurenTweets, laurenTweetBacks, c)
}

func ensureTweetsExist(expected, results []TweetData, c *gocheck.C) {
	// double for-loop is as slow as all the fucks I'm not giving.
	for _, expectTweet := range expected {
		present := false
		for _, resultTweet := range results {
			if expectTweet.Id == resultTweet.Id && expectTweet.Text == resultTweet.Text {
				present = true
				break
			}
		}
		c.Assert(present, gocheck.Equals, true)
	}
}
