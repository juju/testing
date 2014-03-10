// Copyright 2012-2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package logging

import (
	"fmt"
	"time"

	"github.com/juju/loggo"
	gc "launchpad.net/gocheck"

	"github.com/juju/testing"
)

// LoggingSuite redirects the juju logger to the test logger
// when embedded in a gocheck suite type.
type LoggingSuite struct {
	testing.CleanupSuite
}

type gocheckWriter struct {
	c *gc.C
}

func (w *gocheckWriter) Write(level loggo.Level, module, filename string, line int, timestamp time.Time, message string) {
	// Magic calldepth value...
	w.c.Output(3, fmt.Sprintf("%s %s %s", level, module, message))
}

func (t *LoggingSuite) SetUpSuite(c *gc.C) {
	t.CleanupSuite.SetUpSuite(c)
	t.setUp(c)
	t.AddSuiteCleanup(func(*gc.C) {
		loggo.ResetLoggers()
		loggo.ResetWriters()
	})
}

func (t *LoggingSuite) SetUpTest(c *gc.C) {
	t.CleanupSuite.SetUpTest(c)
	t.setUp(c)
}

func (t *LoggingSuite) setUp(c *gc.C) {
	loggo.ResetWriters()
	loggo.ReplaceDefaultWriter(&gocheckWriter{c})
	loggo.ResetLoggers()
}
