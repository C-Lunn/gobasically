package functions

import (
	"basicallygo/terminal"
	"basicallygo/variable"
	"errors"
	"fmt"
	"math"
)

func LOG(term *terminal.Terminal, arguments []variable.User_variable) (variable.User_variable, error) {
	if len(arguments) != 1 {
		return nil, errors.New("LOG: Invalid number of arguments")
	}
	if arguments[0].Type_of() != variable.NUMBER {
		return nil, errors.New("LOG: Invalid type")
	}
	a := arguments[0].Value().(float64)

	res := math.Log(a)

	return variable.VARTYPE_NUMBER{}.New(res), nil
}

func SQR(term *terminal.Terminal, arguments []variable.User_variable) (variable.User_variable, error) {
	if len(arguments) != 1 {
		return nil, errors.New("SQR: Invalid number of arguments")
	}
	if arguments[0].Type_of() != variable.NUMBER {
		return nil, errors.New("SQR: Invalid type")
	}
	a := arguments[0].Value().(float64)

	res := math.Sqrt(a)

	return variable.VARTYPE_NUMBER{}.New(res), nil
}

func SIN(term *terminal.Terminal, arguments []variable.User_variable) (variable.User_variable, error) {
	if len(arguments) != 1 {
		return nil, errors.New("SIN: Invalid number of arguments")
	}
	if arguments[0].Type_of() != variable.NUMBER {
		return nil, errors.New("SIN: Invalid type")
	}
	a := arguments[0].Value().(float64)

	res := math.Sin(a)

	return variable.VARTYPE_NUMBER{}.New(res), nil
}

func COS(term *terminal.Terminal, arguments []variable.User_variable) (variable.User_variable, error) {
	if len(arguments) != 1 {
		return nil, errors.New("COS: Invalid number of arguments")
	}
	if arguments[0].Type_of() != variable.NUMBER {
		return nil, errors.New("COS: Invalid type")
	}
	a := arguments[0].Value().(float64)

	res := math.Cos(a)

	return variable.VARTYPE_NUMBER{}.New(res), nil
}

func TAN(term *terminal.Terminal, arguments []variable.User_variable) (variable.User_variable, error) {
	if len(arguments) != 1 {
		return nil, errors.New("TAN: Invalid number of arguments")
	}
	if arguments[0].Type_of() != variable.NUMBER {
		return nil, errors.New("TAN: Invalid type")
	}
	a := arguments[0].Value().(float64)

	res := math.Tan(a)

	return variable.VARTYPE_NUMBER{}.New(res), nil
}

func ABS(term *terminal.Terminal, arguments []variable.User_variable) (variable.User_variable, error) {
	if len(arguments) != 1 {
		return nil, errors.New("ABS: Invalid number of arguments")
	}
	if arguments[0].Type_of() != variable.NUMBER {
		return nil, errors.New("ABS: Invalid type")
	}
	a := arguments[0].Value().(float64)

	res := math.Abs(a)

	return variable.VARTYPE_NUMBER{}.New(res), nil
}

func EXP(term *terminal.Terminal, arguments []variable.User_variable) (variable.User_variable, error) {
	if len(arguments) != 1 {
		return nil, errors.New("EXP: Invalid number of arguments")
	}
	if arguments[0].Type_of() != variable.NUMBER {
		return nil, errors.New("EXP: Invalid type")
	}
	a := arguments[0].Value().(float64)

	res := math.Exp(a)

	return variable.VARTYPE_NUMBER{}.New(res), nil
}

func VAL(term *terminal.Terminal, arguments []variable.User_variable) (variable.User_variable, error) {
	if len(arguments) != 1 {
		return nil, errors.New("VAL: Invalid number of arguments")
	}
	if arguments[0].Type_of() != variable.STRING {
		return nil, errors.New("VAL: Invalid type")
	}
	a := arguments[0].Value().(string)

	res := 0.0
	_, err := fmt.Sscanf(a, "%f", &res)
	if err != nil {
		return nil, err
	}

	return variable.VARTYPE_NUMBER{}.New(res), nil
}
