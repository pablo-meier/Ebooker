package main

import (
	"launchpad.net/gocheck"
)

// hook up gocheck into the gotest runner.
type TweetFetchSuite struct{}

var _ = gocheck.Suite(&TweetFetchSuite{})

func (t TweetFetchSuite) TestGetUserTimeline(c *gocheck.C) {
	c.Assert(2, gocheck.Equals, 2)
}
