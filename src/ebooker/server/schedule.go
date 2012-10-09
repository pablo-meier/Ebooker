package main

/*
File contains the functions meant to schedule tweets. We use the cron model
for specifying when, and how often.
*/

import (
	"time"
)

const (
	ALL = "*"
)

const CHANNEL_BUFFER = 3

type Schedule struct {
	fireOff   chan time.Time
	killOrder bool

	minute     string
	hour       string
	dayOfMonth string
	month      string
	dayOfWeek  string
}

// Returns when the next Tick should be from now.
func (s *Schedule) next() time.Duration {
	return s.nextFromTime(time.Now())
}

// Isolated for testing.
func (s *Schedule) nextFromTime(t time.Time) time.Duration {
	d, _ := time.ParseDuration("12h")
	return d
}

// start runs the schedule, having it send ticks at the times specified upon
// creation.
func (s *Schedule) start() {
	duration := s.next()

	c := time.After(duration)
	for _ = range c {
		if s.shouldKill() {
			break
		}
		s.fireOff <- time.Now()
		duration = s.next()
		c = time.After(duration)
	}
}

// We then calculate the difference between time.Now() and the next tick,
// and set an After()
func cronParse(s string) Schedule {
	fireOff := make(chan time.Time, CHANNEL_BUFFER)
	schedule := Schedule{fireOff, false, "0", "11,19", ALL, ALL, ALL}

	go schedule.start()

	return schedule
}

// TickingChannel returns the channel we 'tick' on whenever we need to send
// a new Tweet
func (s *Schedule) tickingChannel() chan time.Time {
	return s.fireOff
}

func (s *Schedule) shouldKill() bool {
	return s.killOrder
}

func (s *Schedule) kill() {
	s.killOrder = true
	s.fireOff <- time.Now()
}
