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
// As a convienced, there is a dedicated commandline flag for tests that
// run very quickly as a sanity check of the code:
//
//  go test . --smoke

// These are generally useful tags to use in tests.
const (
	TagBase       = "base"
	TagSmall      = "small"
	TagMedium     = "medium"
	TagLarge      = "large"
	TagSmoke      = "smoke"      // Fast sanity-check.
	TagFunctional = "functional" // Does not use test doubles for low-level.
	// TODO(ericsnow) Add other tags? For example:
	//  - external: test uses external resources (e.g. filesystem)
	//  - cloud: test interacts with a cloud provider's API
	//  - vm: test runs in a local VM (e.g. kvm) for isolation
)

var defaultTags = []string{
	TagBase,
}

var (
	parsedTags []string
)

func init() {
	smoke := flag.Bool("smoke", false, "Run the basic set of fast tests.")
	raw := flag.String("tags", "", "Tagged tests to run.")
	flag.Parse()

	handleCommandline(*raw, *smoke)
}

func handleCommandline(raw string, smoke bool) {
	parsedTags = parseTags(raw)
	if smoke {
		parsedTags = append(parsedTags, TagSmoke)
	}
	if len(parsedTags) == 0 {
		parsedTags = defaultTags
	}
	// TODO(ericsnow) support implied tags (e.g. VM -> Large)?
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
	return MatchTag(tags...) != ""
}

// MatchTag returns the first provided tag that matches the ones passed
// in at the commandline, unless the match is an exclusion (starts with
// "-").  In that case the check automatically fails. This is equivalent
// to OR'ing the parsed tags.
func MatchTag(tags ...string) string {
	for _, tag := range tags {
		for _, parsedTag := range parsedTags {
			if parsedTag[0] == '-' && tag == parsedTag[1:] {
				return ""
			}
			if tag == parsedTag {
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

// SkipTag causes a test or suite to skip if any of the given tags were
// passed in at the commandline.
func SkipTag(c *gc.C, tags ...string) {
	matched := MatchTag(tags...)
	if matched != "" {
		c.Skip(fmt.Sprintf("skipping due to %q tag", matched))
	}
}
