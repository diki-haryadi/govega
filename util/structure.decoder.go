package util

import (
	"errors"
	"reflect"
	"strconv"
	"time"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DecodeJSON(input, output interface{}) error {
	return decodeVal(input, output, "json", customTagDecoder)
	//return decode(input, customTagDecoder(output, "json"))
}

func decodeVal(input, output interface{}, tag string, fn func(interface{}, string) *mapstructure.DecoderConfig) error {

	if IsMapStringInterface(output) && IsStructOrPointerOf(input) {

		s := structs.New(input)
		s.TagName = tag

		if reflect.ValueOf(output).Kind() == reflect.Ptr {
			if reflect.ValueOf(output).Elem().IsNil() {
				return errors.New("map should not nil")
			}
			s.FillMap(*output.(*map[string]interface{}))
			return nil
		}

		if reflect.ValueOf(output).IsNil() {
			return errors.New("map should not nil")
		}

		s.FillMap(output.(map[string]interface{}))

		return nil
	}

	return decode(input, fn(output, tag))
}

func decode(input interface{}, config *mapstructure.DecoderConfig) error {
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

func customTagDecoder(output interface{}, tag string) *mapstructure.DecoderConfig {
	return structDecoder(output, tag, mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		mapstructure.StringToTimeHookFunc(time.RFC3339),
		ToTimeHookFunc(""),
		FromTimeHookFunc(),
		//FromTimeHook(),
	))
}

func structDecoder(output interface{}, tag string, hook mapstructure.DecodeHookFunc) *mapstructure.DecoderConfig {
	c := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook:       hook,
	}

	if tag != "" {
		c.TagName = tag
	}

	return c
}

func ToTimeHookFunc(format string) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		if f == reflect.TypeOf(primitive.DateTime(0)) {
			return (data.(primitive.DateTime)).Time(), nil
		}

		switch f.Kind() {
		case reflect.String:
			if format != "" {
				return time.Parse(format, data.(string))
			}
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

func FromTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		if f != reflect.TypeOf(time.Time{}) && f != reflect.PtrTo(reflect.TypeOf(time.Time{})) {
			return data, nil
		}

		tf, ok := data.(time.Time)

		if !ok {
			tmp, ok := data.(*time.Time)
			if !ok {
				return nil, errors.New("error converting time")
			}
			tf = *tmp
		}

		if t == reflect.TypeOf(time.Time{}) {
			return tf, nil
		}

		switch t.Kind() {
		case reflect.String:
			return tf.Format(time.RFC3339), nil
		case reflect.Float64:
			return float64(tf.UnixNano()), nil
		case reflect.Int64:
			return tf.UnixNano(), nil
		default:
			return tf.Format(time.RFC3339), nil
		}
		// Convert it by parsing
	}
}

func DecodeString(str string) (interface{}, error) {
	if v, err := strconv.ParseBool(str); err == nil {
		return v, nil
	}

	if v, err := strconv.ParseFloat(str, 64); err == nil {
		if v == float64(int64(v)) {
			return int64(v), nil
		}
		return v, nil
	}

	return str, nil
}
