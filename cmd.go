// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/juju/utils"

	gc "gopkg.in/check.v1"
)

var HookChannelSize = 10

// HookCommandOutput intercepts CommandOutput to a function that passes the
// actual command and it's output back via a channel, and returns the error
// passed into this function.  It also returns a cleanup function so you can
// restore the original function
func HookCommandOutput(
	outputFunc *func(cmd *exec.Cmd) ([]byte, error), output []byte, err error) (<-chan *exec.Cmd, func()) {

	cmdChan := make(chan *exec.Cmd, HookChannelSize)
	origCommandOutput := *outputFunc
	cleanup := func() {
		close(cmdChan)
		*outputFunc = origCommandOutput
	}
	*outputFunc = func(cmd *exec.Cmd) ([]byte, error) {
		cmdChan <- cmd
		return output, err
	}
	return cmdChan, cleanup
}

const (
	// EchoQuotedArgs is a simple bash script that prints out the
	// basename of the command followed by the args as quoted strings.
	// If a ; separated list of exit codes is provided in $name.exitcodes
	// then it will return them in turn over multiple calls. If
	// $name.exitcodes does not exist, or the list runs out, return 0.
	EchoQuotedArgsUnix = `#!/bin/bash --norc
name=` + "`basename $0`" + `
argfile="$name.out"
exitcodesfile="$name.exitcodes"
printf "%s" $name | tee -a $argfile
for arg in "$@"; do
  printf " '%s'" "$arg" | tee -a $argfile
done
printf "\n" | tee -a $argfile
if [ -f $exitcodesfile ]
then
	exitcodes=$(cat $exitcodesfile)
	arr=(${exitcodes/;/ })
	echo ${arr[1]} | tee $exitcodesfile
	exit ${arr[0]}
fi
`
	EchoQuotedArgsWindows = `@echo off

setlocal enabledelayedexpansion
set list=%0
set argCount=0
for %%x in (%*) do (
   set /A argCount+=1
   set "argVec[!argCount!]=%%~x"
)
for /L %%i in (1,1,%argCount%) do set list=!list! '!argVec[%%i]!'

IF exist %0.exitcodes (
    FOR /F "tokens=1* delims=;" %%i IN (%0.exitcodes) DO (
        set exitcode=%%i
        IF NOT [%%j]==[] (
            echo %%j > %0.exitcodes
        ) ELSE (
            del %0.exitcodes
        )
    )
)

echo %list%>> %0.out
exit /B %exitcode%
`
)

// EnvironmentPatcher is an interface that requires just one method:
// PatchEnvironment.
type EnvironmentPatcher interface {
	PatchEnvironment(name, value string)
}

// PatchExecutable creates an executable called 'execName' in a new test
// directory and that directory is added to the path.
func PatchExecutable(c *gc.C, patcher CleanupPatcher, execName, script string, exitCodes ...int) {
	dir := c.MkDir()
	patcher.PatchEnvironment("PATH", joinPathLists(dir, os.Getenv("PATH")))
	var filename string
	switch runtime.GOOS {
	case "windows":
		filename = filepath.Join(dir, execName+".bat")
	default:
		filename = filepath.Join(dir, execName)
	}
	os.Remove(filename + ".out")
	err := ioutil.WriteFile(filename, []byte(script), 0755)
	c.Assert(err, gc.IsNil)

	if len(exitCodes) > 0 {
		filename = execName + ".exitcodes"
		codes := make([]string, len(exitCodes))
		for i, code := range exitCodes {
			codes[i] = strconv.Itoa(code)
		}
		s := strings.Join(codes, ";") + ";"
		err = ioutil.WriteFile(filename, []byte(s), 0644)
		c.Assert(err, gc.IsNil)
		patcher.AddCleanup(func(*gc.C) {
			os.Remove(filename)
		})
	}
}

type CleanupPatcher interface {
	PatchEnvironment(name, value string)
	AddCleanup(cleanup CleanupFunc)
}

// PatchExecutableThrowError is needed to test cases in which we expect exit
// codes from executables called from the system path
func PatchExecutableThrowError(c *gc.C, patcher CleanupPatcher, execName string, exitCode int) {
	switch runtime.GOOS {
	case "windows":
		script := fmt.Sprintf(`@echo off
		                       setlocal enabledelayedexpansion
                               echo failing
                               exit /b %d
                               REM see %ERRORLEVEL% for last exit code like $? on linux
                               `, exitCode)
		PatchExecutable(c, patcher, execName, script)
	default:
		script := fmt.Sprintf(`#!/bin/bash --norc
                               echo failing
                               exit %d
                               `, exitCode)
		PatchExecutable(c, patcher, execName, script)
	}
	patcher.AddCleanup(func(*gc.C) {
		os.Remove(execName + ".out")
	})

}

// PatchExecutableAsEchoArgs creates an executable called 'execName' in a new
// test directory and that directory is added to the path. The content of the
// script is 'EchoQuotedArgs', and the args file is removed using a cleanup
// function.
func PatchExecutableAsEchoArgs(c *gc.C, patcher CleanupPatcher, execName string, exitCodes ...int) {
	switch runtime.GOOS {
	case "windows":
		PatchExecutable(c, patcher, execName, EchoQuotedArgsWindows, exitCodes...)
	default:
		PatchExecutable(c, patcher, execName, EchoQuotedArgsUnix, exitCodes...)
	}
	patcher.AddCleanup(func(*gc.C) {
		os.Remove(execName + ".out")
		os.Remove(execName + ".exitcodes")
	})
}

// AssertEchoArgs is used to check the args from an execution of a command
// that has been patchec using PatchExecutable containing EchoQuotedArgs.
func AssertEchoArgs(c *gc.C, execName string, args ...string) {
	// Read in entire argument log file
	content, err := ioutil.ReadFile(execName + ".out")
	c.Assert(err, gc.IsNil)
	lines := strings.Split(string(content), "\n")

	// Create expected output string
	expected := execName
	for _, arg := range args {
		expected = fmt.Sprintf("%s %s", expected, utils.ShQuote(arg))
	}

	// Check that the expected and the first line of actual output are the same
	actual := strings.TrimSuffix(lines[0], "\r")

	c.Assert(actual, gc.Equals, expected)

	// Write out the remaining lines for the next check
	content = []byte(strings.Join(lines[1:], "\n"))
	err = ioutil.WriteFile(execName+".out", content, 0644) // or just call this filename somewhere, once.
}
