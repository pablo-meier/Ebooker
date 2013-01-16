Ebooker
=======

Your way to abstract poetry stardom!

Ebooker is a service that will consume a Twitter stream and generate new text
based on a Markov model of randomization. This is nothing new: most people
speculate @horse\_ebooks came about this way (but I'm a believer it's just
Markov-assisted and not truly botty). Excellent examples of Markov-based Twitter
accounts are [@kpich\_ebooks][4], [@RandomTEDTalks][3], and [@markov\_bible][5].

If you're new to Markov chains in text generation, I wrote [a post][1] on it
trying to explain it in non-technical terms. Conversations with my non-technical
friends who tried to read it give mixed reviews to how much they actually
understood it.

Implemented in [Go][2], because the gopher is cute and I was curious.

Dependencies
============

* sqlite3

That's pretty much it. We use it to store tweets and OAuth tokens.

[Basic][6] Structure
====================

Running `make` (and assuming the project root is in your `$GOPATH`) will
generate two binaries: `ebooker\_server` and `ebooker\_client`. Both, when run
with `--help`, will list their flags.

The server is what does all the work: it retrieves tweets from Twitter,
generates Markov chain tweets, and posts them up on a schedule (by default,
every 11 hours). The schedule was meant to be more configurable, but I decided
to work on other things before I implemented it.

The client is your way of telling the server what to do: you call it with the
appropriate flags to add, list, or delete bots. You can also just call it with
sources to generate Markov tweet text, printed to stdout, and skip the bot
business altogether.

Should I use this to learn Go?
==============================

Probably not! There are a few major faux pas that I note, not that it's been a
few months. Just like [ScrabbleCheat][6] has elements of "Baby's First Erlang,"
this smacks of Baby's First Go. Notably:

* I got scoping all messed up. Many things are exported because I figured they
  had to cross file boundaries, when that really makes them public and exports
  them from the package.

* I got packaging all wrong too. If you go back in the repo enough revisions,
  you'll see that for a while, this was all one big package. When I started
  separating them out, I found many a boo-boo. This exacerbated the visibility
  issue above.

* Some structures that get thrown around a lot (OAuth1, LogMaster, etc.) just
  reek of needing some proper DI tool.

* I only tested what needed testing. Fail code coverage, test-first, etc. I'm
  obviously crying every night over this.

Anything else?
==============

I hand-rolled my own OAuth1 for this. Like all my side projects, this was
simply because it was more fun to do it that way. The relevant code is in the
`ebooker/oauth1` package. Both the client and the server use it: the server
obviously for posting tweets, and the client for generating access tokens if you
try to make a bot that needs to post to an account you don't have tokens for
currently. If you retrieve tokens another way (say, with an HTML frontend and a
proper "Sign in with Twitter" redirect) you can input them to the server
directly.

   [1]: http://morepaul.com/2012/10/loving-yourself-with-ebooks.html
   [2]: http://golang.org/
   [3]: http://twitter.com/RandomTEDTalks
   [4]: http://twitter.com/kpich_ebooks
   [5]: http://twitter.com/markov_bible
   [6]: http://www.youtube.com/watch?v=6WJFjXtHcy4
