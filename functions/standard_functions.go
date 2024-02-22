package functions

import (
	"basicallygo/terminal"
	"basicallygo/variable"
)

type Standard_function func(terminal.Terminal, []variable.User_variable) variable.User_variable

type Standard_functions map[string]Standard_function

var Std_fcns Standard_functions = Standard_functions{
	"PRINT": PRINT,
}
