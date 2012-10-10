package main

import (
	"ebooker/logging"
	"ebooker/oauth1"

	"time"
)

/*
All the data/functions for the bots.
*/

type Bot struct {
	username string
	sources  []string
	gen      *Generator
	token    *oauth1.Token
	sched    *Schedule

	logger *logging.LogMaster
	data   *DataHandle
	oauth  *oauth1.OAuth1
	tf     *TweetFetcher
}

// Runs perpetually, forever tweeting
func (b *Bot) Run() {
	b.logger.StatusWrite("Bot %s ordered to run! Away we go!\n", b.username)

	c := b.sched.tickingChannel()
	for _ = range c {
		if b.sched.shouldKill() {
			b.logger.StatusWrite("Bot %s received killing order! Dying...\n", b.username)
			break
		}
		b.logger.StatusWrite("At %v bot %s received the order to tweet.\n", time.Now(), b.username)

		// Update Sources
		newSources := fetchNewSources(b.sources, b.token, b.data, b.logger, b.tf)

		// Add new seeds to generator.
		for _, val := range newSources {
			b.gen.AddSeeds(val)
		}

		// fire off the new tweet
		message := b.gen.GenerateText()
		b.logger.StatusWrite("Sending \"%s\"\n", message)
		b.tf.sendTweet(message, b.token)
		b.logger.StatusWrite("Success! Next tweet due after %v\n", b.sched.next().String())
	}
}

func (b *Bot) Kill() {
	b.sched.kill()
}
