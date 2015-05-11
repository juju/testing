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
var _ = gc.Suite(&tagMatchingSuite{})

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

func (tagParsingSuite) TestParseTagsDuplicateTags(c *gc.C) {
	tags := testing.ParseTags("spam,ham,eggs,eggs,ham")

	c.Check(tags, jc.DeepEquals, []string{"spam", "ham", "eggs", "eggs", "ham"})
}

func (tagParsingSuite) TestParseTagsEmpty(c *gc.C) {
	tags := testing.ParseTags()

	c.Check(tags, gc.HasLen, 0)
}

func (tagParsingSuite) TestParseTagsMultipleStrings(c *gc.C) {
	tags := testing.ParseTags("spam,ham,eggs", "foo,bar")

	c.Check(tags, jc.DeepEquals, []string{"spam", "ham", "eggs", "foo", "bar"})
}

func (tagParsingSuite) TestParseTagsSkipMissing(c *gc.C) {
	tags := testing.ParseTags(",spam,,ham,eggs,")

	c.Check(tags, jc.DeepEquals, []string{"spam", "ham", "eggs"})
}

type tagMatchingSuite struct{}

func (s *tagMatchingSuite) SetUpTest(c *gc.C) {
	s.setParsed("spam", "eggs")
}

func (tagMatchingSuite) setParsed(tags ...string) {
	*testing.ParsedTags = tags
}

func (s tagMatchingSuite) TestCheckTagTryMultiple(c *gc.C) {
	matched := testing.CheckTag("ham", "eggs")

	c.Check(matched, jc.IsTrue)
}

func (s tagMatchingSuite) TestCheckTagTrySingle(c *gc.C) {
	matched := testing.CheckTag("spam")

	c.Check(matched, jc.IsTrue)
}

func (s tagMatchingSuite) TestCheckTagNoMatch(c *gc.C) {
	matched := testing.CheckTag("ham")

	c.Check(matched, jc.IsFalse)
}

func (s tagMatchingSuite) TestCheckTagNoneParsed(c *gc.C) {
	s.setParsed()

	matched := testing.CheckTag("spam")

	c.Check(matched, jc.IsFalse)
}

func (s tagMatchingSuite) TestCheckTagEmpty(c *gc.C) {
	matched := testing.CheckTag()

	c.Check(matched, jc.IsFalse)
}

func (s tagMatchingSuite) TestCheckTagExcluded(c *gc.C) {
	s.setParsed("spam", "-eggs")

	matched := testing.CheckTag("eggs")

	c.Check(matched, jc.IsFalse)
}

func (s tagMatchingSuite) TestCheckTagNotExcluded(c *gc.C) {
	s.setParsed("spam", "-eggs")

	matched := testing.CheckTag("spam")

	c.Check(matched, jc.IsTrue)
}

func (s tagMatchingSuite) TestCheckTagAlmostExcluded(c *gc.C) {
	s.setParsed("spam", "-eggs")

	matched := testing.CheckTag("spam", "eggs")

	c.Check(matched, jc.IsTrue)
}

func (s tagMatchingSuite) TestMatchTagTryMultipleMatchOne(c *gc.C) {
	matched := testing.MatchTag("ham", "eggs")

	c.Check(matched, gc.Equals, "eggs")
}

func (s tagMatchingSuite) TestMatchTagTryMultipleMatchMultiple(c *gc.C) {
	matched := testing.MatchTag("spam", "ham", "eggs")

	c.Check(matched, gc.Equals, "spam")
}

func (s tagMatchingSuite) TestMatchTagTrySingle(c *gc.C) {
	matched := testing.MatchTag("spam")

	c.Check(matched, gc.Equals, "spam")
}

func (s tagMatchingSuite) TestMatchTagNoMatch(c *gc.C) {
	matched := testing.MatchTag("ham")

	c.Check(matched, gc.Equals, "")
}

func (s tagMatchingSuite) TestMatchTagNoneParsed(c *gc.C) {
	s.setParsed()

	matched := testing.MatchTag("spam")

	c.Check(matched, gc.Equals, "")
}

func (s tagMatchingSuite) TestMatchTagEmpty(c *gc.C) {
	matched := testing.MatchTag()

	c.Check(matched, gc.Equals, "")
}

func (s tagMatchingSuite) TestMatchTagExcluded(c *gc.C) {
	s.setParsed("spam", "-eggs")

	matched := testing.MatchTag("eggs")

	c.Check(matched, gc.Equals, "")
}

func (s tagMatchingSuite) TestMatchTagNotExcluded(c *gc.C) {
	s.setParsed("spam", "-eggs")

	matched := testing.MatchTag("spam")

	c.Check(matched, gc.Equals, "spam")
}

func (s tagMatchingSuite) TestMatchTagAlmostExcluded(c *gc.C) {
	s.setParsed("spam", "-eggs")

	matched := testing.MatchTag("spam", "eggs")

	c.Check(matched, gc.Equals, "spam")
}
