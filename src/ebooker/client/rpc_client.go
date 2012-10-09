/*
Command-line client that uses the janky ebooker package to generate tweets
easily. Requires a running service to connect to.
*/
package main

import (
	"ebooker/defs"

	"flag"
	"fmt"
	"log"
	"net/rpc"
	"strings"
)

func main() {

	var port, userlist string
	var numTweets, prefixLen int
	var reps bool
	flag.StringVar(&port, "port", "8998", "Port to run the server on.")
	flag.StringVar(&userlist, "users", "SrPablo,__MICHAELJ0RDAN", "Comma-seperated list of users to read from (no spaces)")
	flag.IntVar(&numTweets, "numTweets", 15, "Number of tweets to generate.")
	flag.IntVar(&prefixLen, "prefixLen", 2, "Length of generation prefix. Smaller = more random, Larger = more accurate.")
	flag.BoolVar(&reps, "representations", false, "Treat all forms of a text (.e.g \"It's/ITS/its'\") as equivalent.")
	flag.Parse()

	client, err := rpc.DialHTTP("tcp", "127.0.0.1:"+port)
	defer client.Close()
	if err != nil {
		log.Fatal("dialing:", err)
	}

	args := defs.GenParams{strings.Split(userlist, ","), numTweets, reps, prefixLen}
	resp := make([]string, numTweets)

	err = client.Call("Ebooker.GenerateTweets", &args, &resp)
	if err != nil {
		log.Fatal("generate tweets error:", err)
	}

	for i := range resp {
		fmt.Printf("%v\n", resp[i])
	}
}

func newBot(genParams *defs.GenParams, client *rpc.Client) {
	auth := defs.AuthParams{"SrPablo_ebooks", "", ""}
	sched := defs.Schedule{"30 12,18 * * *"}

	args := defs.NewBotParams{*genParams, auth, sched}
	var resp string
	client.Call("Ebooker.NewBot", &args, &resp)

	fmt.Println(resp)
}
