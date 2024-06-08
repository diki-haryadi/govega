package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceContainsAny(t *testing.T) {
	type (
		args struct {
			source []string
			search []string
		}

		wants struct {
			result bool
		}
	)

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "return_false_when_source_doesnt_contains_any_value_from_search",
			args: args{
				source: []string{"foo", "bar"},
				search: []string{"baz"},
			},
			wants: wants{
				result: false,
			},
		},
		{
			name: "return_true_when_source_contains_one_value_from_search",
			args: args{
				source: []string{"foo", "bar"},
				search: []string{"baz", "bar"},
			},
			wants: wants{
				result: true,
			},
		},
		{
			name: "return_true_when_source_contains_multiple_value_from_search",
			args: args{
				source: []string{"foo", "bar", "lorem", "ipsum", "dolor", "sit", "amet"},
				search: []string{"baz", "lorem", "john", "foo", "cat", "ipsum", "dog"},
			},
			wants: wants{
				result: true,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wants.result, SliceContainsAny(tc.args.source, tc.args.search))
		})
	}

}
