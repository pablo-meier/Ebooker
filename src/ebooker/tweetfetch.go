

package ebooker

import (
    "encoding/json"
    "fmt"
    "html"
    "log"
    "net/http"
    "io/ioutil"
    "strings"
)


type TweetData struct {
    Id uint64 `json:"id"`
    Text string `json:"text"`
}

type Tweets []TweetData

// For sorting
func (t Tweets) Len() int { return len(t) }
func (t Tweets) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t Tweets) Less(i, j int) bool { return t[i].Id < t[j].Id }


const urlRequestBase = "http://api.twitter.com/1/statuses/user_timeline.json"
const screenNameParam = "screen_name=%s"
const countParam = "count=50"
const includeRtsParam = "include_rts=false"
const sinceIdParam = "since_id=%d"
const maxIdParam = "max_id=%d"

var params = strings.Join([]string{ screenNameParam, countParam, includeRtsParam }, "&")
var baseQuery = strings.Join([]string{ urlRequestBase, params}, "?")

// DeepDive is for new accounts, of if you're the kind of person who runs
// 'make clean.' We take as much from the user's public-facing Twitter API
// as we can by recursively calling with the max_id. See:
//
// https://dev.twitter.com/docs/working-with-timelines
func DeepDive(username string) Tweets {
    fmt.Println("Doing a deep dive!")
    queryStr := fmt.Sprintf(baseQuery, username)

    tweets := getTweetsFromClient(queryStr)

    // the "- 1" is because max_id is inclusive, and we already have it.
    maxId := tweets[tweets.Len() - 1].Id - 1
    for ;; {
        newQueryBase := strings.Join([]string{ queryStr, maxIdParam }, "&")
        newQueryStr := fmt.Sprintf(newQueryBase, maxId)

        olderTweets := getTweetsFromClient(newQueryStr)
        if olderTweets.Len() == 0 { break }

        newOldestId := olderTweets[olderTweets.Len() - 1].Id

        if maxId == newOldestId {
            break
        } else {
            maxId = newOldestId
            tweets = appendSlices(tweets, olderTweets)
            fmt.Println("Tweets have grown to", tweets.Len())
        }
    }

    return tweets
}


// GetRecentTimeline is the much more common use case: we fetch tweets from the
// timeline, using since_id. This allows us to incrementally build our tweet
// database.
func GetTimelineFromRequest(username string, latest *TweetData) Tweets {
    queryBase := strings.Join([]string{ baseQuery, sinceIdParam }, "&")
    queryStr := fmt.Sprintf(queryBase, username, latest.Id)

    tweets := getTweetsFromClient(queryStr)

    return tweets
}


func getTweetsFromClient(queryStr string) Tweets {

    resp, err := http.Get(queryStr)

    if err != nil { log.Fatal(err) }

    body, err := ioutil.ReadAll(resp.Body)
    defer resp.Body.Close()
    if err != nil { log.Fatal(err) }

    var tweets Tweets

    err = json.Unmarshal(body, &tweets)
    if err != nil { log.Fatal(err) }

    for _, tweet := range tweets {
        tweet.Text = html.UnescapeString(tweet.Text)
    }

    return tweets
}


func appendSlices(slice1, slice2 Tweets) Tweets {
   newslice := make(Tweets, slice1.Len() + slice2.Len())
   copy(newslice, slice1)
   copy(newslice[slice1.Len():], slice2)
   return newslice
}


