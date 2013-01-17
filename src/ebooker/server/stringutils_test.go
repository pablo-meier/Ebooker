package main

import (
	"launchpad.net/gocheck"
)

// hook up gocheck into the gotest runner.
type StringUtilsSuite struct{}

var _ = gocheck.Suite(&StringUtilsSuite{})

func (s StringUtilsSuite) TestStripReply(c *gocheck.C) {

	test1 := "@SrPablo ...really...? tharly...? really?"
	expected1 := "...really...? tharly...? really?"
	runStripReply(test1, expected1, c)

	test2 := "@SelfAwareRoomba nuuuuuuuuuuuuuuuuuuuuuu!!"
	expected2 := "nuuuuuuuuuuuuuuuuuuuuuu!!"
	runStripReply(test2, expected2, c)

	test3 := ".@AdamFelber so everything as normal, then?"
	expected3 := "so everything as normal, then?"
	runStripReply(test3, expected3, c)
}

func (s StringUtilsSuite) TestStripReplyMultiple(c *gocheck.C) {

	test1 := "@SrPablo @Aldaviva @love_that_goku ...really...? tharly...? really?"
	expected1 := "...really...? tharly...? really?"
	runStripReply(test1, expected1, c)

	test2 := "@SelfAwareRoomba @JoseCanseco nuuuuuuuuuuuuuuuuuuuuuu!!"
	expected2 := "nuuuuuuuuuuuuuuuuuuuuuu!!"
	runStripReply(test2, expected2, c)

	test3 := ".@AdamFelber so everything as normal, then?"
	expected3 := "so everything as normal, then?"
	runStripReply(test3, expected3, c)
}


func runStripReply(str1 string, str2 string, c *gocheck.C) {
	strippedStr := StripReply(str1)
	c.Assert(strippedStr, gocheck.Equals, str2)
}

func (s StringUtilsSuite) TestCanonicalize(c *gocheck.C) {
	test1 := "...really...?"
	expected1 := "really"
	runCanonicalize(test1, expected1, c)

	test2 := "GODDAMMIT!!"
	expected2 := "goddammit"
	runCanonicalize(test2, expected2, c)

	test3 := "normal,"
	expected3 := "normal"
	runCanonicalize(test3, expected3, c)

	test4 := "then?ever)"
	expected4 := "thenever"
	runCanonicalize(test4, expected4, c)
}

func runCanonicalize(str1 string, str2 string, c *gocheck.C) {
	strippedStr := Canonicalize(str1)
	c.Assert(strippedStr, gocheck.Equals, str2)
}
