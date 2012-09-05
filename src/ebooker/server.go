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

//    "appengine"
//    "appengine/datastore"
//    "appengine/user"
)

var welcomeTempl = template.Must(template.New("splash").Parse(templateStr))
var textGen = CreateGenerator(1, 140)
var tweetFetch = CreateTweetFetcher("laurelita")

func init() {
    http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
    //c := appengine.NewContext(r)

    display := populateData(r)
    welcomeTempl.Execute(w, display)
}


func populateData(r *http.Request) PageDisplay {
    // Get user data from datastore - accounts, wherefrom, etc.

    // Run a tweetfetcher update
    tweets := tweetFetch.GetTimelineFromRequest(r)
    //   Compare with datastore, insert new values.

    // fetch tweets from datastore, store them in a []TweetData

    // Generate some faux tweets
    count := 30
    for i := 0; i < len(tweets); i++ {
        textGen.AddSeeds(tweets[i].Text)
    }
    var nonsense []string
    for i := 0; i < count; i++ {
        nonsense = append(nonsense, textGen.GenerateText())
    }
    // Populate the TweetDisplay
    //pageDisplay := PageDisplay{}
    pageDisplay := getDummyData()
    pageDisplay.Accounts[0].TweetDisplay = tweets
    pageDisplay.Accounts[0].NonsenseTweets = nonsense
    return pageDisplay
}

func getDummyData() PageDisplay {

    tweets := []TweetData{
        TweetData {1929310283120, "#reasonstoloveSF: Castro theater Alfred Hitchcock festival #vertigo ❤"},
        TweetData {2010203912831, "good job brain, adding an outfit last worn in December to my chai to the slight blustery weather. #prettywrongtho"} }

    nonsense := []string{ "false, false" , "falser!" }
    laurelita_ebooks := Account{ "laurelita_ebooks", "laurelita", time.Now(), len(tweets), tweets, nonsense }
    var accounts []Account
    accounts = append(accounts, laurelita_ebooks)

    pageDisplay := PageDisplay{ accounts , "pablo.a.meier" }
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
</body>
</html>
`

// buttons to do the following things:
//   * update twitter datastore - uses URL fetching, datastore
//   * generate a bunch of faux tweets
//   
//   * authenticate - I can manage laurelita_ebooks, SrPablo_ebooks, etc. from
//     my pablo.a.meier@gmail.com account
