package context

import (
	"basicallygo/functions"
	"basicallygo/terminal"
	"basicallygo/variable"
	"errors"
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
	terminal          *terminal.Terminal
	input_buffer      string
	program           []code_line
	program_counter   int // index in slice of current line
	run_state_stack   []Run_state
	user_variables    map[string]variable.User_variable
	tokeniser_channel chan Token
	exec_state_stack  []Exec_state
}

func New(
	terminal *terminal.Terminal,
) *Context {
	context := new(Context)
	context.terminal = terminal
	context.user_variables = make(map[string]variable.User_variable)
	context.tokeniser_channel = make(chan Token)
	context.exec_state_stack = make([]Exec_state, 0)
	return context
}

func (context *Context) ess_push(state Exec_state) {
	context.exec_state_stack = append(context.exec_state_stack, state)
}

func (context *Context) ess_pop() Exec_state {
	if len(context.exec_state_stack) == 0 {
		return nil
	}
	state := context.exec_state_stack[len(context.exec_state_stack)-1]
	context.exec_state_stack = context.exec_state_stack[:len(context.exec_state_stack)-1]
	return state
}

func (context *Context) ess_replace_top(state Exec_state) {
	context.exec_state_stack[len(context.exec_state_stack)-1] = state
}

func (context *Context) ess_peek(which int) Exec_state {
	return context.exec_state_stack[len(context.exec_state_stack)-1-which]
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
		if line == "\n" || line == "" {
			continue
		}
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

func (context *Context) parse_line(line string) error {
	regex_test := regexp.MustCompile(`^\d{1,5} `)
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
		if context.program == nil {
			context.program = make([]code_line, 0)
		}
		if context.get_idx_of_line_no(line_number) != -1 {
			//replace
			context.program[context.get_idx_of_line_no(line_number)] = code_line{line_number: line_number, line: line}
		} else {
			context.program = append(context.program, code_line{line_number: line_number, line: line})
		}
	} else {
		if context.program == nil || len(context.program) == 0 {
			return errors.New("no line number")
		}
		// This is a continuation of the previous line
		// remove last \n
		line = strings.TrimSuffix(line, "\n")
		// add to last line
		context.program[len(context.program)-1].line += line
	}
	return nil
}

func (context *Context) Accept_line(line string) {
	err := context.parse_line(line)
	if err != nil {
		context.terminal.Printline(err.Error())
	}
}

func (context *Context) List() {
	for _, line := range context.program {
		context.terminal.Printline(strconv.Itoa(line.line_number) + line.line)
	}

}

func (context *Context) tokenise_line(line string) {
	ch := context.tokeniser_channel
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
		// Workaround for <>, OR, AND, ==
		if (next_char == 'O' || next_char == '<' || next_char == '=') && str_offset+1 < len(line) {
			if line[str_offset+1] == 'R' || line[str_offset+1] == '>' || line[str_offset+1] == '=' {
				ch <- Token{token_type: OPERATOR, value: string(line[str_offset : str_offset+2])}
				str_offset += 2
				continue
			}

		} else if next_char == 'A' && str_offset+2 < len(line) {
			if line[str_offset+1] == 'N' && line[str_offset+2] == 'D' {
				ch <- Token{token_type: OPERATOR, value: "AND"}
				str_offset += 3
				continue
			}
		}
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
		// Assume that it's a variable that has yet to be assigned
		ch <- Token{token_type: USER_VAR_UNASSIGNED, value: word}
		str_offset = new_offset
	}
	ch <- Token{token_type: DELIMITER_EOL, value: ""}

}

func (context *Context) next_token_line_aware() (*Token, bool) {
	tkn := <-context.tokeniser_channel
	if tkn.token_type == DELIMITER_EOL {
		//Move to the next line, tokenise, and keep going
		context.program_counter++
		if context.program_counter >= len(context.program) {
			return nil, false
		}
		context.tokeniser_channel = make(chan Token)
		go context.tokenise_line(context.program[context.program_counter].line)
		return &tkn, true
	}
	return &tkn, true
}

func (context *Context) peek_top_run_state_stack() RUN_STATE {
	if len(context.run_state_stack) == 0 {
		return RUN_STATE_NORMAL
	}
	return context.run_state_stack[len(context.run_state_stack)-1].Get_state()
}

/**
 * Execute a statement
 * Returns when a delimiter is seen (i.e. semicolon, comma, or line end)
 * Idea is that it's called recursively
 */
func (context *Context) execute_statement() (variable.User_variable, *Token, error) {
	var state Exec_state
	get_next_token := true
	var t *Token
	var alive bool
	for {
		if get_next_token {
			t, alive = context.next_token_line_aware()
		} else {
			get_next_token = true
		}
		if !alive || t.token_type == DELIMITER_EOL {
			if state != nil {
				return state.(Exec_state_user_var).Variable, t, nil
			}
			if t == nil {
				return nil, nil, nil
			}
			if t.token_type == DELIMITER_EOL {
				return nil, t, nil
			}
			return nil, nil, nil
		}
		switch t.token_type {
		case STD_FCN:
			if state != nil {
				e := errors.New("STD_FCN: Expected a delimiter")
				return nil, nil, e
			}
			state = Exec_state_Fcn{Fcn: functions.Std_fcns[t.value]}
			context.ess_push(state)
			defer context.ess_pop()
			fcn_state := state.(Exec_state_Fcn)
			// slice for the arguments
			fcn_state.Args = make([]variable.User_variable, 0)
			// Evaluate each argument until ; or line end
			var stopped_token *Token
			for {
				user_var, stopped_tkn, err := context.execute_statement()
				if err != nil {
					return nil, nil, err
				}
				if user_var != nil {
					fcn_state.Args = append(fcn_state.Args, user_var)
				}
				if stopped_tkn != nil {
					if stopped_tkn.token_type == DELIMITER_SEMICOLON ||
						(context.peek_top_run_state_stack() == RUN_STATE_IF && stopped_tkn.token_type == KEYWORD && stopped_tkn.value == "END") ||
						(context.peek_top_run_state_stack() == RUN_STATE_IF && stopped_tkn.token_type == KEYWORD && stopped_tkn.value == "ELSE") ||
						(context.peek_top_run_state_stack() == RUN_STATE_IF && stopped_tkn.token_type == KEYWORD && stopped_tkn.value == "THEN") ||
						stopped_tkn.token_type == DELIMITER_EOL {
						stopped_token = stopped_tkn
						break
					} else if stopped_tkn.token_type == DELIMITER_COMMA {
						continue
					}
				} else {
					break
				}
			}
			// Execute the function
			result, _ := fcn_state.Fcn(context.terminal, fcn_state.Args)
			// There's got to be a more idiomatic way to do this. If I change it to *User_variable, strings and numbers don't work any more
			// I'm probably missing something obvious
			if result != nil {
				return result, stopped_token, nil
			}
			return nil, stopped_token, nil
		case USER_VAR:
			// If state hasn't been initialised, then it's the first (or maybe only) user var
			if state == nil {
				state = Exec_state_user_var{Variable: context.user_variables[t.value], Name: t.value}
			}
			continue
		case USER_VAR_UNASSIGNED:
			// If state hasn't been initialised, then it's the first (or maybe only) user var
			if state == nil {
				state = Exec_state_unassigned{Variable: t.value}
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
		case DELIMITER_COMMA, DELIMITER_SEMICOLON:
			if state != nil {
				return state.(Exec_state_user_var).Variable, t, nil
			}
		case OPERATOR_MATHEMATICAL:
			if state == nil && t.value == "-" {
				// minus
				// Check if the next token is a number
				next_tkn, _ := context.next_token_line_aware()
				if next_tkn.token_type == NUMBER {
					state = Exec_state_user_var{Variable: variable.VARTYPE_NUMBER{}.From_string("-" + next_tkn.value)}
					continue
				}
			}
			if state == nil {
				e := errors.New("OPERATOR_MATHEMATICAL: Expected a variable")
				return nil, nil, e
			}
			state = Exec_state_operator_mathematical{}.From_user_var(state.(Exec_state_user_var), t.value)
			context.ess_push(state)
			defer context.ess_pop()
			user_var, stopped_token, err := context.execute_statement()
			if err != nil {
				return nil, nil, err
			}
			// execute the operator
			op_state := state.(Exec_state_operator_mathematical)
			result, _ := op_state.Operator_func(op_state.Left, user_var)
			state = Exec_state_user_var{Variable: result}
			context.ess_replace_top(state)
			if stopped_token != nil {
				if stopped_token.token_type == OPERATOR {
					get_next_token = false
					t = stopped_token
					continue
				} else {
					return result, stopped_token, nil
				}
			}
		case OPERATOR:
			if t.value == "=" {
				// Check the LHS exists
				if state == nil {
					e := errors.New("= Expected a variable")
					return nil, nil, e
				}
				// Check the LHS is an assigned variable
				if state.Get_type() != EXEC_STATE_USER_VAR {
					e := errors.New("= Expected an assigned variable")
					return nil, nil, e
				}
				st := state.(Exec_state_user_var)
				var_name := st.Name
				// Evaluate RHS
				user_var, stopped_token, err := context.execute_statement()
				if err != nil {
					return nil, nil, err
				}
				if user_var == nil {
					e := errors.New("= Expected an expression")
					return nil, nil, e
				}
				context.user_variables[var_name] = user_var
				return user_var, stopped_token, nil
			}
			if context.ess_peek(0).Get_type() == EXEC_STATE_OPERATOR_MATHEMATICAL {
				// stop here
				if state.Get_type() == EXEC_STATE_USER_VAR {
					return state.(Exec_state_user_var).Variable, t, nil
				}
				return nil, t, nil
			}

			state = Exec_state_operator{}.From_user_var(state.(Exec_state_user_var), t.value)
			context.ess_push(state)
			defer context.ess_pop()
			// evaluate for the rhs
			user_var, stopped_token, err := context.execute_statement()
			if err != nil {
				return nil, nil, err
			}
			// execute the operator
			op_state := state.(Exec_state_operator)
			result, _ := op_state.Operator_func(op_state.Left, user_var)
			return result, stopped_token, nil
		case KEYWORD:
			switch t.value {
			case "GOTO":
				// Get the next token
				t, _ := context.next_token_line_aware()
				if t.token_type != NUMBER {
					e := errors.New("GOTO: Expected a line number")
					return nil, nil, e
				}
				line_no_as_int, _ := strconv.Atoi(t.value)
				idx := context.get_idx_of_line_no(line_no_as_int)
				if idx == -1 {
					e := errors.New("GOTO: Line number not found")
					return nil, nil, e
				}
				context.program_counter = idx
				context.tokeniser_channel = make(chan Token)
				go context.tokenise_line(context.program[context.program_counter].line)
				return nil, nil, nil
			case "THEN", "END", "ELSE":
				if state != nil {
					return state.(Exec_state_user_var).Variable, t, nil
				}
				return nil, t, nil
			case "IF":
				_, err := context.handle_if()
				return nil, nil, err
			case "LET":
				stop, err := context.handle_let()
				if err != nil {
					return nil, nil, err
				}
				return nil, stop, nil
			}
		}
	}
}

func (context *Context) handle_let() (*Token, error) {
	// Get the next token
	t, _ := context.next_token_line_aware()
	if t.token_type == USER_VAR {
		e := errors.New("LET: Variable already exists")
		return nil, e
	}
	if t.token_type != USER_VAR_UNASSIGNED {
		e := errors.New("LET: Expected a variable name")
		return nil, e
	}
	var_name := t.value
	// Get the next token
	t, _ = context.next_token_line_aware()
	if t.token_type != OPERATOR || t.value != "=" {
		e := errors.New("LET: Expected =")
		return nil, e
	}
	// Evaluate the expression
	user_var, stop, err := context.execute_statement()
	if err != nil {
		return nil, err
	}
	if user_var == nil {
		e := errors.New("LET: Expected an expression")
		return nil, e
	}
	context.user_variables[var_name] = user_var
	return stop, nil
}

func (context *Context) handle_if() (*Token, error) {
	st := Run_state_if{}
	st.Condition_pc = context.program_counter
	context.run_state_stack = append(context.run_state_stack, st)
	// eval condition
	condition, stopped_token, err := context.execute_statement()
	if err != nil {
		return nil, err
	}
	if stopped_token.token_type != KEYWORD && stopped_token.value != "THEN" {
		e := errors.New("IF: Expected THEN")
		return nil, e
	}
	hit_end := false
	cnd_val := condition.(variable.VARTYPE_NUMBER).Value().(float64)
	if cnd_val == 1 {
		// Keep executing until ELSE or END
		for {
			_, t, err := context.execute_statement()
			if err != nil {
				return nil, err
			}
			if t.token_type == KEYWORD {
				if t.value == "ELSE" {
					break
				}
				if t.value == "END" {
					hit_end = true
					break
				}
			}
		}
	} else {
		// Skip to ELSE or END
		for !hit_end {
			t, _ := context.next_token_line_aware()
			if t.token_type == DELIMITER_EOL {
				continue
			}
			if t.token_type == KEYWORD {
				if t.value == "ELSE" {
					// execute until end
					for {
						_, stopped_token, err := context.execute_statement()
						if err != nil {
							return nil, err
						}
						if stopped_token.token_type == KEYWORD && stopped_token.value == "END" {
							hit_end = true
							break
						}
					}
				} else if t.value == "END" {
					hit_end = true
					break
				}
			}
		}
	}

	if !hit_end {
		// Skip to END
		for {
			t, neof := context.next_token_line_aware()
			if !neof {
				e := errors.New("IF: Expected END")
				return nil, e
			}
			if t.token_type == KEYWORD && t.value == "END" {
				break
			}
		}
	}
	// pop the stack
	context.run_state_stack = context.run_state_stack[:len(context.run_state_stack)-1]
	return nil, nil
}

func (context *Context) get_idx_of_line_no(line_no int) int {
	for idx, line := range context.program {
		if line.line_number == line_no {
			return idx
		}
	}
	return -1
}

func (context *Context) sort_program() {
	if context.program == nil {
		return
	}
	sort.SliceStable(context.program, func(i, j int) bool {
		return context.program[i].line_number < context.program[j].line_number
	})
}

func (context *Context) Run(interrupt chan bool, done chan bool) {
	// clear args
	context.user_variables = make(map[string]variable.User_variable)
	context.sort_program()
	context.program_counter = 0
	go context.tokenise_line(context.program[context.program_counter].line)
	for context.program_counter < len(context.program) {
		select {
		case <-interrupt:
			context.terminal.Printline("INTERRUPT")
			done <- true
			return
		default:
			_, _, err := context.execute_statement()
			if err != nil {
				context.terminal.Printline("ERROR: " + err.Error())
				done <- true
				return
			}
		}
	}
	done <- true
}
