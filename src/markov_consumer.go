/* 
Package for the actual consumption of corpus text, adding to the data model 
for generation. Also contains the functions to generate text from that data 
model.

We first build a list of "Prefixes" (sets of words that come up together, 
such as "hot dog" and "dog food"), and build a map to what word normally 
follows the prefix, weighted by frequency. To generate text, we select 
a prefix, use the frequency to pick a the most likely word to follow it,
then append the new word the prefix while removing the first, forming a new
prefix. We do this until we create a never-before seen prefix (in which case
the sentence ends) or we hit the character limit (this is meant to create
tweets, after all).

An alternative idea would be to store prefix tables of different lengths, 
e.g. prefix length of 3 words would only be a -maximum- prefix length,
but we generate prefix tables for prefixes of 1, 2, and 3 words. If we 
run into a 3-word prefix that contains no match, we trim it to 2 words 
and look for prefixes there, else 1 word, etc. Later we would try to
"gobble" words back up to a 3 word prefix.

Note there is a great codewalk of this algo on the Go website
(http://golang.org/doc/codewalk/markov/), but I'm only taking the algo and
supplying my own implementation, since it's more fun and gives me more insight
into Go, and the problem at hand.

Examples of awesome Markov Twitter bots: 
@RandomTedTalks, @kpich_ebooks, @MarkovBible
*/


package main

import (
    "fmt"
)

// Suffix contains the suffix to any given string, and the probability that
// we encounter it.
type Suffix struct {
    probability float64
    frequency uint32
    suffix string
}

// MarkovMap is the map that attaches prefixes to suffixes.
type MarkovMap map[string]Suffix

// Gives us all we need to build a fresh data model to generate from:
// the MarkovMap for the actual data, and the prefix length.
type Generator struct {
    prefixLength uint32
    charLimit uint32
    dataModel MarkovMap
}


// AddSeeds takes in a string, breaks it into prefixes, and adds it to the 
// data model. Note that the data model isn't ready to use at this point,
// since we need to use the frequencies to calculate the probabilities.
func (g Generator) AddSeeds(s string) {

}

// Generates text from the given generator. It stops when the character limit
// has run out, or it encounters a prefix it has no suffixes for.
func (g Generator) GenerateText() string {
    return ""
}


// Mainline to generate a binary -_-
func main() {
    fmt.Println("I'm alive!")
}
