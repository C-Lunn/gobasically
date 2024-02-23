package functions

import (
	"basicallygo/variable"
	"errors"
)

type Operator_fcn func(left variable.User_variable, right variable.User_variable) (variable.User_variable, error)

type Operators map[string]Operator_fcn

func ADD(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		return variable.VARTYPE_NUMBER{}.New(left.Value().(float64) + right.Value().(float64)), nil
	} else {
		err := errors.New("ADD: Invalid types")
		return nil, err
	}
}

var Ops = Operators{
	"+": ADD,
}
