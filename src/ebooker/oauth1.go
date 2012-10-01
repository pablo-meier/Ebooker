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
- write extensive tests!
- get a rough tweet up
- get and fetch secrets dynamically
- RUN IT
- Maintain access tokens, refresh tokens? Want to add PIN once, not worry about it.
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
	"strconv"
	"time"
)

type OAuth1 struct {
    logger *LogMaster

    applicationKey string
    applicationSecret string
}

type Token struct {
    oauthToken string
    oauthTokenSecret string
}

// Sends a Tweet
func (o OAuth1) sendTweetWithOAuth(status string) {
    url := "https://api.twitter.com/1.1/statuses/update.json"
//	url := "http://127.0.0.1:8888"

    requestToken := o.getRequestToken()
    accessToken := o.getAccessToken(requestToken)

    urlParams := map[string]string{}
    bodyParams := map[string]string{ "status" : status }
    authParams := map[string]string{}
    req := o.createAuthorizedRequest(url, urlParams, bodyParams, authParams, accessToken)
    o.makePostRequest(req)
}

// Gives us an access token to begin an OAuth exhange with Twitter.
func (o OAuth1) getRequestToken() *Token {
    url := "https://api.twitter.com/oauth/request_token"
    urlParams := map[string]string{}
    bodyParams := map[string]string{}
    authParams := map[string]string { "oauth_callback" : "oob" }
    req := o.createAuthorizedRequest(url, urlParams, bodyParams, authParams, nil)
    resp := o.makePostRequest(req)

    return o.parseTokenResponse(resp)
}


func (o OAuth1) getAccessToken(requestToken *Token) *Token {
    userFacingUrl := fmt.Sprintf("https://api.twitter.com/oauth/authenticate?oauth_token=%s", requestToken.oauthToken)

    fmt.Printf("Please sign in to Twitter at the following URL: %s\n", userFacingUrl)
    fmt.Printf("After successful Sign-in, Twitter will provide you a PIN number.\n")
    fmt.Printf("Please enter it here, without spaces: ")

    var pin uint
    fmt.Scanln(pin)

    url := "https://dev.twitter.com/docs/api/1/post/oauth/access_token"
    urlParams := map[string]string{}
    bodyParams := map[string]string{ "oauth_verifier" : string(pin) }
    authParams := map[string]string {}
    req := o.createAuthorizedRequest(url, urlParams, bodyParams, authParams, requestToken)
    resp := o.makePostRequest(req)

    return o.parseTokenResponse(resp)
}


// Handles most of the functionality in a way that's (reasonably) easy to call.
// Makes a POST request to the URL provided, handling the various places you can
// can put parameters (in the URL, e.g. twitter.com/authorize?token_id=900981,
// in the body e.g. status="Sup%20Son", or in the "Authorization:" part of the 
// Header).
//
// Understandable why we have it, and God bless crypto, but what a bloody mess.
func (o OAuth1) createAuthorizedRequest(url string, urlParams, bodyParams, authParams map[string]string, token *Token) *http.Request {

    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    o.logger.StatusWrite("Timestamp is %v\n", timestamp)

	authParamMap := map[string]string{
		"oauth_consumer_key":     o.applicationKey,
		"oauth_nonce" :           createNonce(),
		"oauth_timestamp":        timestamp,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_version":          "1.0"}

    for k,v := range authParams {
        authParamMap[k] = v
    }
    return o.authorizedRequestWithParams(url, urlParams, bodyParams, authParamMap, token)
}

// We seperate this function from the one above for testing.
func (o OAuth1) authorizedRequestWithParams(url string, urlParams, bodyParams, authParams map[string]string, token *Token) *http.Request {

	urlParamString := makeParamStringFromMap(urlParams)
	bodyString := makeParamStringFromMap(bodyParams)

	urlWithParams := strings.Join([]string{ url, urlParamString }, "?")
	req, err := http.NewRequest("POST", urlWithParams, strings.NewReader(bodyString))
	if err != nil {
		o.logger.StatusWrite("Error creating request object for POST\n")
		o.logger.DebugWrite("POST request to url: %v. Error: %v\n", url, err)
		o.logger.DebugWrite("Request is:\n")
		req.Write(os.Stdout)
	}

    if token != nil {
        authParams["oauth_token"] = token.oauthToken
    }


	signature := o.createSignature(urlParams, bodyParams, authParams, url, token)
	authParams["oauth_signature"] = signature
	o.finishHeader(req, authParams)

	return req
}

func makeParamStringFromMap(mp map[string]string) string {
	var total []string
	for k,v := range mp {
        param := strings.Join([]string{ percentEncode(k), percentEncode(v) }, "=")
        total = append(total, param)
	}
	return strings.Join(total, "&")
}


// Parses strings passed back from POST request_token, they look like
// (all concatenated):
//
// oauth_token=NPcudxy0yU5T3tBzho7iCotZ3cnetKwcTIRlX0iwRl0&
// oauth_token_secret=veNRnAWe6inFuo8o2u8SLLZLjolYDmDP7SzL0YfYI&
// oauth_callback_confirmed=true
func (o OAuth1) parseTokenResponse(resp *http.Response) *Token {
    tokenData, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        o.logger.StatusWrite("Error reading from response body %v\n", err)
    }
    return o.parseTokenData(string(tokenData))
}

func (o OAuth1) parseTokenData(tokenData string) *Token {
    params := strings.Split(string(tokenData), "&")
    paramMap := map[string]string{}
    for _, param := range params {
        kvPair := strings.Split(param, "=")
        paramMap[kvPair[0]] = kvPair[1]
    }

    if paramMap["oauth_callback_confirmed"] != "true" {
        o.logger.StatusWrite("oauth_callback_confirmed not true for response:\n")
        //resp.Write(os.Stdout)
    }

    return &Token{ paramMap["oauth_token"], paramMap["oauth_token_secret"] }
}

// The "nonce" is a relatively random alphanumeric string that we generate, and 
// the Twitter server uses to ensure that we're not sending the same request 
// twice. Their example is 42 characters long, so we'll just emulate that.
func createNonce() string {
	nonceLen := 42 // taken from their example
	src := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ09123456789")

	rslt := make([]byte, nonceLen)
	for i := 0; i < nonceLen; i++ {
		rslt[i] = src[rand.Intn(len(src))]
	}

	return string(rslt)
}

// Create the request signature. This is done according to the instructions on
// the API pages linked above.
func (o *OAuth1) createSignature(urlParams, bodyParams, authParams map[string]string, url string, token *Token) string {

    allMaps := []map[string]string { urlParams, bodyParams, authParams }
    allParams := map[string]string {}
    for _, mp := range allMaps {
        for k,v := range mp {
            allParams[k] = v
        }
    }

	signatureBaseString := o.makeSignatureBaseString(allParams, url)
	signingKey := o.makeSigningKey(token)

	hmacSha1 := hmac.New(sha1.New, []byte(signingKey))
	io.WriteString(hmacSha1, signatureBaseString)

	return base64.StdEncoding.EncodeToString(hmacSha1.Sum(nil))
}

func (o *OAuth1) makeSignatureBaseString(allParams map[string]string, url string) string {
	httpMethod := "POST"

	// Percent encode the parameters
	encoded := map[string]string{}
	for k, v := range allParams {
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
	encodedUrl := percentEncode(url)

	return strings.Join([]string{httpMethod, encodedUrl, parameterString}, "&")
}

func (o *OAuth1) makeSigningKey(token *Token) string {
    appKey := percentEncode(o.applicationSecret)
    consumerKey := ""
    if token != nil {
        consumerKey = percentEncode(token.oauthTokenSecret)
    }
	return strings.Join([]string{appKey, consumerKey}, "&")
}

func (o *OAuth1) finishHeader(req *http.Request, authParams map[string]string) {
	var paramstrings []string
	for k, v := range authParams {
		paramstrings = append(paramstrings, fmt.Sprintf("%s=\"%s\"", percentEncode(k), percentEncode(v)))
	}
	authString := strings.Join(paramstrings, ", ")
	authorizationString := strings.Join([]string{"OAuth", authString}, " ")
	req.Header.Add("Authorization", authorizationString)
}

func (o OAuth1) makePostRequest(req *http.Request) *http.Response {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		o.logger.StatusWrite("Error executing POST request\n")
        req.Write(os.Stdout)
	}
	if resp.StatusCode != 200 {
        o.logger.StatusWrite("Twitter returned non-200 status: %v\n", resp.Status)
        o.logger.DebugWrite("POST Request: \n")
        req.Write(os.Stdout)
        o.logger.DebugWrite("Response:\n")
        resp.Write(os.Stdout)
	}

	return resp

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
