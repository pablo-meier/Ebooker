/*
Package for the actual consumption of corpus text, adding to the data model
for generation. Also contains the functions to generate text from that data
model.

We first build a list of prefixes (sets of words that come up together, such as
"hot dog" and "dog food"), and build a map to what word normally follows the
prefix, weighted by frequency. To generate text, we select a prefix, use the
frequency to determine probabilities, and with some random data we pick a
plausible word to follow it. To repeat the process, we append the new word the
prefix while removing the first, forming a new prefix. We do this until we
create a never-before seen prefix (in which case the sentence ends) or we hit
the character limit (this is meant to create tweets, after all).

Further work may be in randomizing or otherwise diversifying the end
conditions on the text generation, such as occassionally deciding to start
over, or taking the average tweet length of the user in question into
consideration.

Note that we have two probabilistic maps in the code: one for prefixes to
suffixes (as "hot dog food", above) and another for the -representation- of each
word of text. The rationale for this was because I didn't want the words "Hot
Dog Food" and "hot dog food" and "Hot Dog! Food?" to be treated as seperate for
their capitalization or punctuation.  Since I didn't want to lose the data or
character of the odd capitalizations or punctuations, the code also records all
the representations of what we call "canonical" form of a word and select
probabilistically from them as well.

While this is still here and seemed cool at the time, the quality of the tweets
it produced was subpar, since sentences had odder-than-desired representional
noise. Usually, the feature is turned off.

Note that for tweets, it's unlikely we'll use any prefix length greater than
1, but it's useful to have in case we'd like to generate a larger output, like
michaelochurch screeds.

Note there is a great codewalk of a simple version of this algo on the Go
website (http://golang.org/doc/codewalk/markov/), but I'm only taking the basic
algo and supplying my own implementation, since it's more fun and gives me more
insight into Go, and the problem at hand.

Examples of awesome Markov Twitter bots:
@RandomTedTalks, @kpich_ebooks, @markov_bible
*/

package main

import (
	"ebooker/logging"

	"math/rand"
	"strings"
)

// Since both maps (the prefix -> suffix and canonical -> representation)
// operate about the same way, we abstract their representation into a notion
// of CountedStrings, where the values of the map contain both the string we
// care about and a count of how often it occurs.
type CountedString struct {
	hits int
	str  string
}

// A CountedStringList is a list of all the CountedStrings for a given prefix,
// and a total number of times that prefix occurs (necessary, with the
// CountedString hits, for probability calculation).
type CountedStringList struct {
	slice []*CountedString
	total int
}

// Map from a prefix in canonical form to CountedStringLists, where one will
// move canonical prefixes to suffixes, and another to words -> representation.
type CountedStringMap map[string]*CountedStringList

// Generators gives us all we need to build a fresh data model to generate
// from.
type Generator struct {
	PrefixLen  int
	CharLimit  int
	Data       CountedStringMap   // suffix map
	Reps       CountedStringMap   // representation map
	Beginnings []string           // acceptable ways to start a tweet.
	canon      bool               // map sources seperately from representations.
	logger     *logging.LogMaster // Lets us debug, emit status.
}

// CreateGenerator returns a Generator that is fully initialized and ready for
// use.
func CreateGenerator(prefixLen int, charLimit int, logger *logging.LogMaster) *Generator {
	markov := make(CountedStringMap)
	reps := make(CountedStringMap)
	beginnings := []string{}
	return &Generator{prefixLen, charLimit, markov, reps, beginnings, false, logger}
}

// Convenience method, already populating the first "hit" of the CountedString.
func createCountedString(str string) *CountedString {
	return &CountedString{1, str}
}

// AddSeeds takes in a string, breaks it into prefixes, and adds it to the
// data model.
func (g *Generator) AddSeeds(input string) {
	source := tokenize(StripReply(input))

	if g.canon {
		var canonical []string
		for i := 0; i < len(source); i++ {
			canonical = append(canonical, Canonicalize(source[i]))
			AddToMap(canonical[i], source[i], g.Reps)
		}
		source = canonical
	}

	first := true
	for len(source) > g.PrefixLen {
		prefix := strings.Join(source[0:g.PrefixLen], " ")
		AddToMap(prefix, source[g.PrefixLen], g.Data)
		source = source[1:]
		if first {
			g.Beginnings = append(g.Beginnings, prefix)
			first = false
		}
	}
}

// Add to map checks if the key/value pair exists in the map. If not, we create
// them, and if so, we either increment the counter on the value or initialize
// it if it didn't exist previously.
func AddToMap(prefix, toAdd string, aMap CountedStringMap) {

	if csList, exists := aMap[prefix]; exists {
		if countedStr, member := csList.hasCountedString(toAdd); member {
			countedStr.hits++
		} else {
			countedStr = createCountedString(toAdd)
			csList.slice = append(csList.slice, countedStr)
		}
		csList.total++
	} else {
		countedStr := createCountedString(toAdd)
		countedStrSlice := make([]*CountedString, 0)
		countedStrSlice = append(countedStrSlice, countedStr)
		csList := &CountedStringList{countedStrSlice, 1}

		aMap[prefix] = csList
	}
}

// tokenize splits the input string into "words" we use as prefixes and
// suffixes. We can't do a naive 'split' by a separator, or even a regex '\W'
// due to corner cases, and the nature of the text we intend to capture: e.g.
// we'd like "forty5" to parse as such, rather than "forty" with "5" being
// interpreted as a "non-word" character. Similarly with hashtags, etc.
func tokenize(input string) []string {
	return strings.Split(input, " ")
}

// hasCountedString searches a CountedStringList for one that contains the string, and
// returns the suffix (if applicable) and a boolean describing whether or not
// we found it.
func (l CountedStringList) hasCountedString(lookFor string) (*CountedString, bool) {
	slice := l.slice
	for i := 0; i < len(slice); i++ {
		curr := slice[i]
		if curr.str == lookFor {
			return curr, true
		}
	}

	return createCountedString(""), false
}

// Generates text from the given generator. It stops when the character limit
// has run out, or it encounters a prefix it has no suffixes for.
func (g *Generator) GenerateText() string {
	return g.GenerateFromPrefix(g.randomPrefix())
}

// We expose this version primarily for testing.
func (g *Generator) GenerateFromPrefix(prefix string) string {

	g.logger.DebugWrite("Generating text from prefix \"%s\"\n", prefix)

	// Representation gets a special case, since you can have a multi-word
	// prefix (e.g. "Paul is") but each word needs it's own representation
	// (e.g. "PAUL" "is" or "pAUL" "Is"). Note that this can break if your
	// prefix's rep is longer than the charLimit, should we generalize
	var result []string
	charLimit := g.CharLimit

	if g.canon {
		split := strings.Split(prefix, " ")
		for _, token := range split {
			rep := g.Reps[token].DrawProbabilistically()
			charLimit -= len(rep)
			result = append(result, rep)
		}
	} else {
		result = append(result, prefix)
	}

	for {
		word, shouldTerminate, newPrefix, newCharLimit := g.PopNextWord(prefix, charLimit)
		prefix = newPrefix
		charLimit = newCharLimit

		if shouldTerminate {
			break
		} else {
			result = append(result, word)
			g.logger.DebugWrite("New Prefix is \"%s\", %d characters remain\n", newPrefix, newCharLimit)
		}
	}

	return strings.Join(result, " ")
}

func (g *Generator) PopNextWord(prefix string, limit int) (string, bool, string, int) {

	csList, exists := g.Data[prefix]

	if !exists {
		g.logger.DebugWrite("Prefix does not exist, terminating this run.\n")
		return "", true, "", 0 // terminate path
	}
	successor := csList.DrawProbabilistically()
	g.logger.DebugWrite("Drew \"%s\" as successor to \"%s\".\n", successor, prefix)
	var rep string
	if g.canon {
		rep = g.Reps[successor].DrawProbabilistically()
		g.logger.DebugWrite("After probabilisic draw, successor is respresented by \"%s\"\n", rep)
	} else {
		rep = successor
	}

	addsTo := len(rep) + 1

	if addsTo <= limit {
		shifted := append(strings.Split(prefix, " ")[1:], rep)
		newPrefix := strings.Join(shifted, " ")
		newLimit := limit - addsTo
		return rep, false, newPrefix, newLimit
	}
	g.logger.DebugWrite("Exceeding character limit. Terminating run.\n")
	return "", true, "", 0
}

func (g *Generator) CanonicalizeSources() {
	g.canon = true
}

func (cs CountedStringList) DrawProbabilistically() string {
	index := rand.Intn(cs.total) + 1
	for i := 0; i < len(cs.slice); i++ {
		if index <= cs.slice[i].hits {
			return cs.slice[i].str
		}
		index -= cs.slice[i].hits
	}
	return ""
}

func (g *Generator) randomPrefix() string {
	index := rand.Intn(len(g.Beginnings))
	return g.Beginnings[index]
}

// For testing.
func (s *CountedStringList) GetSuffix(lookFor string) (*CountedString, bool) {
	for i := 0; i < len(s.slice); i++ {
		if s.slice[i].str == lookFor {
			return s.slice[i], true
		}
	}
	return createCountedString(""), false
}
