package strutil

import (
	"strconv"
)

const (
	bit64  = 64
	base10 = 10

	float64ToStringDefaultPrecision = 6
	float64ToStringDefualtFormat    = byte('f')
)

type (
	float64StringOpts struct {
		format    byte
		precision int
	}

	// Float64StringOption option for converting floa64 to string
	Float64StringOption func(f *float64StringOpts)
)

// ParseInt64 parse string to int64
func ParseInt64(in string) (int64, error) {
	return strconv.ParseInt(in, base10, bit64)
}

// ParseFloat64 transfrom string into float64
func ParseFloat64(in string) (float64, error) {
	return strconv.ParseFloat(in, bit64)
}

// Int64ToString convert int64 value to string
func Int64ToString(in int64) string {
	return strconv.FormatInt(in, base10)
}

// Float64ToString convert float64 value to string
// default format is 'f' (-ddd.dddd, no exponent) with 6 precision
func Float64ToString(in float64, opts ...Float64StringOption) string {
	floatOpts := float64StringOpts{
		format:    float64ToStringDefualtFormat,
		precision: float64ToStringDefaultPrecision,
	}

	for _, o := range opts {
		o(&floatOpts)
	}

	return strconv.FormatFloat(in, floatOpts.format, floatOpts.precision, bit64)
}

// WithFloat64StringFormat define the format of string represention of the float64
// The format value is one of
// 'b' (-ddddp±ddd, a binary exponent),
// 'e' (-d.dddde±dd, a decimal exponent),
// 'E' (-d.ddddE±dd, a decimal exponent),
// 'f' (-ddd.dddd, no exponent),
// 'g' ('e' for large exponents, 'f' otherwise),
// 'G' ('E' for large exponents, 'f' otherwise),
// 'x' (-0xd.ddddp±ddd, a hexadecimal fraction and binary exponent), or
// 'X' (-0Xd.ddddP±ddd, a hexadecimal fraction and binary exponent).
func WithFloat64StringFormat(format byte) Float64StringOption {
	return func(f *float64StringOpts) {
		f.format = format
	}
}

// WithFloat64StringPrecision define the precision of the string value represention of the float64
func WithFloat64StringPrecision(precision int) Float64StringOption {
	return func(f *float64StringOpts) {
		f.precision = precision
	}
}
