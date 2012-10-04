/*
Package retrieves tweets from the Twitter servers, using their GET API.

TODO: 
- Access tokens (default to our own?)
- Turn responses into Tweets
*/
package ebooker

import (
	"encoding/json"
	"html"
	"io/ioutil"
	"net/http"
	"strconv"
)

const applicationKey = "MxIkjx9eCC3j1JC8kTig"
const applicationSecret = "IgOkwoh5m7AS4LplszxcPaF881vjvZYZNCAvvUz1x0"

type TweetFetcher struct {
	logger *LogMaster
	oauth  *OAuth1
	data   *DataHandle
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
const DEFAULT_USER = "SrPablo"

func GetTweetFetcher(logger *LogMaster, dh *DataHandle) TweetFetcher {
	oauth := OAuth1{logger, applicationKey, applicationSecret}

	return TweetFetcher{logger, &oauth, dh}
}

// DeepDive is for new accounts, of if you're the kind of person who runs
// 'make clean.' We take as much from the user's public-facing Twitter API
// as we can by recursively calling with the max_id. See:
//
// https://dev.twitter.com/docs/working-with-timelines
func (tf TweetFetcher) DeepDive(username string) Tweets {
	return tf.AuthorizedDeepDive(username, DEFAULT_USER)
}

func (tf TweetFetcher) AuthorizedDeepDive(username, onBehalfOf string) Tweets {
	token := tf.getAccessToken(onBehalfOf)
	return tf.DeepDiveWithAccess(username, token)
}

func (tf TweetFetcher) DeepDiveWithAccess(username string, accessToken *Token) Tweets {
	tf.logger.StatusWrite("Doing a deep dive!\n")

	url := USER_TIMELINE_URL
	method := "GET"
	urlParams := map[string]string{
		"screen_name": username,
		"count":       "50",
		"include_rts": "false"}
	bodyParams := map[string]string{}
	authParams := map[string]string{}

	req := tf.oauth.createAuthorizedRequest(url, method, urlParams, bodyParams, authParams, accessToken)
	resp := tf.oauth.executeRequest(req)

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

		req := tf.oauth.createAuthorizedRequest(url, method, urlParams, bodyParams, authParams, accessToken)
		resp := tf.oauth.executeRequest(req)

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
func (tf TweetFetcher) GetRecentTimeline(username string, latest *TweetData) Tweets {
	return tf.AuthorizedGetRecentTimeline(username, DEFAULT_USER, latest)
}

func (tf TweetFetcher) AuthorizedGetRecentTimeline(username, onBehalfOf string, latest *TweetData) Tweets {
	token := tf.getAccessToken(onBehalfOf)
	return tf.GetRecentTimelineWithAccess(username, latest, token)
}

func (tf TweetFetcher) GetRecentTimelineWithAccess(username string, latest *TweetData, accessToken *Token) Tweets {
	url := USER_TIMELINE_URL
	method := "GET"
	urlParams := map[string]string{
		"screen_name": username,
		"count":       "50",
		"include_rts": "false",
		"since_id":    strconv.FormatUint(latest.Id, 10)}
	bodyParams := map[string]string{}
	authParams := map[string]string{}

	req := tf.oauth.createAuthorizedRequest(url, method, urlParams, bodyParams, authParams, accessToken)
	resp := tf.oauth.executeRequest(req)

	tweets := tf.getTweetsFromResponse(resp)

	return tweets
}

// Calls the Twitter API's "update" function on the account name provided, with 
// the status text assigned. We assume the user has already provided the app
// access to their credentials with OAuth; in case they haven't, we ask for them
// and otherwise drop the request from this scope.
func (tf TweetFetcher) SendTweet(username, status string) {
	accessToken := tf.getAccessToken(username)

	tf.logger.DebugWrite("Sending Tweet POST request!\n")
	url := UPDATE_STATUS_URL
	method := "POST"
	urlParams := map[string]string{}
	bodyParams := map[string]string{"status": status}
	authParams := map[string]string{}
	req := tf.oauth.createAuthorizedRequest(url, method, urlParams, bodyParams, authParams, accessToken)
	tf.oauth.executeRequest(req)
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

		for _, tweet := range tweets {
			tweet.Text = html.UnescapeString(tweet.Text)
		}

		return tweets
	}
	return Tweets{}
}

func (tf TweetFetcher) getAccessToken(user string) *Token {
	accessToken, exists := tf.data.getUserAccessToken(user)

	if !exists {
		tf.logger.StatusWrite("Access token for %v not present! Beginning OAuth...\n", user)
		requestToken := tf.oauth.getRequestToken()
		token := tf.oauth.getAccessToken(requestToken)

		tf.data.insertUserAccessToken(user, token)
		accessToken = token
	}
	return accessToken
}

func appendSlices(slice1, slice2 Tweets) Tweets {
	for _, tweet := range slice2 {
		slice1 = append(slice1, tweet)
	}
	return slice1
}
