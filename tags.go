// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	gc "gopkg.in/check.v1"
)

// The following code provides a mechanism by which tests may be tagged.
// When thus tagged, a test will only run when one of the tags specified
// at the commandline matches (and doesn't match an excluded tag). The
// commandline usage looks like this:
//
//  go test . --tags small,medium,-functional
//
// This would result in running only the tests that are tagged as small
// or medium or that are not tagged as functional. The first match wins,
// so a medium test would match even if it is tagged as functional.
//
// Use multiple flags for logical-AND:
//
//  go test . --tags medium --tags -functional
//
// This would match all medium tests that are not also marked as
// functional.
//
// As a convenience, there is a dedicated commandline flag for tests
// that run very quickly as a sanity check of the code:
//
//  go test . --smoke
//
// The following helpers are used to tag tests:
//
//  RegisterPackageTagged - use in place of gc.TestingT
//  SuiteTagged - use in place of gc.Suite
//  RequireTag - use in tests, SetUpTest, or SetUpSuite
//
// Note that test tagging is opt-in, so untagged tests will always run.

// These are generally useful tags to use in tests.
const (
	TagSmall      = "small"      // Runs quickly (smoke tests).
	TagLarge      = "large"      // Does not run quickly.
	TagFunctional = "functional" // Does not use test doubles for low-level.
	// TODO(ericsnow) Add other tags? For example:
	//  - default: test runs when no other tags are specified
	//  - external: test uses external resources (e.g. filesystem)
	//  - cloud: test interacts with a cloud provider's API
	//  - vm: test runs in a local VM (e.g. kvm) for isolation
)

var defaultTags = []string{
	TagSmall,
	TagLarge,
	TagFunctional,
}

var smokeTags = []string{
	TagSmall,
}

var (
	rawTags tagsValue
)

func init() {
	flag.Var(&rawTags, "tags", "Tagged tests to run.")
	flag.BoolVar(&rawTags.smoke, "smoke", false, "Run the basic set of fast tests.")
}

type tagsValue struct {
	raw    []string
	smoke  bool
	parsed [][]string
}

// Set implements flag.Value.
func (v *tagsValue) Set(s string) error {
	v.raw = append(v.raw, s)
	return nil
}

// String implements flag.Value.
func (v *tagsValue) String() string {
	return strings.Join(v.raw, ",")
}

func (v *tagsValue) parse() [][]string {
	if v.parsed == nil {
		v.parsed = handleCommandline(v.raw, v.smoke)
	}
	return v.parsed
}

func handleCommandline(rawList []string, smoke bool) [][]string {
	var parsedTags [][]string
	for _, raw := range rawList {
		parsed := parseTags(raw)
		if len(parsed) == 0 {
			continue
		}
		if smoke {
			parsed = append(parsed, smokeTags...)
		}
		parsedTags = append(parsedTags, parsed)
	}
	if len(parsedTags) == 0 {
		if smoke {
			parsedTags = append(parsedTags, smokeTags)
		} else {
			parsedTags = append(parsedTags, defaultTags)
		}
	}
	// TODO(ericsnow) support implied tags (e.g. VM -> Large)?
	return parsedTags
}

func parseTags(rawList ...string) []string {
	var tags []string
	for _, raw := range rawList {
		for _, entry := range strings.Split(raw, ",") {
			if len(entry) == 0 {
				continue
			}
			tag := entry
			tags = append(tags, tag)
		}
	}
	return tags
}

// CheckTag determines whether or not any of the given tags were passed
// in at the commandline. Matches on "excluded" tags automatically fail.
func CheckTag(tags ...string) bool {
	for _, parsed := range rawTags.parse() {
		if MatchTag(parsed, tags...) == "" {
			return false
		}
	}
	return true
}

// MatchTag returns the first provided tag that matches a required tag,
// unless the required tag is an exclusion (starts with "-"). In that
// case the check automatically fails. This is equivalent to OR'ing the
// parsed tags.
func MatchTag(requiredTags []string, tags ...string) string {
	for _, required := range requiredTags {
		for _, tag := range tags {
			if required == "" {
				continue
			}
			if required[0] == '-' && tag == required[1:] {
				return ""
			}
			if tag == required {
				return tag
			}
		}
	}
	return ""
}

// RegisterPackageTagged registers the package for testing if any of the
// given tags were passed in at the commandline.
func RegisterPackageTagged(t *testing.T, tags ...string) {
	if CheckTag(tags...) {
		gc.TestingT(t)
	}
}

// SuiteTagged registers the suite with the test runner if any of the
// given tags were passed in at the commandline.
func SuiteTagged(suite interface{}, tags ...string) {
	if CheckTag(tags...) {
		gc.Suite(suite)
	}
}

// RequireTag causes a test or suite to skip if none of the given tags
// were passed in at the commandline.
func RequireTag(c *gc.C, tags ...string) {
	if !CheckTag(tags...) {
		c.Skip(fmt.Sprintf("skipping due to no matching tags (%v)", tags))
	}
}
