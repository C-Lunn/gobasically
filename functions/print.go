package functions

import (
	"basicallygo/terminal"
	"basicallygo/variable"
)

func PRINT(term terminal.Terminal, arguments []variable.User_variable) (variable.User_variable, error) {
	out := ""
	for _, argument := range arguments {
		out += argument.To_string() + " "
	}
	term.Printline(out)
	return nil, nil
}
