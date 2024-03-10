package functions

import (
	"basicallygo/variable"
	"errors"
	"math"
)

type Operator_fcn func(left variable.User_variable, right variable.User_variable) (variable.User_variable, error)

type Operators map[string]Operator_fcn

func ADD(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		return (&variable.VARTYPE_NUMBER{}).New(left.Value().(float64) + right.Value().(float64)), nil
	} else {
		err := errors.New("ADD: Invalid types")
		return nil, err
	}
}

func SUBTRACT(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		return (&variable.VARTYPE_NUMBER{}).New(left.Value().(float64) - right.Value().(float64)), nil
	} else {
		err := errors.New("SUBTRACT: Invalid types")
		return nil, err
	}
}

func MULTIPLY(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		return (&variable.VARTYPE_NUMBER{}).New(left.Value().(float64) * right.Value().(float64)), nil
	} else {
		err := errors.New("MULTIPLY: Invalid types")
		return nil, err
	}
}

func DIVIDE(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		return (&variable.VARTYPE_NUMBER{}).New(left.Value().(float64) / right.Value().(float64)), nil
	} else {
		err := errors.New("DIVIDE: Invalid types")
		return nil, err
	}
}

func MODULO(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		return (&variable.VARTYPE_NUMBER{}).New(float64(int(left.Value().(float64)) % int(right.Value().(float64)))), nil
	} else {
		err := errors.New("MODULO: Invalid types")
		return nil, err
	}
}

func POW(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		return (&variable.VARTYPE_NUMBER{}).New(
			math.Pow(left.Value().(float64), right.Value().(float64)),
		), nil
	} else {
		err := errors.New("EXP: Invalid types")
		return nil, err
	}
}

func GT(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		if left.Value().(float64) > right.Value().(float64) {
			return (&variable.VARTYPE_NUMBER{}).New(1), nil
		}
		return (&variable.VARTYPE_NUMBER{}).New(0), nil
	} else {
		err := errors.New("GT: Invalid types")
		return nil, err
	}
}

func GTE(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		if left.Value().(float64) >= right.Value().(float64) {
			return (&variable.VARTYPE_NUMBER{}).New(1), nil
		}
		return (&variable.VARTYPE_NUMBER{}).New(0), nil
	} else {
		err := errors.New("GTE: Invalid types")
		return nil, err
	}
}

func LTE(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		if left.Value().(float64) <= right.Value().(float64) {
			return (&variable.VARTYPE_NUMBER{}).New(1), nil
		}
		return (&variable.VARTYPE_NUMBER{}).New(0), nil
	} else {
		err := errors.New("LTE: Invalid types")
		return nil, err
	}
}

func LT(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		if left.Value().(float64) < right.Value().(float64) {
			return (&variable.VARTYPE_NUMBER{}).New(1), nil
		}
		return (&variable.VARTYPE_NUMBER{}).New(0), nil
	} else {
		err := errors.New("LT: Invalid types")
		return nil, err
	}
}

func EQ(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		if left.Value().(float64) == right.Value().(float64) {
			return (&variable.VARTYPE_NUMBER{}).New(1), nil
		}
		return (&variable.VARTYPE_NUMBER{}).New(0), nil
	} else if left.Type_of() == variable.STRING && right.Type_of() == variable.STRING {
		if left.Value().(string) == right.Value().(string) {
			return (&variable.VARTYPE_NUMBER{}).New(1), nil
		}
		return (&variable.VARTYPE_NUMBER{}).New(0), nil
	} else {
		err := errors.New("EQ: Invalid types")
		return nil, err
	}
}

func NEQ(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	res, err := EQ(left, right)
	if err == nil {
		if res.Value().(float64) == 0 {
			return (&variable.VARTYPE_NUMBER{}).New(1), nil
		}
		return (&variable.VARTYPE_NUMBER{}).New(0), nil
	}
	return nil, err
}

func AND(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		if left.Value().(float64) != 0 && right.Value().(float64) != 0 {
			return (&variable.VARTYPE_NUMBER{}).New(1), nil
		}
		return (&variable.VARTYPE_NUMBER{}).New(0), nil
	} else {
		err := errors.New("AND: Invalid types")
		return nil, err
	}
}

func OR(left variable.User_variable, right variable.User_variable) (variable.User_variable, error) {
	if left.Type_of() == variable.NUMBER && right.Type_of() == variable.NUMBER {
		if left.Value().(float64) != 0 || right.Value().(float64) != 0 {
			return (&variable.VARTYPE_NUMBER{}).New(1), nil
		}
		return (&variable.VARTYPE_NUMBER{}).New(0), nil
	} else {
		err := errors.New("OR: Invalid types")
		return nil, err
	}
}

var Ops = Operators{
	"+":   ADD,
	"-":   SUBTRACT,
	"*":   MULTIPLY,
	"/":   DIVIDE,
	"%":   MODULO,
	"^":   POW,
	">":   GT,
	">=":  GTE,
	"<":   LT,
	"<=":  LTE,
	"==":  EQ,
	"<>":  NEQ,
	"AND": AND,
	"OR":  OR,
}
