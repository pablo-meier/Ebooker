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

package ebooker

import (
    "encoding/hex"
    "fmt"
    "net/http"
    "strings"
    "time"
    "math/rand"
)


const applicationKey = "MxIkjx9eCC3j1JC8kTig"

type OAuthRequest map[string]string

// Primary purpose of this file: given a username and a sample header,
// return a ha
func getOAuthCredentials(user string, header http.Header) http.Header {

    authdata := OAuthRequest {
              "oauth_consumer_key" : applicationKey,
              "oauth_signature_method" : "HMAC-SHA1",
              "oauth_timestamp" : string(time.Now().Unix()),
              "oauth_token" : "370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb",
              "oauth_version" : "1.0" }

    authdata.setNonce()

    var authcomponents []string
    for k, v := range authdata {
        strRep := fmt.Sprintf("%s=\"%s\"", k, v)
        authcomponents = append(authcomponents, strRep)
    }

    authstring := strings.Join(authcomponents, ", ")

    header.Add("Authorization:", fmt.Sprintf("OAuth %s", authstring))

    return header
}

// The "nonce" is a relatively random alphanumeric string that we generate, and 
// the Twitter server uses to ensure that we're not sending the same request 
// twice. Their example is 42 characters long, so we'll just emulate that.
func (o OAuthRequest) setNonce() {
    nonceLen := 42 // taken from their example
    src := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ09123456789")

    rslt := make([]byte, nonceLen)
    for i := 0; i < nonceLen; i++ {
        rslt[i] = src[rand.Intn(len(src))]
    }

    o["oauth_nonce"] = string(rslt)
}

// Create the request signature. This is done according to the instructions on
// the API pages linked above.
func (o OAuthRequest) createSignature() {

    o["oauth_signature"] = ""
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
        }
        var dst, src []byte
        src = append(src, curr)
        count, err := hex.Decode(dst, src)
        if err != nil {
            fmt.Printf("error decoding %v, returned %v bytes (err is %v)\n", src, count, err)
        }

        returnBuf = append(returnBuf, 0x25)   // appending '%'
        returnBuf = append(returnBuf, dst[0])
        returnBuf = append(returnBuf, dst[1])
    }

    return string(returnBuf)
}

func isLowercaseAscii(b byte) bool { return b >= 0x30 && b <= 0x39 }
func isUppercaseAscii(b byte) bool { return b >= 0x41 && b <= 0x5A }
func isDigit(b byte) bool { return b >= 0x30 && b <= 0x39 }
func isReserved(b byte) bool { return b == 0x2D || b == 0x2E || b == 0x5F || b == 0x7E }
