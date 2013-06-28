/*
Command-line client that uses the janky ebooker package to generate tweets
easily. Requires a running service to connect to.
*/
package main

import (
	"ebooker/defs"
	"ebooker/logging"
	"ebooker/oauth1"

	"flag"
	"fmt"
	"log"
	"net/rpc"
	"strings"
)

func main() {

	var port, userlist, sched, token, botName, keyFile string
	var numTweets, prefixLen int
	var reps, generate, newBot, cancel, del, list bool
	flag.StringVar(&port, "port", "8998", "Port to server location.")
	flag.StringVar(&userlist, "users", "SrPablo,__MICHAELJ0RDAN", "Comma-seperated list of users to read from (no spaces)")
	flag.IntVar(&numTweets, "numTweets", 15, "Number of tweets to generate.")
	flag.IntVar(&prefixLen, "prefixLen", 2, "Length of generation prefix. Smaller = more random, Larger = more accurate.")
	flag.BoolVar(&reps, "representations", false, "Treat all forms of a text (.e.g \"It's/ITS/its'\") as equivalent.")

	flag.BoolVar(&generate, "generate", true, "Generate tweets and print them to stdout. Overrides \"newbot\".")
	flag.BoolVar(&newBot, "newBot", false, "Creates a new bot to run on the server. Must set \"generate\" to false.")
	flag.StringVar(&botName, "botName", "SrPablo_ebooks", "The name for your new bot.")
	flag.StringVar(&sched, "sched", "0 11,19 * * *", "cron-formatted string for how often the new bot will tweet. NOT IMPLEMENTED.")
	flag.StringVar(&token, "token", "", "Comma-separated pair of token & token secret. If not provided, we require you to complete a Twitter PIN-based authentication")
	flag.StringVar(&keyFile, "keyfile", "keys.txt", "File containing the application keys assigned to you by Twitter.")

	flag.BoolVar(&cancel, "cancelBot", false, "Must be used with botName -- sets the named bot to no longer tweet.")
	flag.BoolVar(&del, "deleteBot", false, "Must be used with botName -- removes the bot entirely from the server.")
	flag.BoolVar(&list, "listBots", false, "Prints a list of all the bots on this server")
	flag.Parse()

	client, err := rpc.DialHTTP("tcp", "127.0.0.1:"+port)
	defer client.Close()
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var authArgs defs.AuthParams
	if token == "" {
		lm := logging.GetLogMaster(false, true, false)
		applicationKey, applicationSecret := oauth1.ParseFromFile(keyFile)
		oauth := oauth1.CreateOAuth1(&lm, applicationKey, applicationSecret)
		requestToken := oauth.ObtainRequestToken()
		tokenObj := oauth.ObtainAccessToken(requestToken)
		lm.StatusWrite("Your access token is %v\n", tokenObj)
		authArgs = defs.AuthParams{botName, tokenObj.OAuthToken, tokenObj.OAuthTokenSecret}
	} else {
		components := strings.Split(token, ",")
		authArgs = defs.AuthParams{botName, components[0], components[1]}
	}

	genArgs := defs.GenParams{strings.Split(userlist, ","), numTweets, reps, prefixLen, authArgs}

	if generate {
		resp := make([]string, numTweets)
		err = client.Call("Ebooker.GenerateTweets", &genArgs, &resp)
		if err != nil {
			log.Fatal("generate tweets error:", err)
		}

		for i := range resp {
			fmt.Printf("%v\n", resp[i])
		}
	} else if newBot && !generate {
		var resp string
		schedArgs := defs.Schedule{sched}

		args := defs.NewBotParams{genArgs, authArgs, schedArgs}
		err = client.Call("Ebooker.NewBot", &args, &resp)
		if err != nil {
			log.Fatal("new bot error:", err)
		}
		fmt.Println(resp)
	} else if !generate && list {
		var toPrint []string
		err := client.Call("Ebooker.ListBots", "", &toPrint)
		if err != nil {
			log.Fatal("listBots error:", err)
		}
		fmt.Println(toPrint)
	} else if !generate && cancel {
		var msg string
		err := client.Call("Ebooker.CancelBot", botName, &msg)
		if err != nil {
			log.Fatal("cancelBot error:", err)
		}
		fmt.Println(msg)
	} else if !generate && del {
		var msg string
		err := client.Call("Ebooker.DeleteBot", botName, &msg)
		if err != nil {
			log.Fatal("deleteBot error:", err)
		}
		fmt.Println(msg)
	}
}

func newBot(genParams *defs.GenParams, client *rpc.Client) {
	sched := defs.Schedule{"30 12,18 * * *"}
	auth := defs.AuthParams{"SrPablo_ebooks", "", ""}

	args := defs.NewBotParams{*genParams, auth, sched}
	var resp string
	client.Call("Ebooker.NewBot", &args, &resp)

	fmt.Println(resp)
}
