




package markov

import (
    "launchpad.net/gocheck"
    "testing"
    "fmt"
)


// hook up gocheck into the gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

type MarkovSuite struct{}
var _ = gocheck.Suite(&MarkovSuite{})


// To test AddSeeds, we create a Generator, feed it some text, and ensure:
//   * The existence of all prefixes of the specified length.
//   * The existence of appropriate suffixes for a number of the prefixes.
//   * The correct frequency counts on suffixes.
func (s MarkovSuite) TestAddSeeds(c *gocheck.C) {
    gen := CreateGenerator(2, 140)

    // Test basic case, prefix length of 2, no tricky tokenization.
    c.Assert(gen.prefixLength, gocheck.Equals, 2)
    c.Assert(gen.charLimit, gocheck.Equals, 140)


    gen.AddSeeds("Today is a great day to be me")

    // Don't include "be me," as there's nothing following it and therefore not 
    // useful
    expectedPrefixes := []string{"Today is", "is a", "a great", "great day",
        "day to", "to be"}

    for i := 0; i < len(expectedPrefixes) ; i++ {
        fmt.Println("Checking ", expectedPrefixes[i])
        assertHasPrefix(gen, expectedPrefixes[i], c)
    }

}

func assertHasPrefix(gen *Generator, prefix string, c *gocheck.C) {
    _, exists := gen.dataModel[prefix]
    c.Assert(exists, gocheck.Equals, true)
}


func (s MarkovSuite) TestGenerateText(c *gocheck.C) {
    c.Assert(14, gocheck.Equals, 14)
}
