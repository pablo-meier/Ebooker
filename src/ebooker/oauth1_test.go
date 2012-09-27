package ebooker

import (
	"launchpad.net/gocheck"
)

// hook up gocheck into the gotest runner.
type OAuthSuite struct{}

var _ = gocheck.Suite(&OAuthSuite{})

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
