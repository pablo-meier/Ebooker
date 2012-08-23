




package markov

import (
    "launchpad.net/gocheck"
    "testing"
)


// hook up gocheck into the gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

type MarkovSuite struct{}
var _ = gocheck.Suite(&MarkovSuite{})

type SuffixFreqTest struct {
    prefix string
    suffix string
    count int
}

// To test AddSeeds, we create a Generator, feed it some text, and ensure:
//   * The existence of all prefixes of the specified length.
//   * The existence of appropriate suffixes for a number of the prefixes.
//   * The correct frequency counts on suffixes.
func (s MarkovSuite) TestAddSeeds(c *gocheck.C) {
    gen := CreateGenerator(2, 140)

    // Test basic case, prefix length of 2, no tricky tokenization.
    c.Assert(gen.prefixLen, gocheck.Equals, 2)
    c.Assert(gen.charLimit, gocheck.Equals, 140)


    gen.AddSeeds("Today is a great day to be me")

    // Don't include "be me," as there's nothing following it and therefore not 
    // useful/included
    expectedPrefixes := []string{"Today is", "is a", "a great", "great day",
        "day to", "to be"}

    for i := 0; i < len(expectedPrefixes) ; i++ {
        assertHasPrefix(gen, expectedPrefixes[i], c)
    }

    gen.AddSeeds("Today is a terrible day to be me")
    gen.AddSeeds("Today is a terrible day to be me")
    gen.AddSeeds("Today is a terrible day to be me")
    gen.AddSeeds("Today is never so beautiful as tomorrow")

    // My kingdom for a tuple type so that I could for loop this, as above!!
    suffixTests := []SuffixFreqTest{ SuffixFreqTest{"Today is", "a", 4},
        SuffixFreqTest{"Today is", "never", 1},
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
        assertSuffixFrequencyCount(gen, triple.prefix, triple.suffix, triple.count, c)
    }
}

func assertHasPrefix(gen *Generator, prefix string, c *gocheck.C) {
    _, exists := gen.data[prefix]
    c.Assert(exists, gocheck.Equals, true)
}

func assertSuffixFrequencyCount(gen *Generator, prefix, suffixStr string, count int, c *gocheck.C) {
    assertHasPrefix(gen, prefix, c)
    suffix, exists := gen.data[prefix].GetSuffix(suffixStr)

    c.Assert(exists, gocheck.Equals, true)

    c.Assert(suffix.hits, gocheck.Equals, count);
}

func (s MarkovSuite) TestAddSeedsOtherPrefixLength(c *gocheck.C) {

}


// Much harder to test in that we require some level of randomness. 
// Essentially, after we generate the appropriate data model, we'll run
// Generate several hundred or thousand times. We then see if the approximate
// number of times that each result came up corresponds to its probability.
//
// It's situations like these that make me wish I paid more attention in 
// Mocks, since I'm sure there's a construct out there to test that your
// random map works without, you know, requiring pseudorandom input...
func (s MarkovSuite) TestGenerateText(c *gocheck.C) {
    c.Assert(14, gocheck.Equals, 14)
}
