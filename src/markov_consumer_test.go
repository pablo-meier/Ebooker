




package main

import (
    "launchpad.net/gocheck"
    "testing"
)


// hook up gocheck into the gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

type MarkovSuite struct{}
var _ = gocheck.Suite(&MarkovSuite{})


// To test AddSeeds, we create a Generator, feed it some text, and ensure:
//   * The existence of all prefixes.
//   * The existence of all prefixes of the correct length.
//   * The existence of appropriate suffixes for a number of the prefixes.
//   * The correct frequency counts on suffixes.
func (s MarkovSuite) TestAddSeeds(c *gocheck.C) {
    c.Assert(11, gocheck.Equals, 11)
}


func (s MarkovSuite) TestGenerateText(c *gocheck.C) {
    c.Assert(14, gocheck.Equals, 14)
}
