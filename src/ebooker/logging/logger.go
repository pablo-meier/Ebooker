/*
Functions for logging, error reporting, debugging... virtually everything that
prints. Like the Android logger, we distinguish between Debug messages,
Status messages, and making it easy to add any other kinds as necessary (e.g.
Android, IIRC, has "Info" messages). Each can be turned on or off as desired.

This also allows us to abstract away the output channel. Currently we only
write to stdout, but we could write to a file, a buffer, many at the same
time, etc.
*/
package logging

import (
	"log"
	"os"
	"strings"
)

// LogMaster is the struct containing all the logging methods, and contains all
// the information we'll need to simply "Do the right thing," per its
// configuration, when we ask to write Debug messages, Status messages, etc.
type LogMaster struct {
	logger *log.Logger
	silent bool
	debug  bool
}

// Creates a new LogMaster.
func GetLogMaster(silent, debug, timestamps bool) LogMaster {
	var flags int
	if timestamps {
		flags = log.LstdFlags
	} else {
		flags = 0
	}

	return LogMaster{log.New(os.Stdout, "", flags), silent, debug}
}

// Writes a new Status message to all the output writers we've given the
// LogMaster.
func (l LogMaster) StatusWrite(format string, a ...interface{}) {
	l.writeToAll(addTag("(S)", format), a...)
}

// Writes a new Debug message to all the output writers we've given the
// LogMaster.
func (l LogMaster) DebugWrite(format string, a ...interface{}) {
	if l.debug {
		l.writeToAll(addTag("(D)", format), a...)
	}
}

// Writes the message to all the output Writers we've given the LogMaster.
func (l LogMaster) writeToAll(format string, a ...interface{}) {
	if !l.silent {
		l.logger.Printf(format, a...)
	}
}

// Helper function to allow messages to "tag" themselves by type.
func addTag(tag, format string) string {
	return strings.Join([]string{tag, format}, " - ")
}
