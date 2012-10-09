/*
Shared definitions that facilitate RPC with the server and any client.
*/
package defs

type Tweets []string

// Parameters needed to Generate Tweets.
type GenParams struct {
	Users     []string // The Twitter users whose Timeline will form our corpus.
	NumTweets int      // The number of tweets to generate.
	Reps      bool     // Whether all variations of text (e.g. "ITS/it's/It's") are treated as equivalent
	PrefixLen int      // Length of generation prefix. Smaller = more random, Larger = more accurate.
}

// Parameters needed to get a new bot up and running.
type NewBotParams struct {
	Gen   GenParams  // Gen parameters so we know what styles of tweets to generate.
	Auth  AuthParams // Auth parameters so we have tweeting privileges.
	Sched Schedule   // How often the bot should tweet.
}

// Parameters needed to Authenticate.
type AuthParams struct {
	User        string // The Username who is represented by this token.
	Token       string // Their publicly-known token value
	TokenSecret string // Their privately-held token secret
}

// Parameters needed to Schedule Tweeting.
type Schedule struct {
	Cron string // schedule as a set of cron-formatted strings
}
