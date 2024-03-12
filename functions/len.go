package functions

import (
	"basicallygo/terminal"
	"basicallygo/variable"
	"errors"
)

func LEN(term *terminal.Terminal, arguments []variable.User_variable) (variable.User_variable, error) {
	if len(arguments) != 1 {
		return nil, errors.New("LEN: ONLY ONE ARGUMENT")
	}
	switch arguments[0].Type_of() {
	case variable.STRING:
		strlen := len(arguments[0].Value().(string))
		return (&variable.VARTYPE_NUMBER{}).New(float64(strlen)), nil
	case variable.ARRAY:
		arrlen := len(arguments[0].Value().([]variable.User_variable))
		return (&variable.VARTYPE_NUMBER{}).New(float64(arrlen)), nil
	default:
		return nil, errors.New("LEN: INVALID TYPE")
	}

}
