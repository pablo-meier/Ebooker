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

An alternative idea would be to store prefix tables of different lengths, e.g.
prefix length of 3 words would really be a -maximum- prefix length, but we
generate prefix tables for prefixes of 1, 2, and 3 words. If we run into a
3-word prefix that contains no match, we trim it to 2 words and look for
prefixes there, else 1 word, etc. Later we would try to "gobble" words back up
to a 3 word prefix.

Note there is a great codewalk of this algo on the Go website
(http://golang.org/doc/codewalk/markov/), but I'm only taking the basic algo and
supplying my own implementation, since it's more fun and gives me more insight
into Go, and the problem at hand.

Examples of awesome Markov Twitter bots: 
@RandomTedTalks, @kpich_ebooks, @MarkovBible
*/

package markov

import (
    "fmt"
    "strings"
    "math/rand"
)

const DEBUG = true;

// Suffix and SuffixList contain what is necessary to calculate the next block 
// of text from a prefix, which includes the frequency and string content of 
// of each particular suffix, as well as the the total frequency of all
// suffixes.
type Suffix struct {
    hits int
    str string
}

type SuffixList struct {
    slice []*Suffix
    total int
}

// MarkovMap is the map that attaches prefixes to suffixes.
type MarkovMap map[string]*SuffixList

// Generators gives us all we need to build a fresh data model to generate 
// from: the MarkovMap for the actual data, as well as the parametrized 
// constraints on the text generation.
type Generator struct {
    prefixLen int
    charLimit int
    data MarkovMap
}


// CreateGenerator returns a Generator that is fully initialized and ready for 
// use.
func CreateGenerator(prefixLen int, charLimit int) *Generator {
    markov := make(MarkovMap)
    return &Generator{ prefixLen, charLimit, markov }
}

func createNewSuffix(str string) *Suffix {
    return &Suffix{ 1, str }
}


// AddSeeds takes in a string, breaks it into prefixes, and adds it to the 
// data model. Note that the data model isn't ready to use at this point,
// since we need to use the frequencies to calculate the probabilities.

// TODO NAMES, REFACTOR BLAHBLAHBLAH
func (g Generator) AddSeeds(input string) {
    words := tokenize(input)

    for len(words) > g.prefixLen {
        prefix := strings.Join(words[0:g.prefixLen], " ")

        if suffixList, exists := g.data[prefix]; exists {
            str := words[g.prefixLen]
            if suffix, member := hasSuffix(suffixList, str); member {
                suffix.hits++
            } else {
                suffix = createNewSuffix(str)
                suffixList.slice = append(suffixList.slice, suffix)
            }
            suffixList.total++
        } else {
            str := words[g.prefixLen]
            suffix := createNewSuffix(str)
            suffixSlice := make([]*Suffix, 0)
            suffixSlice = append(suffixSlice, suffix)
            suffixList := &SuffixList{ suffixSlice, 1 }

            g.data[prefix] = suffixList
        }

        words = words[1:]
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


// hasSuffix searches a SuffixList for one that contains the string, and 
// returns the suffix (if applicable) and a boolean describing whether or not 
// we found it.
func hasSuffix(suffixlist *SuffixList, lookFor string) (*Suffix, bool) {
    slice := suffixlist.slice
    for i := 0; i < len(slice); i++ {
        curr := slice[i]
        if curr.str == lookFor {
            return curr, true
        }
    }

    return createNewSuffix(""), false
}

// Generates text from the given generator. It stops when the character limit
// has run out, or it encounters a prefix it has no suffixes for.
func (g Generator) GenerateText() string {
    return g.GenerateFromPrefix(g.randomPrefix())
}

// We expose this version primarily for testing.
func (g Generator) GenerateFromPrefix(prefix string) string {

    result := []string{ prefix }
    charLimit := g.charLimit

    // gotchas: what if we generate a prefix that isn't in the map?
    // proper termination conditions.
    for ;; {
        //debug("Entering popNextWord -- prefix is", prefix, "and charLimit is", charLimit)
        word, shouldTerminate, newPrefix, newCharLimit := g.PopNextWord(prefix, charLimit)
        prefix = newPrefix
        charLimit = newCharLimit

        //debug("returned! word is", word, "shouldTerminate is", shouldTerminate)
        //debug("newPrefix is", newPrefix, "newCharLimit is", newCharLimit)
        if shouldTerminate {
            break;
        } else {
            result = append(result, word)
            //debug("result will now be", strings.Join(result, " "))
        }
    }

    return strings.Join(result, " ")
}

func (g Generator) PopNextWord(prefix string, limit int) (string, bool, string, int) {
    suffixlist, exists := g.data[prefix]

    if exists {
        index := rand.Intn(suffixlist.total) + 1
        //debug("Random index is", index)
        slice := suffixlist.slice
        for i := 0; i < len(slice); i++ {
            //debug("Testing if", index, "is <=", slice[i].hits)
            if index <= slice[i].hits {
                candidate := slice[i].str
                //debug("It is! Candidate is", candidate)
                if addsTo := len(candidate) + 1; addsTo <= limit {
                    shifted := append(strings.Split(prefix, " ")[1:], candidate)
                    newPrefix := strings.Join(shifted, " ")
                    newLimit := limit - addsTo
                    return candidate, false, newPrefix, newLimit
                }
            }
            //debug("moving on...")
            index -= slice[i].hits
        }
    }
    return "", true, "", 0
}

func (g Generator) randomPrefix() string {
    index := rand.Intn(len(g.data))
    count := 0
    for k, _ := range g.data {
        if count == index {
            return k
        }
        count++
    }
    // Shouldn't ever get here! Satisfy the compiler tho
    return ""
}

// For testing.
func (s SuffixList) GetSuffix(lookFor string) (*Suffix, bool) {
    for i := 0; i < len(s.slice); i++ {
        if s.slice[i].str == lookFor {
            return s.slice[i], true
        }
    }
    return createNewSuffix(""), false
}

func debug(str string) {
    if (DEBUG) {
        fmt.Println(str)
    }
}

