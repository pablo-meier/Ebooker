package main

import (
	"ebooker/logging"

	"fmt"
	"launchpad.net/gocheck"
	"math"
	"testing"
)

// TODO:
//  Test text with punctuation, hashtags, etc.

// hook up gocheck into the gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

type MarkovSuite struct{}

var _ = gocheck.Suite(&MarkovSuite{})

type SuffixFreqTest struct {
	prefix string
	suffix string
	count  int
}

type GenFreqTest struct {
	prefix string
	suffix string
	prob   float64
}

// Reduce some boilerplate.
func makeGenerator(prefix, limit int) *Generator {
	return CreateGenerator(prefix, limit, &logging.LogMaster{})
}

// To test AddSeeds, we create a Generator, feed it some text, and ensure:
//   * The existence of all prefixes of the specified length.
//   * The existence of appropriate suffixes for a number of the prefixes.
//   * The correct frequency counts on suffixes.
func (s MarkovSuite) TestAddSeeds(c *gocheck.C) {
	gen := makeGenerator(2, 140)

	// Test basic case, prefix length of 2, no tricky tokenization.
	c.Assert(gen.PrefixLen, gocheck.Equals, 2)
	c.Assert(gen.CharLimit, gocheck.Equals, 140)

	gen.AddSeeds("today is a great day to be me")

	expectedPrefixes := []string{"today is", "is a", "a great", "great day", "day to", "to be"}

	for i := 0; i < len(expectedPrefixes); i++ {
		assertHasPrefix(gen.Data, expectedPrefixes[i], c)
	}

	gen.AddSeeds("today is a terrible day to be me")
	gen.AddSeeds("today is a terrible day to be me")
	gen.AddSeeds("today is a terrible day to be me")
	gen.AddSeeds("today is never so beautiful as tomorrow")

	// My kingdom for a tuple literal type!!!
	suffixTests := []SuffixFreqTest{SuffixFreqTest{"today is", "a", 4},
		SuffixFreqTest{"today is", "never", 1},
		SuffixFreqTest{"is a", "great", 1},
		SuffixFreqTest{"is a", "terrible", 3},
		SuffixFreqTest{"a great", "day", 1},
		SuffixFreqTest{"a terrible", "day", 3},
		SuffixFreqTest{"great day", "to", 1},
		SuffixFreqTest{"terrible day", "to", 3},
		SuffixFreqTest{"day to", "be", 4},
		SuffixFreqTest{"to be", "me", 4},
		SuffixFreqTest{"is never", "so", 1},
		SuffixFreqTest{"never so", "beautiful", 1},
		SuffixFreqTest{"so beautiful", "as", 1},
		SuffixFreqTest{"beautiful as", "tomorrow", 1}}

	for i := 0; i < len(suffixTests); i++ {
		triple := suffixTests[i]
		assertSuffixFrequencyCount(gen.Data, triple.prefix, triple.suffix, triple.count, c)
	}
}

// Similar to TestAddSeeds, we want to ensure that we can add multiple
// representations for the same canonical prefix, e.g. "Daddy says" ==
// "daddy says" == "DADDY SAYS"
func (s MarkovSuite) TestRepresentationCount(c *gocheck.C) {
	gen := makeGenerator(2, 140)
	gen.CanonicalizeSources()

	gen.AddSeeds("I've NEVER BEEN so mad")
	gen.AddSeeds("Ive never \"been\" so sad")
	gen.AddSeeds("IVE NEVER KILLED A MAN STOP ASKING")
	gen.AddSeeds("you have been so sad!!!")

	assertHasPrefix(gen.Data, "ive never", c)
	assertHasPrefix(gen.Reps, "ive", c)
	assertHasPrefix(gen.Reps, "never", c)
	assertSuffixFrequencyCount(gen.Data, "ive never", "been", 2, c)
	assertSuffixFrequencyCount(gen.Reps, "ive", "IVE", 1, c)
	assertSuffixFrequencyCount(gen.Reps, "ive", "I've", 1, c)
	assertSuffixFrequencyCount(gen.Reps, "ive", "Ive", 1, c)
	assertSuffixFrequencyCount(gen.Reps, "never", "never", 1, c)
	assertSuffixFrequencyCount(gen.Reps, "never", "NEVER", 2, c)

	assertHasPrefix(gen.Reps, "been", c)
	assertHasPrefix(gen.Reps, "so", c)
	assertSuffixFrequencyCount(gen.Data, "been so", "sad", 2, c)
	assertSuffixFrequencyCount(gen.Data, "been so", "mad", 1, c)
	assertSuffixFrequencyCount(gen.Reps, "mad", "mad", 1, c)
	assertSuffixFrequencyCount(gen.Reps, "sad", "sad!!!", 1, c)
	assertSuffixFrequencyCount(gen.Reps, "sad", "sad", 1, c)
}

// Much harder to test in that we require some level of randomness.
// Essentially, after we generate the appropriate data model, we'll run
// Generate several hundred or thousand times. We then see if the approximate
// number of times that each result came up corresponds to its probability.
//
// It's situations like these that make me wish I paid more attention in Unit
// Testing workshops,since I'm sure there's a construct out there to test that
// your random map works without, you know, requiring pseudorandom input...
//
// * First we ensure that when we generate off a single prefix with a single
//   invocation of AddSeeds, we produce the same sentence.
//
// * We then call AddSeeds multiple times, then generate on a prefix thousands
//   of times.
func (s MarkovSuite) TestGenerateText(c *gocheck.C) {
	// Single sentence case.
	gen := makeGenerator(2, 140)
	gen.AddSeeds("today is a great day to be alive")

	returnText := gen.GenerateFromPrefix("today is")

	c.Assert(returnText, gocheck.Equals, "today is a great day to be alive")

	gen.AddSeeds("today is a terrible day to be baking")
	gen.AddSeeds("today is a terrible day to be smelling")
	gen.AddSeeds("today is a great day to be moulting")

	tests := []GenFreqTest{GenFreqTest{"is a", "great", 0.5},
		GenFreqTest{"a great", "day", 1.0},
		GenFreqTest{"a terrible", "day", 1.0},
		GenFreqTest{"great day", "to", 1.0},
		GenFreqTest{"terrible day", "to", 1.0},
		GenFreqTest{"day to", "be", 1.0},
		GenFreqTest{"to be", "alive", 0.25},
		GenFreqTest{"to be", "moulting", 0.25},
		GenFreqTest{"to be", "smelling", 0.25},
		GenFreqTest{"to be", "baking", 0.25}}

	for i := 0; i < len(tests); i++ {
		curr := tests[i]
		assertProperFrequencyGeneration(gen, curr.prefix, curr.suffix, curr.prob, c)
	}
}

func assertHasPrefix(aMap CountedStringMap, prefix string, c *gocheck.C) {
	_, exists := aMap[prefix]
	if !exists {
		fmt.Printf("failure to find prefix \"%s\"", prefix)
	}
	c.Assert(exists, gocheck.Equals, true)
}

func assertSuffixFrequencyCount(aMap CountedStringMap, prefix, suffix string, count int, c *gocheck.C) {
	assertHasPrefix(aMap, prefix, c)
	countedStr, exists := aMap[prefix].GetSuffix(suffix)

	c.Assert(exists, gocheck.Equals, true)

	if count != countedStr.hits {
		fmt.Printf("expecting %d and got %d for '%s' -> '%s'", count, countedStr.hits, prefix, suffix)
	}
	c.Assert(countedStr.hits, gocheck.Equals, count)
}

func assertProperFrequencyGeneration(g *Generator, prefix, suffix string, prob float64, c *gocheck.C) {

	trials := 5000
	hits := 0

	epsilon := float64(0.075)

	for i := 0; i < trials; i++ {
		nextWord, shouldTerminate, _, _ := g.PopNextWord(prefix, 100)
		c.Assert(shouldTerminate, gocheck.Equals, false)
		if suffix == nextWord {
			hits++
		}
	}

	success := math.Abs(((float64(hits) / float64(trials)) - prob)) < epsilon

	if !success {
		fmt.Printf("%s -> %s had probability %f, expected %f\n", prefix, suffix, float64(hits)/float64(trials), prob)
	}

	c.Assert(success, gocheck.Equals, true)
}
