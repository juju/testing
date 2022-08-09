// Copyright 2022 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package checkers

import (
	"fmt"
	"math"
	"reflect"
	"time"

	gc "gopkg.in/check.v1"
)

type timeCompareChecker struct {
	*gc.CheckerInfo
	compareFunc func(time.Time, time.Time) bool
}

// After checks whether the obtained time.Time is After the want time.Time.
var After gc.Checker = &timeCompareChecker{
	CheckerInfo: &gc.CheckerInfo{Name: "After", Params: []string{"obtained", "want"}},
	compareFunc: func(t1, t2 time.Time) bool {
		return t1.After(t2)
	},
}

// Before checks whether the obtained time.Time is Before the want time.Time.
var Before gc.Checker = &timeCompareChecker{
	CheckerInfo: &gc.CheckerInfo{Name: "Before", Params: []string{"obtained", "want"}},
	compareFunc: func(t1, t2 time.Time) bool {
		return t1.Before(t2)
	},
}

// Almost checks whether the obtained time.Time is within 1s of the want time.Time.
var Almost gc.Checker = &timeCompareChecker{
	CheckerInfo: &gc.CheckerInfo{Name: "Almost", Params: []string{"obtained", "want"}},
	compareFunc: func(t1, t2 time.Time) bool {
		return math.Abs(t1.Sub(t2).Seconds()) <= 1.0
	},
}

func (checker *timeCompareChecker) Check(params []interface{}, names []string) (result bool, error string) {
	if len(params) != 2 {
		return false, fmt.Sprintf("expected 2 parameters, received %d", len(params))
	}
	t1, ok := params[0].(time.Time)
	if !ok {
		return false, fmt.Sprintf("obtained param: expected type time.Time, received type %s", reflect.ValueOf(params[0]).Type())
	}
	t2, ok := params[1].(time.Time)
	if !ok {
		return false, fmt.Sprintf("want param: expected type time.Time, received type %s", reflect.ValueOf(params[1]).Type())
	}
	return checker.compareFunc(t1, t2), ""
}
