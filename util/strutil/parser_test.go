package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInt64(t *testing.T) {
	type (
		args struct {
			in string
		}

		wants struct {
			result    int64
			errString string
		}
	)

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "return_error_when_string_value_is_not_integer",
			args: args{
				in: "invalid int string",
			},
			wants: wants{
				errString: "strconv.ParseInt: parsing \"invalid int string\": invalid syntax",
			},
		},
		{
			name: "return_int_value_of_string_when_string_is_integer",
			args: args{
				in: "10",
			},
			wants: wants{
				result: 10,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseInt64(tc.args.in)

			switch tc.wants.errString {
			case "":
				require.NoError(t, err)
				assert.Equal(t, tc.wants.result, result)
			default:
				assert.EqualError(t, err, tc.wants.errString)
			}
		})
	}

}

func TestParseFloat64(t *testing.T) {
	type (
		args struct {
			in string
		}

		wants struct {
			result    float64
			errString string
		}
	)

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "return_error_when_string_value_is_not_float",
			args: args{
				in: "invalid float string",
			},
			wants: wants{
				errString: "strconv.ParseFloat: parsing \"invalid float string\": invalid syntax",
			},
		},
		{
			name: "return_int_value_of_string_when_string_is_integer",
			args: args{
				in: "10",
			},
			wants: wants{
				result: 10,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseFloat64(tc.args.in)

			switch tc.wants.errString {
			case "":
				require.NoError(t, err)
				assert.Equal(t, tc.wants.result, result)
			default:
				assert.EqualError(t, err, tc.wants.errString)
			}
		})
	}
}

func TestInt64ToString(t *testing.T) {
	type (
		args struct {
			in int64
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
			name: "return_10_string",
			args: args{
				in: 10,
			},
			wants: wants{
				result: "10",
			},
		},
		{
			name: "return_24_string",
			args: args{
				in: 24,
			},
			wants: wants{
				result: "24",
			},
		},
		{
			name: "return_234791_string",
			args: args{
				in: 234791,
			},
			wants: wants{
				result: "234791",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wants.result, Int64ToString(tc.args.in))
		})
	}
}

func TestFloat64ToString(t *testing.T) {
	type (
		args struct {
			in   float64
			opts []Float64StringOption
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
			name: "default_format_return_string_with_6_precision",
			args: args{
				in: 23.234,
			},
			wants: wants{
				result: "23.234000",
			},
		},
		{
			name: "return_string_representation_with_3_precision",
			args: args{
				in: 23.234,
				opts: []Float64StringOption{
					WithFloat64StringPrecision(3),
				},
			},
			wants: wants{
				result: "23.234",
			},
		},
		{
			name: "return_string_representation_with_decimal_exponent",
			args: args{
				in: 23.234,
				opts: []Float64StringOption{
					WithFloat64StringFormat('E'),
				},
			},
			wants: wants{
				result: "2.323400E+01",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wants.result, Float64ToString(tc.args.in, tc.args.opts...))
		})
	}
}
