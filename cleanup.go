// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"os/exec"

	gc "gopkg.in/check.v1"
)

type CleanupFunc func(*gc.C)
type cleanupStack []CleanupFunc

// CleanupSuite adds the ability to add cleanup functions that are called
// during either test tear down or suite tear down depending on the method
// called.
type CleanupSuite struct {
	testStack  cleanupStack
	suiteStack cleanupStack
	suiteSuite *CleanupSuite
	testSuite  *CleanupSuite
}

func (s *CleanupSuite) SetUpSuite(c *gc.C) {
	s.suiteStack = nil
	s.suiteSuite = s
}

func (s *CleanupSuite) TearDownSuite(c *gc.C) {
	s.callStack(c, s.suiteStack)
	s.suiteSuite = nil
}

func (s *CleanupSuite) SetUpTest(c *gc.C) {
	s.testStack = nil
	s.testSuite = s
}

func (s *CleanupSuite) TearDownTest(c *gc.C) {
	s.callStack(c, s.testStack)
	s.testSuite = nil
}

func (s *CleanupSuite) callStack(c *gc.C, stack cleanupStack) {
	for i := len(stack) - 1; i >= 0; i-- {
		stack[i](c)
	}
}

// AddCleanup pushes the cleanup function onto the stack of functions to be
// called during TearDownTest or TearDownSuite. TearDownTest will be used if
// SetUpTest has already been called, else we will use TearDownSuite
func (s *CleanupSuite) AddCleanup(cleanup CleanupFunc) {
	if s.suiteSuite == nil {
		// This is either called before SetUpSuite or after
		// TearDownSuite. Either way, we can't really trust that we're
		// going to call Cleanup correctly.
		panic("unsafe to call AddCleanup without a Suite")
	}
	if s != s.suiteSuite {
		// If you write a test like:
		// func (s MySuite) TestFoo(c *gc.C) {
		//   s.AddCleanup(foo)
		// }
		// The AddCleanup call is unsafe because it modifes
		// s.suiteSuite but that object disappears once TestFoo
		// returns. So you have to use:
		// func (s *MySuite) TestFoo(c *gc.C) if you want the Cleanup
		// funcs.
		panic("unsafe to call AddCleanup from non pointer receiver test")
	}
	if s.testSuite == nil {
		// We either haven't called SetUpTest or we've already called
		// TearDownTest, consider this a Suite level cleanup.
		s.suiteStack = append(s.suiteStack, cleanup)
		return
	}
	s.testStack = append(s.testStack, cleanup)
}

// AddSuiteCleanup is deprecated. Just call AddCleanup and it will use the
// right lifetime for when to call the cleanup based on whether we are in a
// Test right now or not.
func (s *CleanupSuite) AddSuiteCleanup(cleanup CleanupFunc) {
	s.AddCleanup(cleanup)
}

// PatchEnvironment sets the environment variable 'name' the the value passed
// in. The old value is saved and returned to the original value at test tear
// down time using a cleanup function.
func (s *CleanupSuite) PatchEnvironment(name, value string) {
	restore := PatchEnvironment(name, value)
	s.AddCleanup(func(*gc.C) { restore() })
}

// PatchEnvPathPrepend prepends the given path to the environment $PATH and restores the
// original path on test teardown.
func (s *CleanupSuite) PatchEnvPathPrepend(dir string) {
	restore := PatchEnvPathPrepend(dir)
	s.AddCleanup(func(*gc.C) { restore() })
}

// PatchValue sets the 'dest' variable the the value passed in. The old value
// is saved and returned to the original value at test tear down time using a
// cleanup function. The value must be assignable to the element type of the
// destination.
func (s *CleanupSuite) PatchValue(dest, value interface{}) {
	restore := PatchValue(dest, value)
	s.AddCleanup(func(*gc.C) { restore() })
}

// HookCommandOutput calls the package function of the same name to mock out
// the result of a particular comand execution, and will call the restore
// function on test teardown.
func (s *CleanupSuite) HookCommandOutput(
	outputFunc *func(cmd *exec.Cmd) ([]byte, error),
	output []byte,
	err error,
) <-chan *exec.Cmd {
	result, restore := HookCommandOutput(outputFunc, output, err)
	s.AddCleanup(func(*gc.C) { restore() })
	return result
}
