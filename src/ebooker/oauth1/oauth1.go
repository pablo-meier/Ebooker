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
*/

package oauth1

import (
	"ebooker/logging"

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
	"strconv"
	"strings"
	"time"
)

// OAuth key names
const (
	OAUTH_CALLBACK         = "oauth_callback"
	OAUTH_CONSUMER_KEY     = "oauth_consumer_key"
	OAUTH_NONCE            = "oauth_nonce"
	OAUTH_SIGNATURE_METHOD = "oauth_signature_method"
	OAUTH_TIMESTAMP        = "oauth_timestamp"
	OAUTH_VERIFIER         = "oauth_verifier"
	OAUTH_VERSION          = "oauth_version"

	// Twitter URLS
	AUTHENTICATE_URL  = "https://api.twitter.com/oauth/authenticate"
	ACCESS_TOKEN_URL  = "https://api.twitter.com/oauth/access_token"
	REQUEST_TOKEN_URL = "https://api.twitter.com/oauth/request_token"

	// Other 'naked' values
	OUT_OF_BAND = "oob"
)

type OAuth1 struct {
	logger *logging.LogMaster

	applicationKey    string
	applicationSecret string
}

type Token struct {
	OAuthToken       string
	OAuthTokenSecret string
}

func CreateOAuth1(l *logging.LogMaster, key, secret string) OAuth1 {
	return OAuth1{l, key, secret}
}

// Return a Token based on the application's credentials.
func (o OAuth1) MakeToken() *Token {
	return &Token{o.applicationKey, o.applicationSecret}
}

// Gives us a request token to begin an OAuth exchange with Twitter.
func (o OAuth1) ObtainRequestToken() *Token {
	o.logger.DebugWrite("Making a POST request for a request token...\n")
	url := REQUEST_TOKEN_URL
	method := "POST"
	urlParams := map[string]string{}
	bodyParams := map[string]string{}
	authParams := map[string]string{OAUTH_CALLBACK: OUT_OF_BAND}
	req := o.CreateAuthorizedRequest(url, method, urlParams, bodyParams, authParams, nil)
	resp := o.ExecuteRequest(req)

	return o.parseTokenResponse(resp)
}

// Given a request token, we get an Access token for the user account.
func (o OAuth1) ObtainAccessToken(requestToken *Token) *Token {

	tokenUrl := fmt.Sprintf("oauth_token=%s", requestToken.OAuthToken)
	userFacingUrl := strings.Join([]string{AUTHENTICATE_URL, tokenUrl}, "?")

	fmt.Printf("Please sign in to Twitter at the following URL: %s\n", userFacingUrl)
	fmt.Printf("After successful Sign-in, Twitter will provide you a PIN number.\n")
	fmt.Printf("Please enter it here, without spaces: ")

	var pin int
	fmt.Scanf("%d", &pin)

	o.logger.DebugWrite("Making a POST request for an Access token...\n")
	url := ACCESS_TOKEN_URL
	method := "POST"
	urlParams := map[string]string{}
	bodyParams := map[string]string{OAUTH_VERIFIER: strconv.Itoa(pin)}
	authParams := map[string]string{}
	req := o.CreateAuthorizedRequest(url, method, urlParams, bodyParams, authParams, requestToken)
	resp := o.ExecuteRequest(req)

	return o.parseTokenResponse(resp)
}

// Handles most of the functionality in a way that's (reasonably) easy to call.
// Makes a POST request to the URL provided, handling the various places you can
// can put parameters (in the URL, e.g. twitter.com/authorize?token_id=900981,
// in the body e.g. status="Sup%20Son", or in the "Authorization:" part of the
// Header).
//
// Understandable why we have it, and God bless crypto, but what a bloody mess.
func (o OAuth1) CreateAuthorizedRequest(url, method string,
	urlParams, bodyParams, authParams map[string]string,
	token *Token) *http.Request {

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	authParamMap := map[string]string{
		OAUTH_NONCE:            createNonce(),
		OAUTH_CONSUMER_KEY:     o.applicationKey,
		OAUTH_TIMESTAMP:        timestamp,
		OAUTH_SIGNATURE_METHOD: "HMAC-SHA1",
		OAUTH_VERSION:          "1.0"}

	for k, v := range authParams {
		authParamMap[k] = v
	}

	return o.authorizedRequestWithParams(url, method, urlParams, bodyParams, authParamMap, token)
}

// We seperate this function from the one above for testing.
func (o OAuth1) authorizedRequestWithParams(urlRaw, method string,
	urlParams, bodyParams, authParams map[string]string,
	token *Token) *http.Request {

	urlParamString := makeParamStringFromMap(urlParams)
	bodyString := makeParamStringFromMap(bodyParams)

	urlWithParams := strings.Join([]string{urlRaw, urlParamString}, "?")

	req, err := http.NewRequest(method, urlWithParams, strings.NewReader(bodyString))

	if err != nil {
		o.logger.StatusWrite("Error creating request object for %v\n", method)
		o.logger.DebugWrite("%v request to url: %v. Error: %v\n", method, urlRaw, err)
		o.logger.DebugWrite("Request is:\n")
		req.Write(os.Stdout)
	}

	if token != nil {
		authParams["oauth_token"] = token.OAuthToken
	}

	signature := o.createSignature(urlParams, bodyParams, authParams, urlRaw, method, token)
	authParams["oauth_signature"] = signature
	o.finishHeader(req, authParams)

	return req
}

func makeParamStringFromMap(mp map[string]string) string {
	var total []string
	for k, v := range mp {
		param := strings.Join([]string{percentEncode(k), percentEncode(v)}, "=")
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
	tokenBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		o.logger.StatusWrite("Error reading from response body %v\n", err)
	}
	tokenData := string(tokenBytes)
	return o.parseTokenData(tokenData)
}

func (o OAuth1) parseTokenData(tokenData string) *Token {
	params := strings.Split(string(tokenData), "&")
	paramMap := map[string]string{}
	for _, param := range params {
		kvPair := strings.Split(param, "=")
		paramMap[kvPair[0]] = kvPair[1]
	}

	if confirmed, exists := paramMap["oauth_callback_confirmed"]; exists {
		if confirmed != "true" {
			o.logger.StatusWrite("oauth_callback_confirmed not true for response.\n")
		}
	}

	return &Token{paramMap["oauth_token"], paramMap["oauth_token_secret"]}
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
func (o *OAuth1) createSignature(urlParams, bodyParams, authParams map[string]string, url, method string, token *Token) string {

	signatureBaseString := o.makeSignatureBaseString(urlParams, bodyParams, authParams, url, method)
	signingKey := o.makeSigningKey(token)

	hmacSha1 := hmac.New(sha1.New, []byte(signingKey))
	io.WriteString(hmacSha1, signatureBaseString)

	return base64.StdEncoding.EncodeToString(hmacSha1.Sum(nil))
}

func (o *OAuth1) makeSignatureBaseString(urlParams, bodyParams, authParams map[string]string, url, method string) string {

	allMaps := []map[string]string{urlParams, bodyParams, authParams}
	allParams := map[string]string{}
	for _, mp := range allMaps {
		for k, v := range mp {
			allParams[k] = v
		}
	}

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

	return strings.Join([]string{method, encodedUrl, parameterString}, "&")
}

func (o *OAuth1) makeSigningKey(token *Token) string {
	appKey := percentEncode(o.applicationSecret)
	consumerKey := ""
	if token != nil {
		consumerKey = percentEncode(token.OAuthTokenSecret)
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

	// For Twitter 1.1 API, update will fail unless this is in Header.
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "*/*")
}

func (o OAuth1) ExecuteRequest(req *http.Request) *http.Response {
	client := &http.Client{}
//	req.Write(os.Stdout)
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		o.logger.StatusWrite("Error executing POST request: %v\n", err)
		req.Write(os.Stdout)
	} else if resp.StatusCode != http.StatusOK {
		o.logger.StatusWrite("Twitter returned non-200 status: %v\n", resp.Status)
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
