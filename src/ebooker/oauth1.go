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
- Consumer secret a big-ass nono. OBTAIN it from users. Store it?
- test lol
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
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"
)

const applicationKey = "MxIkjx9eCC3j1JC8kTig"

type OAuthRequest struct {
	parameterStringMap map[string]string
	url                string
	Request            *http.Request
}

func createOAuthRequest(url, status string) *OAuthRequest {

	paramMap := map[string]string{
		"oauth_consumer_key":     applicationKey,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        string(time.Now().Unix()),
		"oauth_token":            "370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb",
		"oauth_version":          "1.0",
		"status":                 status}

	req, err := http.NewRequest("POST", url, strings.NewReader(percentEncode(status)))
	if err != nil {
		fmt.Printf("error %v in making the request to %v\n", err, url)
	}

	authdata := OAuthRequest{paramMap, url, req}
	authdata.setNonce()
	authdata.createSignature()
	authdata.finishHeader()

	return &authdata
}

// The "nonce" is a relatively random alphanumeric string that we generate, and 
// the Twitter server uses to ensure that we're not sending the same request 
// twice. Their example is 42 characters long, so we'll just emulate that.
func (o *OAuthRequest) setNonce() {
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
func (o *OAuthRequest) createSignature() {

	signatureBaseString := o.makeSignatureBaseString()
	signingKey := o.makeSigningKey()

	hmacSha1 := hmac.New(sha1.New, []byte(signingKey))
	io.WriteString(hmacSha1, signatureBaseString)

	o.parameterStringMap["oauth_signature"] = base64.NewEncoding("").EncodeToString(hmacSha1.Sum(nil))
}

func (o *OAuthRequest) makeSigningKey() string {
	return strings.Join([]string{percentEncode(applicationKey), percentEncode("ConsumerSecret!")}, "&")
}

func (o *OAuthRequest) makeSignatureBaseString() string {
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

func (o *OAuthRequest) finishHeader() {
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
