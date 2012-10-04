/*
Command-line client that uses the janky ebooker package to generate tweets
easily. Requires a running service to connect to.
*/
package main

import (
	"fmt"
	//"flag"
	"log"
	"net/rpc"
)

type GenerateTweetsArgs struct {
	Users     []string
	NumTweets int
	Reps      bool
	PrefixLen int
}

func main() {
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	numTweets := 15
	args := GenerateTweetsArgs{[]string{"SrPablo"}, numTweets, false, 2}
	resp := make([]string, numTweets)

	err = client.Call("EbookerRequest.GenerateTweets", &args, &resp)
	if err != nil {
		log.Fatal("generate tweets error:", err)
	}

	for _, str := range resp {
		fmt.Println(str)
	}
}
