package oauth1

import (
	"launchpad.net/gocheck"

	"ebooker/defs"

	"net/http"
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
	o := OAuth1{&logger, "xvz1evFS4wEEPTGEFPHBog", "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw"}

	url := "https://api.twitter.com/1/statuses/update.json"
	method := "POST"
	urlParams := map[string]string{"include_entities": "true"}
	bodyParams := map[string]string{"status": "Hello Ladies + Gentlemen, a signed OAuth request!"}
	authParams := map[string]string{
		"oauth_consumer_key":     "xvz1evFS4wEEPTGEFPHBog",
		"oauth_nonce":            "kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        "1318622958",
		"oauth_token":            "370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb",
		"oauth_version":          "1.0"}

	token := defs.Token{"370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb", "LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE"}

	// check signature base string
	baseStringExpected := "POST&https%3A%2F%2Fapi.twitter.com%2F1%2Fstatuses%2Fupdate.json&include_entities%3Dtrue%26oauth_consumer_key%3Dxvz1evFS4wEEPTGEFPHBog%26oauth_nonce%3DkYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg%26oauth_signature_method%3DHMAC-SHA1%26oauth_timestamp%3D1318622958%26oauth_token%3D370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb%26oauth_version%3D1.0%26status%3DHello%2520Ladies%2520%252B%2520Gentlemen%252C%2520a%2520signed%2520OAuth%2520request%2521"
	c.Assert(o.makeSignatureBaseString(urlParams, bodyParams, authParams, url, method), gocheck.Equals, baseStringExpected)

	// check signature
	req := o.authorizedRequestWithParams(url, method, urlParams, bodyParams, authParams, &token)
	authstring := req.Header.Get("Authorization")
	regex, _ := regexp.Compile("oauth_signature=\"([^\"]+)\"")
	signature := regex.FindStringSubmatch(authstring)[1]
	c.Assert(signature, gocheck.Equals, percentEncode("tnnArxj06cWHq44gCs1OSKk/jLY="))
}

// Making this example from the OAuth tool for update.json, since it's still getting hung up on it.
func (oa OAuthSuite) TestAuthOnUpdate(c *gocheck.C) {

	logger := GetLogMaster(false, true, false)
	o := OAuth1{&logger, "MxIkjx9eCC3j1JC8kTig", "IgOkwoh5m7AS4LplszxcPaF881vjvZYZNCAvvUz1x0"}

	url := "https://api.twitter.com/1.1/statuses/update.json"
	method := "POST"
	urlParams := map[string]string{}
	bodyParams := map[string]string{"status": "IMMATWEET"}
	authParams := map[string]string{
		"oauth_consumer_key":     "MxIkjx9eCC3j1JC8kTig",
		"oauth_nonce":            "1bd818f5d8e62ceb172aad5bae030fd3",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_token":            "27082544-JW0JZKi69R6OloylRBbs85By30kvZ7IfoGmGoiFvt",
		"oauth_timestamp":        "1349163796",
		"oauth_version":          "1.0"}

	token := defs.Token{"27082544-JW0JZKi69R6OloylRBbs85By30kvZ7IfoGmGoiFvt", "R2ieHCPMIECQnDhXMLOh3zL0w2CC484gFKVdBq6E"}

	baseStringExpected := "POST&https%3A%2F%2Fapi.twitter.com%2F1.1%2Fstatuses%2Fupdate.json&oauth_consumer_key%3DMxIkjx9eCC3j1JC8kTig%26oauth_nonce%3D1bd818f5d8e62ceb172aad5bae030fd3%26oauth_signature_method%3DHMAC-SHA1%26oauth_timestamp%3D1349163796%26oauth_token%3D27082544-JW0JZKi69R6OloylRBbs85By30kvZ7IfoGmGoiFvt%26oauth_version%3D1.0%26status%3DIMMATWEET"
	c.Assert(o.makeSignatureBaseString(urlParams, bodyParams, authParams, url, method), gocheck.Equals, baseStringExpected)

	// check signature
	req := o.authorizedRequestWithParams(url, method, urlParams, bodyParams, authParams, &token)
	authstring := req.Header.Get("Authorization")
	regex, _ := regexp.Compile("oauth_signature=\"([^\"]+)\"")
	signature := regex.FindStringSubmatch(authstring)[1]
	c.Assert(signature, gocheck.Equals, "Jk3epY305uVOD9dRFdqpioYBXHA%3D")
}

// This example comes from
//
// https://dev.twitter.com/docs/auth/implementing-sign-twitter
//
// when requesting a request token, before you get an "oauth_token" value.
func (oa OAuthSuite) TestSecondTwitterExample(c *gocheck.C) {
	logger := GetLogMaster(false, true, false)
	o := OAuth1{&logger, "cChZNFj6T5R0TigYB9yd1w", "L8qq9PZyRg6ieKGEKhZolGC0vJWLw8iEJ88DRdyOg"}

	url := "https://api.twitter.com/oauth/request_token"
	method := "POST"
	urlParams := map[string]string{}
	bodyParams := map[string]string{}
	authParams := map[string]string{
		"oauth_callback":         "http://localhost/sign-in-with-twitter/",
		"oauth_consumer_key":     "cChZNFj6T5R0TigYB9yd1w",
		"oauth_nonce":            "ea9ec8429b68d6b77cd5600adbbb0456",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        "1318467427",
		"oauth_version":          "1.0"}

	req := o.authorizedRequestWithParams(url, method, urlParams, bodyParams, authParams, nil)

	checkSignature(req, "F1Li3tvehgcraF8DMJ7OyxO4w9Y%3D", c)
}

func (oa OAuthSuite) TestStatusUpdateWithoutEncoding(c *gocheck.C) {
	logger := GetLogMaster(false, true, false)
	o := OAuth1{&logger, "MxIkjx9eCC3j1JC8kTig", "IgOkwoh5m7AS4LplszxcPaF881vjvZYZNCAvvUz1x0"}

	url := "https://api.twitter.com/1.1/statuses/update.json"
	method := "POST"
	urlParams := map[string]string{}
	bodyParams := map[string]string{"status": "IMMATWEET"}
	authParams := map[string]string{
		"oauth_consumer_key":     "MxIkjx9eCC3j1JC8kTig",
		"oauth_nonce":            "ef1efdb1c6b03c70ae2800543caae04d",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_token":            "27082544-JW0JZKi69R6OloylRBbs85By30kvZ7IfoGmGoiFvt",
		"oauth_timestamp":        "1349229371",
		"oauth_version":          "1.0"}

	token := defs.Token{"27082544-JW0JZKi69R6OloylRBbs85By30kvZ7IfoGmGoiFvt", "R2ieHCPMIECQnDhXMLOh3zL0w2CC484gFKVdBq6E"}

	baseStringExpected := "POST&https%3A%2F%2Fapi.twitter.com%2F1.1%2Fstatuses%2Fupdate.json&oauth_consumer_key%3DMxIkjx9eCC3j1JC8kTig%26oauth_nonce%3Def1efdb1c6b03c70ae2800543caae04d%26oauth_signature_method%3DHMAC-SHA1%26oauth_timestamp%3D1349229371%26oauth_token%3D27082544-JW0JZKi69R6OloylRBbs85By30kvZ7IfoGmGoiFvt%26oauth_version%3D1.0%26status%3DIMMATWEET"

	c.Assert(o.makeSignatureBaseString(urlParams, bodyParams, authParams, url, method), gocheck.Equals, baseStringExpected)

	req := o.authorizedRequestWithParams(url, method, urlParams, bodyParams, authParams, &token)
	checkSignature(req, "P6IKBc5LPV7Cz%2F%2FXnjpQuCisgek%3D", c)
}

func (oa OAuthSuite) TestGetTimelineWithAuthorization(c *gocheck.C) {
	logger := GetLogMaster(false, true, false)
	o := OAuth1{&logger, "MxIkjx9eCC3j1JC8kTig", "IgOkwoh5m7AS4LplszxcPaF881vjvZYZNCAvvUz1x0"}

	url := "https://api.twitter.com/1.1/statuses/user_timeline.json"
	method := "GET"
	urlParams := map[string]string{"screen_name": "theletterjeff"}
	bodyParams := map[string]string{}
	authParams := map[string]string{
		"oauth_consumer_key":     "MxIkjx9eCC3j1JC8kTig",
		"oauth_nonce":            "e0420a453875a19723a3873c9d6af3f0",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_token":            "27082544-QJA8iu2G4s7xG9OBRFKLlPntzJakmxidUgrlYtlIy",
		"oauth_timestamp":        "1349331540",
		"oauth_version":          "1.0"}

	token := defs.Token{"27082544-QJA8iu2G4s7xG9OBRFKLlPntzJakmxidUgrlYtlIy", "yuNeA8Z2DPLu8wwU7zYlsxIGIEyMqqxzaczCafdtvYY"}

	baseStringExpected := "GET&https%3A%2F%2Fapi.twitter.com%2F1.1%2Fstatuses%2Fuser_timeline.json&oauth_consumer_key%3DMxIkjx9eCC3j1JC8kTig%26oauth_nonce%3De0420a453875a19723a3873c9d6af3f0%26oauth_signature_method%3DHMAC-SHA1%26oauth_timestamp%3D1349331540%26oauth_token%3D27082544-QJA8iu2G4s7xG9OBRFKLlPntzJakmxidUgrlYtlIy%26oauth_version%3D1.0%26screen_name%3Dtheletterjeff"

	c.Assert(o.makeSignatureBaseString(urlParams, bodyParams, authParams, url, method), gocheck.Equals, baseStringExpected)

	req := o.authorizedRequestWithParams(url, method, urlParams, bodyParams, authParams, &token)
	checkSignature(req, "J8pKvAeHfse6dhv8Z06epOEOArQ%3D", c)
}

func checkSignature(req *http.Request, expected string, c *gocheck.C) {
	authstring := req.Header.Get("Authorization")
	regex, _ := regexp.Compile("oauth_signature=\"([^\"]+)\"")
	signature := regex.FindStringSubmatch(authstring)[1]
	c.Assert(signature, gocheck.Equals, expected)
}

func (oa OAuthSuite) TestMakingSigningKey(c *gocheck.C) {
	logger := GetLogMaster(false, true, false)
	o := OAuth1{&logger, "xvz1evFS4wEEPTGEFPHBog", "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw"}

	token := defs.Token{"370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb", "LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE"}

	// With a valid token
	c.Assert(o.makeSigningKey(&token), gocheck.Equals, "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw&LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE")

	// With a nil token
	c.Assert(o.makeSigningKey(nil), gocheck.Equals, "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw&")
}

func (oa OAuthSuite) TestTokenStringParsing(c *gocheck.C) {
	logger := GetLogMaster(false, true, false)
	o := OAuth1{&logger, "xvz1evFS4wEEPTGEFPHBog", "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw"}

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
