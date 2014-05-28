// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	gc "launchpad.net/gocheck"

	"github.com/juju/loggo"

	"github.com/juju/testing"
)

var logger = loggo.GetLogger("test")

type logSuite struct {
	testing.LoggingSuite
}

var _ = gc.Suite(&logSuite{})

func (s *logSuite) SetUpSuite(c *gc.C) {
	s.LoggingSuite.SetUpSuite(c)
	logger.SetLogLevel(loggo.INFO)
	logger.Infof("testing-SetUpSuite")
	c.Assert(c.GetTestLog(), gc.Matches, ".*INFO test testing-SetUpSuite\n")
}

func (s *logSuite) TearDownSuite(c *gc.C) {
	// Unfortunately there's no way of testing that the
	// log output is printed, as the logger is printing
	// a previously set up *gc.C. We print a message
	// anyway so that we can manually verify it.
	logger.Infof("testing-TearDownSuite")
	s.LoggingSuite.TearDownSuite(c)
	logger.Infof("YOU SHOULD NOT SEE THIS")
}

func (s *logSuite) SetUpTest(c *gc.C) {
	s.LoggingSuite.SetUpTest(c)
	// The SetUpTest resets the logging levels.
	logger.SetLogLevel(loggo.INFO)
	logger.Infof("testing-SetUpTest")
	c.Assert(c.GetTestLog(), gc.Matches, ".*INFO test testing-SetUpTest\n")
}

func (s *logSuite) TearDownTest(c *gc.C) {
	// The same applies here as to TearDownSuite.
	logger.Infof("testing-TearDownTest")
}

func (s *logSuite) TestLog(c *gc.C) {
	logger.Infof("testing-Test")
	c.Assert(c.GetTestLog(), gc.Matches,
		".*INFO test testing-SetUpTest\n"+
			".*INFO test testing-Test\n",
	)
}
