package util

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Lookup lookup value from interface
func Lookup(name string, context ...interface{}) (interface{}, bool) {
	val, ok := lookup(name, context...)
	switch v := val.(type) {
	case float64:
		if v == float64(int64(v)) {
			return int64(v), ok
		}
		return v, ok
	default:
		return val, ok
	}
}

// Match check if value is match with given context and path
func Match(name string, context interface{}, value interface{}) bool {
	return match(name, context, value)
}

func match(name string, context interface{}, value interface{}) bool {

	if !strings.Contains(name, "_") {

		val, ok := lookup(name, context)
		if ok {
			k, err := regexp.Match(fmt.Sprintf("%v", value), []byte(fmt.Sprintf("%v", val)))
			if err != nil {
				return false
			}
			return k
		}

		if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", val) {
			return true
		}

		if val == nil {
			switch value.(type) {
			case string:
				val = ""
				return fmt.Sprintf("%v", value) == val
			default:
				return val == value
			}
		}
		return false
	}

	if name == "_" {
		rv := reflect.ValueOf(context)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			l := rv.Len()
			for i := 0; i < l; i++ {
				item := rv.Index(i)
				if item.IsValid() {
					ok, err := regexp.Match(fmt.Sprintf("%v", value), []byte(fmt.Sprintf("%v", item.Interface())))
					if err != nil {
						return false
					}
					return ok
				}
			}
		}
	}

	parts := strings.SplitN(name, ".", 2)
	if parts[0] == "_" {
		rv := reflect.ValueOf(context)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			l := rv.Len()
			for i := 0; i < l; i++ {
				item := rv.Index(i)
				if item.IsValid() {
					if match(parts[1], item.Interface(), value) {
						return true
					}
				}
			}
		}
	}

	if val, ok := lookup(parts[0], context); ok {
		return match(parts[1], val, value)
	}

	return false
}

// This function is taken from https://github.com/alexkappa/mustache
// Since this function is not exposed as public function, hence we copy the source directly
// The lookup function searches for a property that matches name within the
// context chain. We first start from the first item in the context chain which
// is the most likely to have the value we're looking for. If not found, we'll
// move up the chain and repeat.
func lookup(name string, context ...interface{}) (interface{}, bool) {
	// If the dot notation was used we split the word in two and perform two
	// consecutive lookups. If the first one fails we return no value and a
	// negative truth. Taken from github.com/hoisie/mustache.

	if len(context) == 0 || context == nil {
		return nil, false
	}

	if name != "." && strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		if value, ok := lookup(parts[0], context...); ok {
			return lookup(parts[1], value)
		}
		return nil, false
	}
	// Iterate over the context chain and try to match the name to a value.
	for _, c := range context {
		// Reflect on the value of the current context.
		reflectValue := reflect.ValueOf(c)
		// If the name is ".", we should return the whole context as-is.
		if name == "." {
			return c, truth(reflectValue)
		}
		switch reflectValue.Kind() {
		case reflect.Ptr:
			return lookup(name, reflectValue.Elem().Interface())
		// If the current context is a map, we'll look for a key in that map
		// that matches the name.
		case reflect.Map:
			item := reflectValue.MapIndex(reflect.ValueOf(name))
			if item.IsValid() {
				return item.Interface(), truth(item)
			}
		// If the current context is a struct, we'll look for a property in that
		// struct that matches the name. In the near future I'd like to add
		// support for matching struct names to tags so we can use lower_case
		// names in our templates which makes it more mustache like.
		case reflect.Struct:
			field := reflectValue.FieldByName(name)
			if field.IsValid() {
				return field.Interface(), truth(field)
			}

			method := reflectValue.MethodByName(name)
			if method.IsValid() && method.Type().NumIn() == 1 {
				out := method.Call(nil)[0]
				return out.Interface(), truth(out)
			}

			if !field.IsValid() && !method.IsValid() {
				return nil, false
			}

		case reflect.Slice, reflect.Array:
			idx, err := strconv.Atoi(name)
			if err != nil {
				return nil, false
			}
			if idx >= reflectValue.Len() {
				return nil, false
			}
			item := reflectValue.Index(idx)
			if item.IsValid() {
				return item.Interface(), truth(item)
			}
		}
		// If by this point no value was matched, we'll move up a step in the
		// chain and try to match a value there.
	}
	// We've exhausted the whole context chain and found nothing. Return a nil
	// value and a negative truth.
	return nil, false
}

// The truth function will tell us if r is a truthy value or not. This is
// important for sections as they will render their content based on the output
// of this function.
//
// Zero values are considered falsy. For example an empty string, the integer 0
// and so on are all considered falsy.
func truth(r reflect.Value) bool {
out:
	switch r.Kind() {
	case reflect.Array, reflect.Slice:
		return r.Len() > 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return r.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return r.Uint() > 0
	case reflect.Float32, reflect.Float64:
		return r.Float() != 0
	case reflect.String:
		return r.String() != ""
	case reflect.Bool:
		return r.Bool()
	case reflect.Ptr, reflect.Interface:
		r = r.Elem()
		goto out
	default:
		// check if value IsValid first, otherwise IsZero may panic
		if !r.IsValid() {
			return false
		}

		// additional check to ensure not a zero struct
		if r.IsZero() {
			return false
		}

		return r.Interface() != nil
	}
}

func Assert(name string, context interface{}, value interface{}, op string) bool {

	if !strings.Contains(name, "._") && !strings.Contains(name, "_.") && name != "_" {

		val, ok := lookup(name, context)
		if ok {
			k, err := CompareValue(val, value, op)
			if err != nil {
				return false
			}
			return k
		}

		if ok, _ := CompareValue(val, value, op); ok {
			return true
		}

		if val == nil {
			switch value.(type) {
			case string:
				val = ""
				ok, _ := CompareValue(val, value, op)
				return ok
				//return fmt.Sprintf("%v", value) == val
			default:
				ok, _ := CompareValue(val, value, op)
				return ok
				//return val == value
			}
		}
		return false
	}

	if name == "_" {
		rv := reflect.ValueOf(context)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			l := rv.Len()
			for i := 0; i < l; i++ {
				item := rv.Index(i)
				if item.IsValid() {
					ok, err := CompareValue(item.Interface(), value, op)
					if err != nil {
						return false
					}
					return ok
				}
			}
		}
	}

	parts := strings.SplitN(name, ".", 2)
	if parts[0] == "_" {
		rv := reflect.ValueOf(context)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			l := rv.Len()
			for i := 0; i < l; i++ {
				item := rv.Index(i)
				if item.IsValid() {
					if Assert(parts[1], item.Interface(), value, op) {
						return true
					}
				}
			}
		}
	}

	if val, ok := lookup(parts[0], context); ok {
		return Assert(parts[1], val, value, op)
	}

	return false
}
