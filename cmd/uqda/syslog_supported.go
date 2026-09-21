//go:build !windows

package main

import (
	"github.com/gologme/log"
	gsyslog "github.com/hashicorp/go-syslog"

	"github.com/Uqda/Core/src/version"
)

func newSystemLogger() *log.Logger {
	syslogger, err := gsyslog.NewLogger(gsyslog.LOG_NOTICE, "DAEMON", version.BuildName())
	if err != nil {
		return nil
	}
	return log.New(syslogger, "", log.Flags()&^(log.Ldate|log.Ltime))
}
