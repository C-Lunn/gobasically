package context

import (
	"basicallygo/functions"
	"basicallygo/terminal"
	"basicallygo/variable"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type code_line struct {
	line_number int
	line        string
}

type Context struct {
	terminal        terminal.Terminal
	input_buffer    string
	program         []code_line
	program_counter int // index in slice of current line
	line_number     int // current line number

	user_variables map[string]variable.User_variable
}

func New(
	terminal terminal.Terminal,
) *Context {
	context := new(Context)
	context.terminal = terminal
	context.user_variables = make(map[string]variable.User_variable)
	return context
}

func (context *Context) Set_input_buffer(input_buffer string) {
	context.input_buffer = input_buffer
}

func (context *Context) parse_program() {
	context.program = make([]code_line, 0)
	split_program := strings.Split(context.input_buffer, "\n")
	// If the string starts with 1-5 numbers followed by a space, we can assume it is a line number, else continue the previous line
	regex_test := regexp.MustCompile(`^\d{1,5} `)
	for _, line := range split_program {
		if regex_test.MatchString(line) {
			// This is a new line
			// Split the line into line number and code
			line_number_string := regex_test.FindString(line)
			// Trim the space off the end
			line_number_string = strings.TrimSuffix(line_number_string, " ")
			line_number, _ := strconv.Atoi(line_number_string)
			line = strings.TrimPrefix(line, line_number_string)
			// remove last \n
			line = strings.TrimSuffix(line, "\n")
			context.program = append(context.program, code_line{line_number: line_number, line: line})

		} else {
			// This is a continuation of the previous line
			// remove last \n
			line = strings.TrimSuffix(line, "\n")
			// add to last line
			context.program[len(context.program)-1].line += line

		}
	}
	// Now, sort the slice according to line number
	sort.SliceStable(context.program, func(i, j int) bool {
		return context.program[i].line_number < context.program[j].line_number
	})

}

func (context *Context) tokenise_line(line string, ch chan Token) {
	str_offset := 0
	for {
		if str_offset >= len(line) {
			break
		}
		// consume whitespace
		for {
			if line[str_offset] == ' ' {
				str_offset++
			} else {
				break
			}
		}
		// read next char
		next_char := line[str_offset]
		// Check if the token is in the LUT
		if token_type, ok := Token_type_lookup[string(next_char)]; ok {
			ch <- Token{token_type: token_type, value: string(next_char)}
			str_offset++
			continue
		}
		// Check if the token is a number
		if is_digit(next_char) {
			number, new_offset := read_float(line, str_offset)
			ch <- Token{token_type: NUMBER, value: number}
			str_offset = new_offset
			continue
		}
		// Check if the token is a string
		if next_char == '"' {
			str, new_offset := read_rest_of_string(line, str_offset+1)
			ch <- Token{token_type: STRING, value: str}
			str_offset = new_offset + 1 // last "
			continue
		}
		// Check if it's an operator
		if next_char == '+' || next_char == '-' || next_char == '*' || next_char == '/' || next_char == '%' || next_char == '^' || next_char == '=' || next_char == '!' {
			ch <- Token{token_type: Token_type_lookup[string(next_char)], value: string(next_char)}
			str_offset++
			continue
		}
		// Get the next set of alpha characters
		word, new_offset := read_until_next_non_alpha(line, str_offset)
		// Check if the token is a keyword
		if Is_keyword(word) {
			ch <- Token{token_type: KEYWORD, value: word}
			str_offset = new_offset
			continue
		}
		// Check if the token is a variable
		if _, ok := context.user_variables[word]; ok {
			ch <- Token{token_type: USER_VAR, value: word}
			str_offset = new_offset
			continue
		}
		// Check if it's a std function
		if _, ok := functions.Std_fcns[word]; ok {
			ch <- Token{token_type: STD_FCN, value: word}
			str_offset = new_offset
			continue
		}
		// If it's none of the above, it's an error
		panic("Unrecognised token: " + word)
	}
	close(ch)

}

func (context *Context) next_token(ch chan Token) (Token, bool) {
	tkn, open := <-ch
	return tkn, open
}

/**
 * Execute a statement
 * Returns when a delimiter is seen (i.e. semicolon, comma, or line end)
 * Idea is that it's called recursively
 */
func (context *Context) execute_statement(ch chan Token) (variable.User_variable, *Token) {
	var state Exec_state
	for {
		t, alive := context.next_token(ch)
		if !alive {
			if state != nil {
				return state.(Exec_state_user_var).Variable, &t
			}
			return nil, nil
		}
		switch t.token_type {
		case STD_FCN:
			state = Exec_state_Fcn{Fcn: functions.Std_fcns[t.value]}
			fcn_state := state.(Exec_state_Fcn)
			// slice for the arguments
			fcn_state.Args = make([]variable.User_variable, 0)
			// Evaluate each argument until ; or line end
			var stopped_token *Token
			for {
				user_var, stopped_token := context.execute_statement(ch)
				if user_var != nil {
					fcn_state.Args = append(state.(Exec_state_Fcn).Args, user_var)
				}
				if stopped_token != nil {
					if stopped_token.token_type == DELIMITER_SEMICOLON {
						break
					}
				} else {
					break
				}
			}
			// Execute the function
			result := fcn_state.Fcn(context.terminal, fcn_state.Args)
			// There's got to be a more idiomatic way to do this. If I change it to *User_variable, strings and numbers don't work any more
			// I'm probably missing something obvious
			if result != nil {
				return *result, stopped_token
			}
			return nil, stopped_token
		case USER_VAR:
			// If state hasn't been initialised, then it's the first (or maybe only) user var
			if state == nil {
				state = Exec_state_user_var{Variable: context.user_variables[t.value]}
			}
			continue
		case NUMBER:
			// Get the number
			if state == nil {
				state = Exec_state_user_var{Variable: variable.VARTYPE_NUMBER{}.From_string(t.value)}
			}
			continue
		case STRING:
			// Get the string
			if state == nil {
				state = Exec_state_user_var{Variable: variable.VARTYPE_STRING{}.New(t.value)}
			}
			continue
		case DELIMITER_COMMA:
		case DELIMITER_SEMICOLON:
			if state != nil {
				return state.(Exec_state_user_var).Variable, &t
			}
			break
		case OPERATOR_ADD:
			state = Exec_state_operator{}.From_user_var(state.(Exec_state_user_var), t.value)
			// evaluate for the rhs
			user_var, stopped_token := context.execute_statement(ch)
			// execute the operator
			op_state := state.(Exec_state_operator)
			result, _ := op_state.Operator_func(op_state.Left, user_var)
			return result, stopped_token
		}

	}
}

func (context *Context) get_idx_of_line_no(line_no int) int {
	for idx, line := range context.program {
		if line.line_number == line_no {
			return idx
		}
	}
	return -1
}

func (context *Context) Run() {
	context.parse_program()
	context.program_counter = 0
	for context.program_counter < len(context.program) {
		ch := make(chan Token)
		go context.tokenise_line(context.program[context.program_counter].line, ch)
		_, _ = context.execute_statement(ch)
		context.program_counter++
	}
}
