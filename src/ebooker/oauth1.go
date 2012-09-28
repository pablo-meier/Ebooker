/*
So Twitter is cool and doesn't support OAuth 2, which we have a great Go library
for:

http://goauth2.googlecode.com/

While there are a few Githubbed OAuth 1 libraries available, I'd rather check
this out for myself. Most of it is taken from the descriptions on these
Twitter Developer pages:

https://dev.twitter.com/docs/auth/authorizing-request      - Authorizing a request
https://dev.twitter.com/docs/auth/creating-signature       - Creating a signature
https://dev.twitter.com/docs/auth/pin-based-authorization  - PIN based auth

TODO:
- Make signature creation cleaner. Update uses same functions as tokens.
- get a rough tweet up
- get and fetch secrets, both consumer's and app's
- refactor with mocks for testing?
- RUN IT
*/

package ebooker

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const applicationKey = "MxIkjx9eCC3j1JC8kTig"
const applicationSecret = "nopenopenope"

type OAuth1 struct {
    logger *LogMaster
}

type Token struct {
    oauthToken string
    oauthTokenSecret string
}

type AuthorizedRequest struct {
	parameterStringMap map[string]string
	url                string
	Request            *http.Request

	consumerSecret string
    tokenSecret string
}

func (o OAuth1) sendTweetWithOAuth(status string) {
    url := "https://api.twitter.com/1.1/statuses/update.json"
//	krl := "http://127.0.0.1:8888"

//    params := map[string]string{}
    params := ""
    body := ""
    o.createAuthorizedRequest(url, params, body)
}


func (o OAuth1) createAuthorizedRequest(url, status, tokensecret string) *AuthorizedRequest {

	paramMap := map[string]string{
		"oauth_consumer_key":     applicationKey,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        string(time.Now().Unix()),
		"oauth_version":          "1.0",
		"status":                 status}

	req, err := http.NewRequest("POST", url, strings.NewReader(percentEncode(status)))
	if err != nil {
		o.logger.StatusWrite("Error creating request object for POST\n")
		o.logger.DebugWrite("POST request to url: %v. Error: %v\n", url, err)
	}

    requestToken := o.getRequestToken()
    accessToken := o.getAccessToken(requestToken)

    paramMap["oauth_token"] = accessToken.oauthToken
    tokenSecret := accessToken.oauthTokenSecret

	authdata := AuthorizedRequest{paramMap, url, req, applicationSecret, tokenSecret}
	authdata.setNonce()
	authdata.createSignature()
	authdata.finishHeader()

	return &authdata
}


func (o OAuth1) getAccessToken(token *Token) *Token {
    userFacingUrl := fmt.Sprintf("https://api.twitter.com/oauth/authenticate?oauth_token=%s", token.oauthToken)

    fmt.Printf("Please sign in to Twitter at the following URL: %s\n", userFacingUrl)
    fmt.Printf("After successful Sign-in, Twitter will provide you a PIN number.\n")
    fmt.Printf("Please enter it here, without spaces: ")

    var pin uint
    fmt.Scanln(pin)

    url := "https://dev.twitter.com/docs/api/1/post/oauth/access_token"
    tokenStr := fmt.Sprintf("oauth_token=%s", token.oauthToken)
    verifierStr := fmt.Sprintf("oauth_verifier=%d", pin)
    params := map[string]string { "Authorization" : tokenStr }

    accessToken := o.makePostRequest(url, params, verifierStr)
    return accessToken
}

// Gives us an access token to begin an OAuth exhange with Twitter.
func (o OAuth1) getRequestToken() *Token {
    url := "https://api.twitter.com/oauth/request_token"
    params := map[string]string { "Authorization" : "oauth_callback=\"oob\"" }
    requestToken := o.makePostRequest(url, params, "")
    return requestToken
}

func (o OAuth1) makePostRequest(url string, params map[string]string, body string) *Token {
    req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		o.logger.StatusWrite("Error creating request object for POST\n")
		o.logger.DebugWrite("POST request to url: %v. Error: %v\n", url, err)
	}

	for k,v := range params {
	    req.Header.Add(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		o.logger.StatusWrite("Error executing POST request\n")
		o.logger.DebugWrite("POST request to url: %v. Error: %v\n", url, err)
	}
	if resp.StatusCode != 200 {
        o.logger.StatusWrite("Twitter returned non-200 status: %v\n", resp.Status)
        o.logger.DebugWrite("POST Request: \n")
        req.Write(os.Stdout)
        o.logger.DebugWrite("Response:\n")
        resp.Write(os.Stdout)
	}

    tokenData, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        o.logger.StatusWrite("Error reading from response body %v\n", err)
    }

    return parseRequestTokenParams(string(tokenData))
}


// Parses strings passed back from POST request_token, they look like
// (all concatenated):
//
// oauth_token=NPcudxy0yU5T3tBzho7iCotZ3cnetKwcTIRlX0iwRl0&
// oauth_token_secret=veNRnAWe6inFuo8o2u8SLLZLjolYDmDP7SzL0YfYI&
// oauth_callback_confirmed=true
func parseRequestTokenParams(s string) *Token {
    params := strings.Split(s, "&")
    paramMap := map[string]string{}
    for _, param := range params {
        kvPair := strings.Split(param, "=")
        paramMap[kvPair[0]] = kvPair[1]
    }

    if paramMap["oauth_callback_confirmed"] != "true" {
        fmt.Println()
    }

    return &Token{ paramMap["oauth_token"], paramMap["oauth_token_secret"] }
}

// The "nonce" is a relatively random alphanumeric string that we generate, and 
// the Twitter server uses to ensure that we're not sending the same request 
// twice. Their example is 42 characters long, so we'll just emulate that.
func (o *AuthorizedRequest) setNonce() {
	nonceLen := 42 // taken from their example
	src := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ09123456789")

	rslt := make([]byte, nonceLen)
	for i := 0; i < nonceLen; i++ {
		rslt[i] = src[rand.Intn(len(src))]
	}

	o.parameterStringMap["oauth_nonce"] = string(rslt)
}

// Create the request signature. This is done according to the instructions on
// the API pages linked above.
func (o *AuthorizedRequest) createSignature() {

	signatureBaseString := o.makeSignatureBaseString()
	signingKey := o.makeSigningKey()

	hmacSha1 := hmac.New(sha1.New, []byte(signingKey))
	io.WriteString(hmacSha1, signatureBaseString)

	o.parameterStringMap["oauth_signature"] = base64.StdEncoding.EncodeToString(hmacSha1.Sum(nil))
}

func (o *AuthorizedRequest) makeSigningKey() string {
    appKey := percentEncode(o.consumerSecret)
    consumerKey := percentEncode(o.tokenSecret)
	return strings.Join([]string{appKey, consumerKey}, "&")
}

func (o *AuthorizedRequest) makeSignatureBaseString() string {
	httpMethod := "POST"

	// Percent encode the parameters
	encoded := map[string]string{}
	for k, v := range o.parameterStringMap {
		encoded[percentEncode(k)] = percentEncode(v)
	}

	// sort the parameters
	var keys []string
	for k, _ := range encoded {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// create parameter string 
	for i, key := range keys {
		keys[i] = strings.Join([]string{key, encoded[key]}, "=")

	}
	parameterString := percentEncode(strings.Join(keys, "&"))
	encodedUrl := percentEncode(o.url)

	return strings.Join([]string{httpMethod, encodedUrl, parameterString}, "&")
}

func (o *AuthorizedRequest) finishHeader() {
	paramMap := o.parameterStringMap
	authSection := map[string]string{}
	relevantKeys := []string{
		"oauth_consumer_key",
		"oauth_nonce",
		"oauth_signature",
		"oauth_signature_method",
		"oauth_timestamp",
		"oauth_token",
		"oauth_version"}

	for _, key := range relevantKeys {
		authSection[key] = paramMap[key]
	}

	var paramstrings []string
	for k, v := range authSection {
		paramstrings = append(paramstrings, fmt.Sprintf("%s=\"%s\"", percentEncode(k), percentEncode(v)))
	}
	authString := strings.Join(paramstrings, ", ")
	authorizationString := strings.Join([]string{"OAuth", authString}, " ")
	o.Request.Header.Add("Authorization", authorizationString)
}

// wow... am I actually implementing this? Instructions from here:
//
// https://dev.twitter.com/docs/auth/percent-encoding-parameters
//
// Debated putting this in stringutils, but decided against it because
// this is purely for OAuth.
func percentEncode(str string) string {
	asBytes := []byte(str)

	var returnBuf []byte
	for _, curr := range asBytes {
		if isLowercaseAscii(curr) || isUppercaseAscii(curr) || isDigit(curr) || isReserved(curr) {
			returnBuf = append(returnBuf, curr)
		} else {
			dst := make([]byte, 2, 2)
			hex.Encode(dst, []byte{curr})
			dst = bytes.ToUpper(dst)

			returnBuf = append(returnBuf, 0x25) // appending '%'
			returnBuf = append(returnBuf, dst[0])
			returnBuf = append(returnBuf, dst[1])
		}
	}

	return string(returnBuf)
}

func isLowercaseAscii(b byte) bool { return b >= 97 && b <= 122 }
func isUppercaseAscii(b byte) bool { return b >= 65 && b <= 90 }
func isDigit(b byte) bool          { return b >= 48 && b <= 57 }
func isReserved(b byte) bool       { return b == 45 || b == 46 || b == 95 || b == 126 }
