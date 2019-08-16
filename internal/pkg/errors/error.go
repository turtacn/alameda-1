package errors

import (
	"strconv"
	"fmt"
)

func NewError(id int, args ...interface{}) *InternalError {
	internalError := InternalError{}
	internalError.ErrorID = id
	internalError.Args = make([]interface{}, 0)

	for _, v := range args {
		switch v.(type) {
		case string:
			internalError.Args = append(internalError.Args, v.(string))
		case int:
			internalError.Args = append(internalError.Args, strconv.Itoa(v.(int)))
		}
	}

	return &internalError
}

type InternalError struct {
	ErrorID int
	Args    []interface{}
}

func (e *InternalError) Error() string {
	return fmt.Sprintf(reasonMap[e.ErrorID], e.Args...)
}

func (e *InternalError) Append(args ...interface{}) {
	for _, v := range args {
		switch v.(type) {
		case string:
			e.Args = append(e.Args, v.(string))
		case int:
			e.Args = append(e.Args, strconv.Itoa(v.(int)))
		}
	}
}
