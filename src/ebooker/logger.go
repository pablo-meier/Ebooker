package ebooker

import (
    "log"
    "os"
    "strings"
)

type LogMaster struct {
    logger *log.Logger
    silent bool
    debug bool
}


func GetLogMaster(silent, debug, timestamps bool) LogMaster {
    var flags int
    if timestamps {
        flags = log.LstdFlags
    } else {
        flags = 0
    }

    return LogMaster{log.New(os.Stdout, "", flags), silent, debug }
}


func (l LogMaster) StatusWrite(format string, a ...interface{}) {
    l.writeToAll(addTag("(S)", format), a...)
}

func (l LogMaster) DebugWrite(format string, a ...interface{}) {
    if l.debug {
        l.writeToAll(addTag("(D)", format), a...)
    }
}

func (l LogMaster) writeToAll(format string, a ...interface{}) {
    if !l.silent {
        l.logger.Printf(format, a...)
    }
}

func addTag(tag, format string) string {
    return strings.Join([]string{ tag, format}, " - ")
}
