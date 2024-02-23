package functions

import (
	"basicallygo/terminal"
	"basicallygo/variable"
)

type Function func(terminal.Terminal, []variable.User_variable) (variable.User_variable, error)

type Standard_functions map[string]Function

var Std_fcns Standard_functions = Standard_functions{
	"PRINT": PRINT,
	"LOG":   LOG,
	"SQR":   SQR,
	"SIN":   SIN,
	"COS":   COS,
	"ABS":   ABS,
	"EXP":   EXP,
	"VAL":   VAL,
}
