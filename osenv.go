// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"os"
	"runtime"
	"strings"

	gc "launchpad.net/gocheck"
)

// OsEnvSuite isolates the tests from the underlaying system environment.
// Environment variables are reset in SetUpTest and restored in TearDownTest.
type OsEnvSuite struct {
	oldEnvironment map[string]string
}

// windowsVariables is a whitelist of windows environment variables
// that will be retained if found. Some of these variables are needed
// by standard go packages (such as os.TempDir()), as well as powershell
var windowsVariables = []string{
	"ALLUSERSPROFILE",
	"APPDATA",
	"CommonProgramFiles",
	"CommonProgramFiles(x86)",
	"CommonProgramW6432",
	"COMPUTERNAME",
	"ComSpec",
	"FP_NO_HOST_CHECK",
	"HOMEDRIVE",
	"HOMEPATH",
	"LOCALAPPDATA",
	"LOGONSERVER",
	"NUMBER_OF_PROCESSORS",
	"OS",
	"Path",
	"PATHEXT",
	"PROCESSOR_ARCHITECTURE",
	"PROCESSOR_IDENTIFIER",
	"PROCESSOR_LEVEL",
	"PROCESSOR_REVISION",
	"ProgramData",
	"ProgramFiles",
	"ProgramFiles(x86)",
	"ProgramW6432",
	"PROMPT",
	"PSModulePath",
	"PUBLIC",
	"SESSIONNAME",
	"SystemDrive",
	"SystemRoot",
	"TEMP",
	"TMP",
	"USERDOMAIN",
	"USERDOMAIN_ROAMINGPROFILE",
	"USERNAME",
	"USERPROFILE",
	"windir",
}

func (s *OsEnvSuite) setEnviron() {
	var envList []string
	switch runtime.GOOS {
	case "windows":
		envList = windowsVariables
	default:
		envList = []string{}
	}
	for _, envVar := range envList {
		if value, ok := s.oldEnvironment[envVar]; ok {
			os.Setenv(envVar, value)
		}
	}
}

// osDependendClearenv will clear the environment, and based on platform, will repopulate
// with whitelisted values previously saved in s.oldEnvironment
func (s *OsEnvSuite) osDependendClearenv() {
	os.Clearenv()
	// Currently, this will only do something if we are running on windows
	s.setEnviron()
}

func (s *OsEnvSuite) SetUpSuite(c *gc.C) {
	s.oldEnvironment = make(map[string]string)
	for _, envvar := range os.Environ() {
		parts := strings.SplitN(envvar, "=", 2)
		s.oldEnvironment[parts[0]] = parts[1]
	}
	s.osDependendClearenv()
}

func (s *OsEnvSuite) TearDownSuite(c *gc.C) {
	os.Clearenv()
	for name, value := range s.oldEnvironment {
		os.Setenv(name, value)
	}
}

func (s *OsEnvSuite) SetUpTest(c *gc.C) {
	s.osDependendClearenv()
}

func (s *OsEnvSuite) TearDownTest(c *gc.C) {
}
