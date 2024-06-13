// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package checkers_test

import (
	"reflect"
	"testing"

	gc "gopkg.in/check.v1"

	jc "github.com/juju/testing/checkers"
)

type listEqualsSuite struct{}

var _ = gc.Suite(&listEqualsSuite{})

type testCase struct {
	description  string
	list1, list2 any
	equal        bool
	error        string
}

var testCases = []testCase{{
	description: "both not slices",
	list1:       struct{}{},
	list2:       map[string]string{},
	error:       "expected value is not a slice",
}, {
	description: "obtained is not a slice",
	list1:       1,
	list2:       []string{},
	error:       "obtained value is not a slice",
}, {
	description: "expected is not a slice",
	list1:       []string{},
	list2:       "foobar",
	error:       "expected value is not a slice",
}, {
	description: "same contents but different element type",
	list1:       []string{"A", "B", "C", "DEF"},
	list2:       []any{"A", "B", "C", "DEF"},
	error:       "element types are not equal",
}, {
	description: "different type in last position",
	list1:       []any{"A", "B", "C", "DEF"},
	list2:       []any{"A", "B", "C", 321},
	equal:       false,
	error: `difference:
    - at index 3: obtained element "DEF", expected 321`,
}, {
	description: "incomparable element type",
	list1:       [][]string{{"A"}},
	list2:       [][]string{{"A"}},
	error:       "element type is not comparable",
}, {
	description: "incomparable values are fine",
	list1:       []any{"A"},
	list2:       []any{[]string{"A"}},
	error: `difference:
    - at index 0: obtained element "A", expected \[\]string\{"A"\}`,
}, {
	description: "elements missing at start",
	list1:       []int{5, 6},
	list2:       []int{3, 4, 5, 6},
	equal:       false,
	error: `difference:
    - at index 0: missing element 3
    - at index 1: missing element 4`,
}, {
	description: "elements added at start",
	list1:       []int{1, 2, 3, 4, 5, 6},
	list2:       []int{3, 4, 5, 6},
	equal:       false,
	error: `difference:
    - at index 0: unexpected element 1
    - at index 0: unexpected element 2`,
}, {
	description: "elements missing at end",
	list1:       []int{3, 4},
	list2:       []int{3, 4, 5, 6},
	equal:       false,
	error: `difference:
    - at index 2: missing element 5
    - at index 3: missing element 6`,
}, {
	description: "elements added at end",
	list1:       []int{3, 4, 5, 6, 7, 8},
	list2:       []int{3, 4, 5, 6},
	equal:       false,
	error: `difference:
    - at index 4: unexpected element 7
    - at index 4: unexpected element 8`,
}, {
	description: "basic test",
	list1:       []int{0, 2, 62, 4, 43, 5, 7, 104, 9, 56, 10},
	list2:       []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	equal:       false,
	error: `difference:
    - at index 1: missing element 1
    - at index 3: obtained element 62, expected 3
    - at index 5: unexpected element 43
    - at index 6: missing element 6
    - at index 8: obtained element 104, expected 8
    - at index 10: unexpected element 56
    - at index 11: missing element 11`,
}, {
	description: "replaced elements",
	list1:       []string{"A", "Z", "C", "Y", "E", "X", "G", "W", "I"},
	list2:       []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"},
	equal:       false,
	error: `difference:
    - at index 1: obtained element "Z", expected "B"
    - at index 3: obtained element "Y", expected "D"
    - at index 5: obtained element "X", expected "F"
    - at index 7: obtained element "W", expected "H"`,
}}

func init() {
	// Add a test case with two super long but equal arrays. In this case, we
	// should find equality first in O(n), so it shouldn't be too slow.

	N := 10000
	superLongSlice := make([]int, N)
	for i := 0; i < N; i++ {
		superLongSlice[i] = i
	}

	testCases = append(testCases, testCase{
		description: "super long slice",
		list1:       superLongSlice,
		list2:       superLongSlice,
		equal:       true,
	})
}

func (s *listEqualsSuite) Test(c *gc.C) {
	for _, test := range testCases {
		c.Log(test.description)
		res, err := jc.ListEquals.Check([]any{test.list1, test.list2}, nil)
		c.Check(res, gc.Equals, test.equal)
		c.Check(err, gc.Matches, test.error)
	}
}

func BenchmarkListEquals(b *testing.B) {
	for _, test := range testCases {
		b.Run(test.description, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				jc.ListEquals.Check([]any{test.list1, test.list2}, nil)
			}
		})
	}
}

func FuzzListEquals(f *testing.F) {
	f.Fuzz(func(t *testing.T, list1, list2 []byte) {
		eq, errMsg := jc.ListEquals.Check([]any{list1, list1}, nil)
		if eq == false || errMsg != "" {
			t.Errorf("should ListEquals itself: %v", list1)
		}

		eq, errMsg = jc.ListEquals.Check([]any{list1, list2}, nil)
		if eq != (reflect.DeepEqual(list1, list2)) {
			t.Errorf(`ListEquals returned incorrect value for
list1: %v
list2: %v`, list1, list2)
		}
	})
}
