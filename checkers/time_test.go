// Copyright 2022 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package checkers_test

import (
	"time"

	gc "gopkg.in/check.v1"

	jc "github.com/juju/testing/checkers"
)

type TimeSuite struct{}

var _ = gc.Suite(&TimeSuite{})

func (s *TimeSuite) TestBefore(c *gc.C) {
	now := time.Now()
	c.Assert(now, jc.Before, now.Add(time.Second))
	c.Assert(now, gc.Not(jc.Before), now.Add(-time.Second))

	result, msg := jc.Before.Check([]interface{}{time.Time{}}, nil)
	c.Assert(result, gc.Equals, false)
	c.Check(msg, gc.Equals, `expected 2 parameters, received 1`)

	result, msg = jc.Before.Check([]interface{}{42, time.Time{}}, nil)
	c.Assert(result, gc.Equals, false)
	c.Assert(msg, gc.Equals, `obtained param: expected type time.Time, received type int`)

	result, msg = jc.Before.Check([]interface{}{time.Time{}, "wow"}, nil)
	c.Assert(result, gc.Equals, false)
	c.Assert(msg, gc.Matches, `want param: expected type time.Time, received type string`)
}

func (s *TimeSuite) TestAfter(c *gc.C) {
	now := time.Now()
	c.Assert(now, gc.Not(jc.After), now.Add(time.Second))
	c.Assert(now, jc.After, now.Add(-time.Second))

	result, msg := jc.After.Check([]interface{}{time.Time{}}, nil)
	c.Assert(result, gc.Equals, false)
	c.Check(msg, gc.Equals, `expected 2 parameters, received 1`)

	result, msg = jc.After.Check([]interface{}{42, time.Time{}}, nil)
	c.Assert(result, gc.Equals, false)
	c.Assert(msg, gc.Equals, `obtained param: expected type time.Time, received type int`)

	result, msg = jc.After.Check([]interface{}{time.Time{}, "wow"}, nil)
	c.Assert(result, gc.Equals, false)
	c.Assert(msg, gc.Matches, `want param: expected type time.Time, received type string`)
}

func (s *TimeSuite) TestAlmost(c *gc.C) {
	now := time.Now()
	c.Assert(now, gc.Not(jc.Almost), now.Add(1001*time.Millisecond))
	c.Assert(now, jc.Almost, now.Add(-time.Second))
	c.Assert(now, jc.Almost, now.Add(time.Second))

	result, msg := jc.Almost.Check([]interface{}{time.Time{}}, nil)
	c.Assert(result, gc.Equals, false)
	c.Check(msg, gc.Equals, `expected 2 parameters, received 1`)

	result, msg = jc.Almost.Check([]interface{}{42, time.Time{}}, nil)
	c.Assert(result, gc.Equals, false)
	c.Assert(msg, gc.Equals, `obtained param: expected type time.Time, received type int`)

	result, msg = jc.Almost.Check([]interface{}{time.Time{}, "wow"}, nil)
	c.Assert(result, gc.Equals, false)
	c.Assert(msg, gc.Matches, `want param: expected type time.Time, received type string`)
}
