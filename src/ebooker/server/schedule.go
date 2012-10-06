package main

/*
File contains the functions meant to schedule tweets. We use the cron model
for specifying when, and how often.
*/

import (
	"time"
)

// should receive a Task T to do on every 'tick,' and a specification
// (originally by input string in cron format, which we parse and store).
// We then calculate the difference between time.Now() and the next tick,
// and set an After()

func cronParse(s string) string {
	tm := time.Now()
	return tm.String()
}
