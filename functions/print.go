package functions

import (
	"basicallygo/terminal"
	"basicallygo/variable"
)

func PRINT(term terminal.Terminal, arguments []variable.User_variable) *variable.User_variable {
	for _, argument := range arguments {
		term.Printline(argument.To_string())
	}
	return nil
}
