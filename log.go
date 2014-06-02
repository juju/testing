// Copyright 2012-2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"fmt"
	"os"
	"time"

	"github.com/juju/loggo"
	gc "launchpad.net/gocheck"
)

// LoggingSuite redirects the juju logger to the test logger
// when embedded in a gocheck suite type.
type LoggingSuite struct{}

type gocheckWriter struct {
	c *gc.C
}

var logConfig = func() string {
	if cfg := os.Getenv("JUJU_LOGGING_CONFIG"); cfg != "" {
		return cfg
	}
	return "DEBUG"
}()

func (w *gocheckWriter) Write(level loggo.Level, module, filename string, line int, timestamp time.Time, message string) {
	// Magic calldepth value...
	// TODO (frankban) Document why we are using this magic value.
	w.c.Output(3, fmt.Sprintf("%s %s %s", level, module, message))
}

func (s *LoggingSuite) SetUpSuite(c *gc.C) {
	s.setUp(c)
}

func (s *LoggingSuite) TearDownSuite(c *gc.C) {
	loggo.ResetLoggers()
	loggo.ResetWriters()
}

func (s *LoggingSuite) SetUpTest(c *gc.C) {
	s.setUp(c)
}

func (s *LoggingSuite) TearDownTest(c *gc.C) {
}

func (s *LoggingSuite) setUp(c *gc.C) {
	loggo.ResetWriters()
	loggo.ReplaceDefaultWriter(&gocheckWriter{c})
	loggo.ResetLoggers()
	err := loggo.ConfigureLoggers(logConfig)
	c.Assert(err, gc.IsNil)
}

// LoggingCleanupSuite is defined for backward compatibility.
// Do not use this suite in new tests.
type LoggingCleanupSuite struct {
	LoggingSuite
	CleanupSuite
}

func (s *LoggingCleanupSuite) SetUpSuite(c *gc.C) {
	s.CleanupSuite.SetUpSuite(c)
	s.LoggingSuite.SetUpSuite(c)
}

func (s *LoggingCleanupSuite) TearDownSuite(c *gc.C) {
	s.LoggingSuite.TearDownSuite(c)
	s.CleanupSuite.TearDownSuite(c)
}

func (s *LoggingCleanupSuite) SetUpTest(c *gc.C) {
	s.CleanupSuite.SetUpTest(c)
	s.LoggingSuite.SetUpTest(c)
}

func (s *LoggingCleanupSuite) TearDownTest(c *gc.C) {
	s.CleanupSuite.TearDownTest(c)
	s.LoggingSuite.TearDownTest(c)
}
