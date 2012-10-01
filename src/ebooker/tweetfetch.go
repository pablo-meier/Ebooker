/*
Package retrieves tweets from the Twitter servers, using their GET API.
*/
package ebooker

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"strings"
)

const applicationKey = "MxIkjx9eCC3j1JC8kTig"
const applicationSecret = "nopenopenope"



type TweetFetcher struct {
	logger *LogMaster
}

type TweetData struct {
	Id   uint64 `json:"id"`
	Text string `json:"text"`
}

type Tweets []TweetData

// For sorting
func (t Tweets) Len() int           { return len(t) }
func (t Tweets) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t Tweets) Less(i, j int) bool { return t[i].Id < t[j].Id }

const urlRequestBase = "http://api.twitter.com/1/statuses/user_timeline.json"
const screenNameParam = "screen_name=%s"
const countParam = "count=50"
const includeRtsParam = "include_rts=false"
const sinceIdParam = "since_id=%d"
const maxIdParam = "max_id=%d"

var params = strings.Join([]string{screenNameParam, countParam, includeRtsParam}, "&")
var baseQuery = strings.Join([]string{urlRequestBase, params}, "?")

func GetTweetFetcher(logger *LogMaster) TweetFetcher {
	return TweetFetcher{logger}
}

// DeepDive is for new accounts, of if you're the kind of person who runs
// 'make clean.' We take as much from the user's public-facing Twitter API
// as we can by recursively calling with the max_id. See:
//
// https://dev.twitter.com/docs/working-with-timelines
func (tf TweetFetcher) DeepDive(username string) Tweets {
	tf.logger.StatusWrite("Doing a deep dive!\n")

	queryStr := fmt.Sprintf(baseQuery, username)
	tweets := tf.getTweetsFromQuery(queryStr)

	// the "- 1" is because max_id is inclusive, and we already have the tweet
	// represented by this ID.
	maxId := tweets[tweets.Len()-1].Id - 1
	for {
		newQueryBase := strings.Join([]string{queryStr, maxIdParam}, "&")
		newQueryStr := fmt.Sprintf(newQueryBase, maxId)

		olderTweets := tf.getTweetsFromQuery(newQueryStr)
		if olderTweets.Len() == 0 {
			break
		}

		newOldestId := olderTweets[olderTweets.Len()-1].Id

		if maxId == newOldestId {
			break
		} else {
			maxId = newOldestId
			tweets = appendSlices(tweets, olderTweets)
			tf.logger.StatusWrite("Tweets have grown to %d\n", tweets.Len())
		}
	}

	return tweets
}

// GetRecentTimeline is the much more common use case: we fetch tweets from the
// timeline, using since_id. This allows us to incrementally build our tweet
// database.
func (tf TweetFetcher) GetTimelineFromRequest(username string, latest *TweetData) Tweets {
	queryBase := strings.Join([]string{baseQuery, sinceIdParam}, "&")
	queryStr := fmt.Sprintf(queryBase, username, latest.Id)

	tweets := tf.getTweetsFromQuery(queryStr)

	return tweets
}

func (tf TweetFetcher) getTweetsFromQuery(queryStr string) Tweets {

	tf.logger.DebugWrite("Calling Twitter with GET String \"%s\"\n", queryStr)
	resp, err := http.Get(queryStr)

	if err != nil {
		tf.logger.StatusWrite("Received unexpected error from Twitter GET call.\n")
		tf.logger.DebugWrite("error is: %v\n", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		tf.logger.StatusWrite("Received unexpected error from reading HTTP Response.\n")
		tf.logger.DebugWrite("error is: %v\n", err)
	}

	var tweets Tweets

	err = json.Unmarshal(body, &tweets)
	if err != nil {
		tf.logger.StatusWrite("Received unexpected error from Unmarshalling JSON Response.\n")
		tf.logger.DebugWrite("error is: %v\n", err)
	}

	for _, tweet := range tweets {
		tweet.Text = html.UnescapeString(tweet.Text)
	}

	return tweets
}

// Calls the Twitter API's "update" function on the account name provided, with 
// the status text assigned. We assume the user has already provided the app
// access to their credentials with OAuth; in case they haven't, we ask for them
// and otherwise drop the request from this scope.
func SendTweet(status string) {
    logger := GetLogMaster(false, true, false)
	o := OAuth1 { &logger , applicationKey, applicationSecret }

	o.sendTweetWithOAuth(status)
}

func appendSlices(slice1, slice2 Tweets) Tweets {
	for _, tweet := range slice2 {
		slice1 = append(slice1, tweet)
	}
	return slice1
}
