// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
)

var _ = gc.Suite(&tagsCommandlineSuite{})
var _ = gc.Suite(&tagParsingSuite{})

type tagsCommandlineSuite struct{}

func (tagsCommandlineSuite) parsedTags() []string {
	return *testing.ParsedTags
}

func (s tagsCommandlineSuite) TestHandleCommandlineMultipleTags(c *gc.C) {
	testing.HandleCommandline("spam,ham,eggs", false)
	tags := s.parsedTags()

	c.Check(tags, jc.DeepEquals, []string{"spam", "ham", "eggs"})
}

func (s tagsCommandlineSuite) TestHandleCommandlineSingleTag(c *gc.C) {
	testing.HandleCommandline("spam", false)
	tags := s.parsedTags()

	c.Check(tags, jc.DeepEquals, []string{"spam"})
}

func (s tagsCommandlineSuite) TestHandleCommandlineSmokeOnly(c *gc.C) {
	testing.HandleCommandline("", true)
	tags := s.parsedTags()

	c.Check(tags, jc.DeepEquals, []string{testing.TagSmoke})
}

func (s tagsCommandlineSuite) TestHandleCommandlineSmokeAdded(c *gc.C) {
	testing.HandleCommandline("spam", true)
	tags := s.parsedTags()

	c.Check(tags, jc.DeepEquals, []string{"spam", testing.TagSmoke})
}

func (s tagsCommandlineSuite) TestHandleCommandlineDefault(c *gc.C) {
	testing.HandleCommandline("", false)
	tags := s.parsedTags()

	c.Check(tags, jc.DeepEquals, []string{testing.TagBase})
}

func (s tagsCommandlineSuite) TestHandleCommandlineExcludedOnly(c *gc.C) {
	testing.HandleCommandline("-spam", false)
	tags := s.parsedTags()

	c.Check(tags, jc.DeepEquals, []string{"-spam"})
}

func (s tagsCommandlineSuite) TestHandleCommandlineExcludedMixed(c *gc.C) {
	testing.HandleCommandline("spam,-eggs", false)
	tags := s.parsedTags()

	c.Check(tags, jc.DeepEquals, []string{"spam", "-eggs"})
}

type tagParsingSuite struct{}

func (tagParsingSuite) TestParseTagsMultipleTags(c *gc.C) {
	tags := testing.ParseTags("spam,ham,eggs")

	c.Check(tags, jc.DeepEquals, []string{"spam", "ham", "eggs"})
}

func (tagParsingSuite) TestParseTagsSingleTag(c *gc.C) {
	tags := testing.ParseTags("spam")

	c.Check(tags, jc.DeepEquals, []string{"spam"})
}

func (tagParsingSuite) TestParseTagsEmpty(c *gc.C) {
	tags := testing.ParseTags()

	c.Check(tags, gc.HasLen, 0)
}

func (tagParsingSuite) TestParseTagsMultipleStrings(c *gc.C) {
	tags := testing.ParseTags("spam,ham,eggs", "foo,bar")

	c.Check(tags, jc.DeepEquals, []string{"spam", "ham", "eggs", "foo", "bar"})
}
