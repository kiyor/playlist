/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : logger.go

* Purpose :

* Creation Date : 05-31-2015

* Last Modified : Sun 31 May 2015 07:09:08 PM UTC

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : logger.go

* Purpose :

* Creation Date : 02-26-2015

* Last Modified : Wed 01 Apr 2015 09:58:27 PM PDT

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package main

import (
	"github.com/op/go-logging"
	"os"
)

type LogOptions struct {
	Name      string
	ShowErr   bool
	ShowDebug bool
	ShowColor bool
}

var Logger *logging.Logger

func NewLogger(options *LogOptions) *logging.Logger {
	log := logging.MustGetLogger(options.Name)

	// init default to null
	var out, err *os.File
	if options.ShowErr {
		err = os.Stderr
	}
	if options.ShowDebug {
		out = os.Stdout
	}

	// setup logger
	stdout := logging.NewLogBackend(out, "", 0)
	stderr := logging.NewLogBackend(err, "", 0)

	format := logging.MustStringFormatter(
		"%{time:15:04:05.000} [" + options.Name + "] %{level:.4s} %{id:03x} %{shortfile} %{shortfunc} ▶ \"%{message}\"",
	)
	if options.ShowColor {
		format = logging.MustStringFormatter(
			"%{color}%{time:15:04:05.000} [" + options.Name + "] %{level:.4s} %{id:03x} %{shortfile} %{shortfunc} ▶%{color:reset} \"%{message}\"",
		)
	}

	stdoutFormatter := logging.NewBackendFormatter(stdout, format)
	stderrFormatter := logging.NewBackendFormatter(stderr, format)

	stderrLeveled := logging.AddModuleLevel(stderrFormatter)
	stdoutLeveled := logging.AddModuleLevel(stdoutFormatter)

	stdoutLeveled.SetLevel(logging.DEBUG, "")
	stderrLeveled.SetLevel(logging.ERROR, "")

	logging.SetBackend(stdoutLeveled, stderrLeveled)

	return log
}
