package main

/*
Functions to retrieve tweets from the Twitter servers using their GET API, or
to post tweets.

TODO: Dynamic application secrets
*/
import (
	"ebooker/logging"
	"ebooker/oauth1"

	"encoding/json"
	"html"
	"io/ioutil"
	"net/http"
	"strconv"
)

type TweetFetcher struct {
	logger *logging.LogMaster
	oauth  *oauth1.OAuth1
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

const USER_TIMELINE_URL = "http://api.twitter.com/1/statuses/user_timeline.json"
const UPDATE_STATUS_URL = "https://api.twitter.com/1.1/statuses/update.json"

func getTweetFetcher(logger *logging.LogMaster, oauth *oauth1.OAuth1) TweetFetcher {
	return TweetFetcher{logger, oauth}
}

// DeepDive is for new accounts, of if you're the kind of person who runs
// 'make clean.' We take as much from the user's public-facing Twitter API
// as we can by recursively calling with the max_id. See:
//
// https://dev.twitter.com/docs/working-with-timelines
func (tf TweetFetcher) DeepDive(username string, accessToken *oauth1.Token) Tweets {
	tf.logger.StatusWrite("Doing a deep dive!\n")

	url := USER_TIMELINE_URL
	method := "GET"
	urlParams := map[string]string{
		"screen_name": username,
		"count":       "50",
		"include_rts": "false"}
	bodyParams := map[string]string{}
	authParams := map[string]string{}

	req := tf.oauth.CreateAuthorizedRequest(url, method, urlParams, bodyParams, authParams, accessToken)
	resp := tf.oauth.ExecuteRequest(req)

	tweets := tf.getTweetsFromResponse(resp)

	// in case of error
	if len(tweets) == 0 {
		return tweets
	}

	// the "- 1" is because max_id is inclusive, and we already have the tweet
	// represented by this ID.
	maxId := tweets[tweets.Len()-1].Id - 1
	for {
		urlParams["max_id"] = strconv.FormatUint(maxId, 10)

		req := tf.oauth.CreateAuthorizedRequest(url, method, urlParams, bodyParams, authParams, accessToken)
		resp := tf.oauth.ExecuteRequest(req)

		olderTweets := tf.getTweetsFromResponse(resp)
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
func (tf TweetFetcher) GetRecentTimeline(username string, latest *TweetData, accessToken *oauth1.Token) Tweets {
	url := USER_TIMELINE_URL
	method := "GET"
	urlParams := map[string]string{
		"screen_name": username,
		"count":       "50",
		"include_rts": "false",
		"since_id":    strconv.FormatUint(latest.Id, 10)}
	bodyParams := map[string]string{}
	authParams := map[string]string{}

	req := tf.oauth.CreateAuthorizedRequest(url, method, urlParams, bodyParams, authParams, accessToken)
	resp := tf.oauth.ExecuteRequest(req)

	tweets := tf.getTweetsFromResponse(resp)

	return tweets
}

// Calls the Twitter API's "update" function on the account name provided, with
// the status text assigned. We assume the user has already provided the app
// access to their credentials with OAuth; in case they haven't, we ask for them
// and otherwise drop the request from this scope.
func (tf TweetFetcher) sendTweet(status string, accessToken *oauth1.Token) {
	tf.logger.DebugWrite("Sending Tweet POST request!\n")
	url := UPDATE_STATUS_URL
	method := "POST"
	urlParams := map[string]string{}
	bodyParams := map[string]string{"status": status}
	authParams := map[string]string{}
	req := tf.oauth.CreateAuthorizedRequest(url, method, urlParams, bodyParams, authParams, accessToken)
	tf.oauth.ExecuteRequest(req)
}

func (tf TweetFetcher) getTweetsFromResponse(resp *http.Response) Tweets {

	if resp.StatusCode == http.StatusOK {
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

		for i := range tweets {
			tweets[i].Text = html.UnescapeString(tweets[i].Text)
		}

		return tweets
	}
	return Tweets{}
}

func appendSlices(slice1, slice2 Tweets) Tweets {
	for _, tweet := range slice2 {
		slice1 = append(slice1, tweet)
	}
	return slice1
}
