// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"os"
	"strings"

	gc "launchpad.net/gocheck"
)

// OsEnvSuite isolates the tests from the underlaying system environment.
// Environment variables are reset in SetUpTest and restored in TearDownTest.
type OsEnvSuite struct {
	oldEnvironment map[string]string
}

func (s *OsEnvSuite) SetUpSuite(c *gc.C) {
	s.oldEnvironment = make(map[string]string)
	for _, envvar := range os.Environ() {
		parts := strings.SplitN(envvar, "=", 2)
		s.oldEnvironment[parts[0]] = parts[1]
	}
	os.Clearenv()
}

func (s *OsEnvSuite) TearDownSuite(c *gc.C) {
	os.Clearenv()
	for name, value := range s.oldEnvironment {
		os.Setenv(name, value)
	}
}

func (s *OsEnvSuite) SetUpTest(c *gc.C) {
	os.Clearenv()
}

func (s *OsEnvSuite) TearDownTest(c *gc.C) {
}
