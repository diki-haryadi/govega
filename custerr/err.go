package custerr

import (
	"encoding/json"
	"fmt"
)

type ErrChain struct {
	Message string
	Cause   error
	Fields  map[string]interface{}
	Type    error
}

func (err ErrChain) Error() string {
	bcoz := ""
	fields := ""
	if err.Cause != nil {
		bcoz = fmt.Sprint(" because {", err.Cause.Error(), "}")
		if len(err.Fields) > 0 {
			fields = fmt.Sprintf(" with Fields {%+v}", err.Fields)
		}
	}
	return fmt.Sprint(err.Message, bcoz, fields)
}

func Type(err error) error {
	switch err.(type) {
	case ErrChain:
		return err.(ErrChain).Type
	}
	return nil
}

func toString(m map[string]string) string {
	v, _ := json.Marshal(m)
	return string(v)
}

func (err ErrChain) SetField(key string, value string) ErrChain {
	if err.Fields == nil {
		err.Fields = map[string]interface{}{}
	}
	err.Fields[key] = value
	return err
}

type InvalidError struct {
	message string
}

func (ie *InvalidError) Error() string {
	return ie.message
}

func NewInvalidError(msg string) *InvalidError {
	return &InvalidError{message: msg}
}

func NewInvalidErrorf(msg string, args ...interface{}) *InvalidError {
	return NewInvalidError(fmt.Sprintf(msg, args...))
}
