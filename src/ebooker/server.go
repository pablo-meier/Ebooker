/* The server is responsible for tasks unrelated to text generation, namely:
 * fetching tweets for source, storing the corpus in a DB, keeping our generated
 * text in the DB, doing all this on a schedule. I imagine this will eventually
 * become a Goroutine that will be called from a 'real' server (e.g. GET with
 * parameters user=SrPablo&freq=daily will spawn one of these that creates
 * tweets off of @SrPablo, and create a new tweet every day.
 */

package ebooker

import (
    "net/http"
    "html/template"
    "time"
    "fmt"

    "appengine"
    "appengine/datastore"
)

var welcomeTempl = template.Must(template.New("splash").Parse(templateStr))

func init() {
    http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
    display := populateData(r)
    welcomeTempl.Execute(w, display)
}


func populateData(r *http.Request) PageDisplay {
    // Get user data from datastore - accounts, wherefrom, etc.
    c := appengine.NewContext(r)
    debugText := ""
    twitterUser := "SrPablo"

    // fetch Generator from datastore
    gen := fetchGenerator(c)

    // get datastore tweets
    oldTweets := getDatastoreTweets(c, twitterUser)
    fmt.Printf("datastore size is %d\n", len(oldTweets))

    var tweets []TweetData
    if len(oldTweets) == 0 {
        tweets = DeepDive(c, twitterUser)
    } else {
        tweets = GetTimelineFromRequest(c, twitterUser, oldTweets[0])
    }

    // Insert new values to datastore and Generator
    insertFreshTweets(tweets, gen, c)

    for i := range oldTweets {
        gen.AddSeeds(oldTweets[i].Text)
    }

    // Generate some faux tweets
    var nonsense []string
    count := 50
    for i := 0; i < count; i++ {
        nonsense = append(nonsense, gen.GenerateText())
    }
    // Populate the TweetDisplay
    //pageDisplay := PageDisplay{}
    pageDisplay := getDummyData()
    pageDisplay.Accounts[0].NonsenseTweets = nonsense
    pageDisplay.DebugText = debugText
    return pageDisplay
}


// fetchGenerator finds the appropriate Generator to retrieve given the context.
// Unfortunately, GAE can't store maps in the datastore, so we're pretty fucked
// for persistent storage. In the meantime, we're hacking it with just recreating
// it each time, since it's a fast operation.
func fetchGenerator(c appengine.Context) *Generator {
    gen := CreateGenerator("laurelita_ebooks", 1, 140)
    return gen
}


// getDatastoreTweets retrieves the 10 most recent tweets "on record" for a 
// twitter user.
func getDatastoreTweets(c appengine.Context, twitterUser string) []*TweetData {
    q := datastore.NewQuery("TweetData").Filter("Screen_name =", twitterUser).Order("-Id")

    var oldTweets []*TweetData
    if _, err := q.GetAll(c, &oldTweets); err != nil {
        fmt.Printf("Getall had non-nil error! %v\n", err)
    }

    return oldTweets
}


// insertFreshTweets takes the array of tweets just received, finds all that
// aren't present in the data store, and adds them to the Generator and 
// datastore. We then update the generator in the datastore.
func insertFreshTweets(fresh []TweetData, gen *Generator, c appengine.Context) {
    fmt.Printf("inserting %d tweets to data store!\n", len(fresh))

    for i := range fresh {
        tweetData := fresh[i]
        gen.AddSeeds(tweetData.Text)
        key := datastore.NewIncompleteKey(c, "TweetData", nil)
        _, err := datastore.Put(c, key, &tweetData)
        if err != nil {
            fmt.Println("error is: ", err)
        }
    }
}

func getDummyData() PageDisplay {

    nonsense := []string{ "false, false" , "falser!" }
    laurelita_ebooks := Account{ "laurelita_ebook", "laurelita", time.Now(), nonsense }
    var accounts []Account
    accounts = append(accounts, laurelita_ebooks)

    pageDisplay := PageDisplay{ accounts , "pablo.a.meier", "" }
    return pageDisplay
}

type Account struct {
    Name string
    BasedOff string
    LastUpdate time.Time
    NonsenseTweets []string
}

type PageDisplay struct {
    Accounts []Account
    Username string
    DebugText string
}

const templateStr = `
<!DOCTYPE html>
<html>
<head>
<title>Ebooker!</title>
</head>

<body>
<h1>Ebooker!</h1>
{{range .Accounts}}
<div class="bot-instance">
  <h2>{{.Name}}</h2>

  <h3>Generated Tweets</h3>
  <ul>
    {{range .NonsenseTweets}}
    <li>{{.}}</li>
    {{end}}
  </ul>
  
  <h3>Generate Tweets</h3>
  <ul>
    <li>Click on one to "accept!"</li>
  </ul>

  <h3>Data</h3>
  <p><strong>Based on:</strong> <a href="https://twitter.com/{{.BasedOff}}">@{{.BasedOff}}</a></p>
  <p><strong>Last updated:</strong> {{.LastUpdate}}</p>
</div>
{{end}}
<p>Hope it was fun, {{.Username}}</p>
<p>Debug:</p>
<p>{{.DebugText }}</p>
</body>
</html>
`

// buttons to do the following things:
//   * update twitter datastore - uses URL fetching, datastore
//   * generate a bunch of faux tweets
//   
//   * authenticate - I can manage laurelita_ebooks, SrPablo_ebooks, etc. from
//     my pablo.a.meier@gmail.com account
