package main

/*
A lot of string manipulation happens in ebooker: We have to strip replies, we
have to uppercase and lowercase for various effects, create new hashtags. This
package lets us write our transformations.
*/

import (
	"regexp"
	"strings"
)

// Removes the reply-to field from the text of a tweet, e.g. "@SrPablo forget
// it!" -> "forget it!" Also removes "public directed tweets"; those tweets that
// are effectively replies but designed to be seen publicly, such as ".@walmart
// what happens when u die?"
func StripReply(str string) string {
	replyCheck, _ := regexp.Compile("^(?:\\s*\\.)?(\\s*@[a-zA-Z0-9_=]+\\s*)")
	return removePattern(str, replyCheck)
}

// A string is "canonicalized" when it's content has been "neutralized" of what
// data other than it's core content. We optionally do this on maps to ensure
// that, for  example, odd capitalizations, or hashtags, or punctuation won't
// block the algo from continuing to chain.
//
// Note that this is meant to be run not on an entire tweet, but on the tokens
// that come after splitting into it's non-whitespace components. Also note that
// we don't "lose" what we strip away -- we put it back in the output phase.
//
// A token is in canonical form when:
//   * It is lowercased.
//   * All punctuation has been removed.
func Canonicalize(str string) string {
	punctuation, _ := regexp.Compile("([\"'\\[\\]().,;:{}|\\\\+=_\\-?<>!@#$%^&*]|\\s)+")
	stripped := removeAllOfPattern(str, punctuation)
	return strings.ToLower(stripped)
}

// Removes the first parts of a string that match the pattern.
func removePattern(str string, pattern *regexp.Regexp) string {
	found := pattern.Find([]byte(str))

	if found != nil {
		return strings.Replace(str, string(found), "", 1)
	}

	return str
}

// removePattern only removes the first instance of a pattern on a string, this
// function removes all instances.
func removeAllOfPattern(str string, pattern *regexp.Regexp) string {
	compare := removePattern(str, pattern)
	for compare != removePattern(compare, pattern) {
		compare = removePattern(compare, pattern)
	}
	return compare
}
