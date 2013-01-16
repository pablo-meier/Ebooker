package oauth1

/*
 I used to hard code the application secrets, but sharing the code makes this a
 liability. Here we parse keys from a supplied file.
*/

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

const (
	KEY_PATTERN    = "^Consumer Key ?= ?(.+)"
	SECRET_PATTERN = "^Consumer Secret ?= ?(.+)"
)

func ParseFromFile(filename string) (string, string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Non-nil error opening file: ", err)
	}
	reader := bufio.NewReader(file)

	var consumerKey string
	var consumerSecret string

	isKey, err := regexp.Compile(KEY_PATTERN)
	if err != nil {
		fmt.Println("error with key pattern", err)
	}
	isSecret, err := regexp.Compile(SECRET_PATTERN)
	if err != nil {
		fmt.Println("error with key pattern", err)
	}

	for {
		str, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Non-nil error reading string: ", err)
		}

		keyMatches := isKey.FindStringSubmatch(str)
		secretMatches := isSecret.FindStringSubmatch(str)

		if keyMatches != nil {
			consumerKey = keyMatches[1]
		} else if secretMatches != nil {
			consumerSecret = secretMatches[1]
		}
	}

	if len(consumerKey) == 0 || len(consumerSecret) == 0 {
		fmt.Printf("File %v doesn't contain appropriate key format. Should contain\n"+
			"Consumer Key = <key>\nConsumer Secret = <secret>\n", filename)
	}

	return consumerKey, consumerSecret
}
