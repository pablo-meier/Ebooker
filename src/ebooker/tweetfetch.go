

package ebooker

import (
    "encoding/json"
    "net/http"
    "log"
    "io/ioutil"
    "fmt"
    "appengine"
    "appengine/urlfetch"
)


type TweetFetcher struct {
    user string
}

type TweetData struct {
    Id uint64
    Text string
    Screen_name string
}

func CreateTweetFetcher(user string) *TweetFetcher {
    return &TweetFetcher{user}
}

func GetTimelineFromRequest(c appengine.Context, username string) []TweetData {
    client := urlfetch.Client(c)
    count := 100

    fetchStr := fmt.Sprintf("http://api.twitter.com/1/statuses/user_timeline.json?screen_name=%s&count=%d", username, count)
    resp, err := client.Get(fetchStr)

    if err != nil { log.Fatal(err) }

    body, err := ioutil.ReadAll(resp.Body)
    defer resp.Body.Close()
    if err != nil { log.Fatal(err) }

    var tweets []TweetData;

    err = json.Unmarshal(body, &tweets)
    if err != nil { log.Fatal(err) }

    return tweets
}


// TODO: Generalize this with GAE version above...
func (t TweetFetcher) GetUserTimeline() []TweetData {
    count := 100
    fetchStr := fmt.Sprintf("http://api.twitter.com/1/statuses/user_timeline.json?screen_name=%s&count=%d", t.user, count)
    resp, err := http.Get(fetchStr)
    if err != nil { log.Fatal(err) }

    body, err := ioutil.ReadAll(resp.Body)
    defer resp.Body.Close()
    if err != nil { log.Fatal(err) }

    var tweets []TweetData;

    err = json.Unmarshal(body, &tweets)
    if err != nil { log.Fatal(err) }

//    fmt.Println("Response is\n", resp)
//    fmt.Println("body is\n", body)
//    fmt.Println("Tweets as structs are\n%+v", tweets)

    return tweets
}

