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
    "log"

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

    // fetch Generator from datastore
    gen := fetchGenerator(c)

    // get datastore tweets
    oldTweets := getDatastoreTweets(c)

    var tweets []TweetData
    if len(oldTweets) == 0 {
        tweets = DeepDive(c, "laurelita")
    } else {
        freshId := oldTweets[0].Id
        tweets = GetTimelineFromRequest(c, "laurelita", freshId)
    }

    // Insert new values to datastore and Generator
    insertFreshTweets(tweets, gen, c)

    // Generate some faux tweets
    var nonsense []string
    count := 50
    for i := 0; i < count; i++ {
        nonsense = append(nonsense, gen.GenerateText())
    }
    // Populate the TweetDisplay
    //pageDisplay := PageDisplay{}
    pageDisplay := getDummyData()
    pageDisplay.Accounts[0].TweetDisplay = tweets
    pageDisplay.Accounts[0].NonsenseTweets = nonsense
    pageDisplay.DebugText = debugText
    return pageDisplay
}


// fetchGenerator finds the appropriate Generator to retrieve given the context.
func fetchGenerator(c appengine.Context) *Generator {
    key := datastore.NewKey(c, "Generator", "laurelita_ebooks", 0, nil)

    var gen Generator
    err := datastore.Get(c, key, &gen)
    if err == datastore.ErrNoSuchEntity {
        gen = *(CreateGenerator("laurelita_ebooks", 1, 140))
    }
    return &gen
}

// insertFreshTweets takes the array of tweets just received, finds all that
// aren't present in the data store, and adds them to the Generator and datastore.
// We then update the generator in the datastore.
func insertFreshTweets(fresh []TweetData, gen *Generator, c appengine.Context) {
    tweetkey := datastore.NewKey(c, "TweetData", "laurelita", 0, nil)
    genkey := datastore.NewKey(c, "Generator", "laurelita_ebook", 0, nil)

    if len(fresh) > 0 {
        for i := range fresh {
            gen.AddSeeds(fresh[i].Text)
            datastore.Put(c, tweetkey, &fresh[i])
        }
        datastore.Put(c, genkey, &gen)
    }
}

func getDatastoreTweets(c appengine.Context) []TweetData {
    twitterUser := "laurelita"
    q := datastore.NewQuery("TweetData").Filter("Screen_name =", twitterUser).Order("-Id")

    var oldTweets []TweetData
    for t := q.Run(c); ; {
        var tData TweetData
        _, err := t.Next(&tData)
        if err == datastore.Done {
            break
        }
        if err != nil {
            log.Fatal("error!", err)
            return []TweetData{}
        }
        oldTweets = append(oldTweets, tData)
    }
    return oldTweets
}


// getFreshTweets returns the set difference between the fetched tweets from the 
// datastore and the tweets extracted from the Twitter API. TODO make this not 
// awful, just doing the obvious, midnight answer.
func getFreshTweets(oldTweets, newTweets []TweetData) []TweetData {
    seen := make(map[uint64]bool)
    for i := range oldTweets {
        seen[oldTweets[i].Id] = true
    }

    difference := []TweetData{}
    for i := range newTweets {
        _, dup := seen[newTweets[i].Id]
        if !dup {
            difference = append(difference, newTweets[i])
        }
    }
    return difference
}

func getDummyData() PageDisplay {

    tweets := []TweetData{
        TweetData {1929310283120, "#reasonstoloveSF: Castro theater Alfred Hitchcock festival #vertigo ❤", "laurelita" },
        TweetData {2010203912831, "good job brain, adding an outfit last worn in December to my chai to the slight blustery weather. #prettywrongtho", "laurelita" } }

    nonsense := []string{ "false, false" , "falser!" }
    laurelita_ebooks := Account{ "laurelita_ebook", "laurelita", time.Now(), len(tweets), tweets, nonsense }
    var accounts []Account
    accounts = append(accounts, laurelita_ebooks)

    pageDisplay := PageDisplay{ accounts , "pablo.a.meier", "" }
    return pageDisplay
}

type Account struct {
    Name string
    BasedOff string
    LastUpdate time.Time
    TotalTweets int
    TweetDisplay []TweetData
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

  <h3>Fetched Tweets</h3>
  <p>Total tweets: {{.TotalTweets}}</p>
  <table>
    {{range .TweetDisplay}}
    <tr>
      <td>{{.Id}}</td><td>{{.Text}}</td>
    </tr>
    {{end}}
  </table>

  <h3>Accepted Tweets</h3>
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
