

package ebooker

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "strings"
)


type TweetData struct {
    // Storing the Id in an int64 rather than uint64 because datastore only 
    // allows signed ints :(
    Id int64 `json:"id"`
    Text string `json:"text"`
    Screen_name string `json:"screen_name"`
}


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
func DeepDive(username string) []TweetData {
    fmt.Println("Doing a deep dive!")
    queryStr := fmt.Sprintf(baseQuery, username)

    tweets := getTweetsFromClient(queryStr)

    // HAX HAX HAX
    for i := range tweets {
        tweets[i].Screen_name = username
    }

    maxId := tweets[len(tweets) - 1].Id
    for ;; {
        newQueryBase := strings.Join([]string{ queryStr, maxIdParam }, "&")
        newQueryStr := fmt.Sprintf(newQueryBase, maxId)

        olderTweets := getTweetsFromClient(newQueryStr)
        if len(olderTweets) == 0 { break }

        newOldestId := olderTweets[len(olderTweets) - 1].Id

        if maxId == newOldestId {
            break
        } else {
            maxId = newOldestId
            tweets = appendSlices(tweets, olderTweets)
            fmt.Println("Tweets have grown to", len(tweets))
        }
    }

    return tweets
}


// GetRecentTimeline is the much more common use case: we fetch tweets from the
// timeline, using since_id. This allows us to incrementally build our tweet
// database.
func GetTimelineFromRequest(username string, latest *TweetData) []TweetData {
    queryBase := strings.Join([]string{ baseQuery, sinceIdParam }, "&")
    queryStr := fmt.Sprintf(queryBase, latest.Screen_name, latest.Id)

    tweets := getTweetsFromClient(queryStr)

    return tweets
}


func getTweetsFromClient(queryStr string) []TweetData {

    resp, err := http.Get(queryStr)

    if err != nil { log.Fatal(err) }

    body, err := ioutil.ReadAll(resp.Body)
    defer resp.Body.Close()
    if err != nil { log.Fatal(err) }

    var tweets []TweetData;

    err = json.Unmarshal(body, &tweets)
    if err != nil { log.Fatal(err) }

    return tweets
}


func appendSlices(slice1, slice2 []TweetData) []TweetData {
   newslice := make([]TweetData, len(slice1) + len(slice2))
   copy(newslice, slice1)
   copy(newslice[len(slice1):], slice2)
   return newslice
}


