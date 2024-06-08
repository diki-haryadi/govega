package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCommaSeparatedList(t *testing.T) {
	type (
		args struct {
			elems []string
		}

		wants struct {
			result string
		}
	)

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "return_comma_separated_list_string",
			args: args{
				[]string{"foo", "bar", "baz"},
			},
			wants: wants{
				result: "foo,bar,baz",
			},
		},
		{
			name: "return_trimmed_comma_seperated_list_string",
			args: args{
				[]string{"foo ", " bar", " baz "},
			},
			wants: wants{
				result: "foo,bar,baz",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wants.result, ToCommaSeparatedList(tc.args.elems))
		})
	}
}

func TestFromCommaSeparatedList(t *testing.T) {
	type (
		args struct {
			input string
		}

		wants struct {
			result []string
		}
	)

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "return_array_list",
			args: args{
				input: "foo,bar,baz",
			},
			wants: wants{
				result: []string{"foo", "bar", "baz"},
			},
		},
		{
			name: "return_trimmed_array_list",
			args: args{
				input: "foo , bar, baz ",
			},
			wants: wants{
				result: []string{"foo", "bar", "baz"},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wants.result, FromCommaSeparatedList(tc.args.input))
		})
	}
}
