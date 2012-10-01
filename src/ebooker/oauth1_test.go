package ebooker

import (
	"launchpad.net/gocheck"

	"regexp"
)

type OAuthSuite struct{}

var _ = gocheck.Suite(&OAuthSuite{})






// This tests against the example Twitter themselves walk you through, once 
// you've obtained an access token.
//
// https://dev.twitter.com/docs/auth/creating-signature
func (oa OAuthSuite) TestTwitterSignatureExample(c *gocheck.C) {
    logger := GetLogMaster(false, true, false)
    dh := GetDataHandle("./ebooker_tweets.db", &logger)
    o := OAuth1{ &logger , &dh, "xvz1evFS4wEEPTGEFPHBog", "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw" }

    url := "https://api.twitter.com/1/statuses/update.json"
    urlParams := map[string]string{ "include_entities" : "true" }
    bodyParams := map[string]string{ "status" : "Hello Ladies + Gentlemen, a signed OAuth request!" }
    authParams := map[string]string {
        "oauth_consumer_key":     "xvz1evFS4wEEPTGEFPHBog",
		"oauth_nonce":            "kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        "1318622958",
		"oauth_token":            "370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb",
		"oauth_version":          "1.0" }

	token := Token{ "370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb" , "LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE" }

    req := o.authorizedRequestWithParams(url, urlParams, bodyParams, authParams, &token)

    authstring := req.Header.Get("Authorization")
    regex, _ := regexp.Compile("oauth_signature=\"([^\"]+)\"")
    signature := regex.FindStringSubmatch(authstring)[1]
    c.Assert(signature, gocheck.Equals, percentEncode("tnnArxj06cWHq44gCs1OSKk/jLY="))
}

// This example comes from 
//
// https://dev.twitter.com/docs/auth/implementing-sign-twitter
//
// when requesting a request token, before you get an "oauth_token" value.
func (oa OAuthSuite) TestSecondTwitterExample(c *gocheck.C) {
    logger := GetLogMaster(false, true, false)
    dh := GetDataHandle("./ebooker_tweets.db", &logger)
    o := OAuth1{ &logger , &dh, "cChZNFj6T5R0TigYB9yd1w", "L8qq9PZyRg6ieKGEKhZolGC0vJWLw8iEJ88DRdyOg" }

    url := "https://api.twitter.com/oauth/request_token"
    urlParams := map[string]string{}
    bodyParams := map[string]string{}
    authParams := map[string]string {
        "oauth_callback" :        "http://localhost/sign-in-with-twitter/",
        "oauth_consumer_key":     "cChZNFj6T5R0TigYB9yd1w",
		"oauth_nonce":            "ea9ec8429b68d6b77cd5600adbbb0456",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        "1318467427",
		"oauth_version":          "1.0" }

    req := o.authorizedRequestWithParams(url, urlParams, bodyParams, authParams, nil)

    authstring := req.Header.Get("Authorization")
    regex, _ := regexp.Compile("oauth_signature=\"([^\"]+)\"")
    signature := regex.FindStringSubmatch(authstring)[1]
    c.Assert(signature, gocheck.Equals, "F1Li3tvehgcraF8DMJ7OyxO4w9Y%3D")
}


func (oa OAuthSuite) TestMakingSigningKey(c *gocheck.C) {
    logger := GetLogMaster(false, true, false)
    dh := GetDataHandle("./ebooker_tweets.db", &logger)
    o := OAuth1{ &logger , &dh, "xvz1evFS4wEEPTGEFPHBog", "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw" }

	token := Token{ "370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb" , "LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE" }


    // With a valid token
    c.Assert(o.makeSigningKey(&token), gocheck.Equals, "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw&LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE")

    // With a nil token
    c.Assert(o.makeSigningKey(nil), gocheck.Equals, "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw&")
}


func (oa OAuthSuite) TestTokenStringParsing(c *gocheck.C) {
    logger := GetLogMaster(false, true, false)
    dh := GetDataHandle("./ebooker_tweets.db", &logger)
    o := OAuth1{ &logger , &dh, "xvz1evFS4wEEPTGEFPHBog", "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw" }

    testcase := "oauth_token=NPcudxy0yU5T3tBzho7iCotZ3cnetKwcTIRlX0iwRl0&oauth_token_secret=veNRnAWe6inFuo8o2u8SLLZLjolYDmDP7SzL0YfYI&oauth_callback_confirmed=true"
    token := o.parseTokenData(testcase)

    c.Assert(token.oauthToken, gocheck.Equals, "NPcudxy0yU5T3tBzho7iCotZ3cnetKwcTIRlX0iwRl0")
    c.Assert(token.oauthTokenSecret, gocheck.Equals, "veNRnAWe6inFuo8o2u8SLLZLjolYDmDP7SzL0YfYI")
}


func (o OAuthSuite) TestPercentEncode(c *gocheck.C) {

	testCases := map[string]string{
		"hello":                          "hello",
		"CAPS":                           "CAPS",
		"11WithDigits9":                  "11WithDigits9",
		"a space and exclamation point!": "a%20space%20and%20exclamation%20point%21",
		"Dogs, Cats & Mice":              "Dogs%2C%20Cats%20%26%20Mice",
		"Reserved Chars -._~":            "Reserved%20Chars%20-._~",
		"Ladies + Gentlemen":             "Ladies%20%2B%20Gentlemen"}

	for k, v := range testCases {
		c.Assert(percentEncode(k), gocheck.Equals, v)
	}
}


