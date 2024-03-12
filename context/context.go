package context

import (
	"basicallygo/functions"
	"basicallygo/terminal"
	"basicallygo/variable"
	"errors"
	"fmt"
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
	if len(context.exec_state_stack) == 0 {
		return nil
	}
	return context.exec_state_stack[len(context.exec_state_stack)-1-which]
}

func (context *Context) rss_push(state Run_state) {
	context.run_state_stack = append(context.run_state_stack, state)
}

func (context *Context) rss_pop() Run_state {
	if len(context.run_state_stack) == 0 {
		return nil
	}
	state := context.run_state_stack[len(context.run_state_stack)-1]
	context.run_state_stack = context.run_state_stack[:len(context.run_state_stack)-1]
	return state
}

func (context *Context) rss_replace_top(state Run_state) {
	context.run_state_stack[len(context.run_state_stack)-1] = state
}

func (context *Context) rss_peek(which int) Run_state {
	if len(context.run_state_stack) == 0 {
		return nil
	}
	return context.run_state_stack[len(context.run_state_stack)-1-which]
}

func (context *Context) rss_peek_type(which int) RUN_STATE {
	if len(context.run_state_stack) == 0 {
		return RUN_STATE_NORMAL
	}
	return context.run_state_stack[len(context.run_state_stack)-1-which].Get_state()
}

func (context *Context) Set_input_buffer(input_buffer string) {
	context.input_buffer = input_buffer
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

func (context *Context) Accept_line(line string) bool {
	err := context.parse_line(line)
	if err != nil {
		context.terminal.Printline(err.Error())
		return false
	}
	return true
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
				if str_offset >= len(line) {
					ch <- Token{token_type: DELIMITER_EOL, value: ""}
					return
				}
			} else {
				break
			}
		}
		// read next char
		next_char := line[str_offset]
		// Workaround for <>, OR, AND, ==, <=, >=
		if (next_char == 'O' || next_char == '<' || next_char == '=' || next_char == '>') && str_offset+1 < len(line) {
			if next_char == 'O' && line[str_offset+1] == 'R' {
				ch <- Token{token_type: OPERATOR, value: "OR"}
				str_offset += 2
				continue
			}
			if w := string(next_char) + string(line[str_offset+1]); w == "<=" || w == ">=" || w == "<>" || w == "==" {
				ch <- Token{token_type: OPERATOR, value: w}
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
		if _, ok := functions.Std_fcns[strings.ToUpper(word)]; ok {
			ch <- Token{token_type: STD_FCN, value: strings.ToUpper(word)}
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

func (context *Context) is_in_if() bool {
	return context.rss_peek_type(0) == RUN_STATE_IF
}

/**
 * Execute a statement
 * Returns when a delimiter is seen (i.e. semicolon, comma, or line end)
 * Idea is that it's called recursively
 */
func (context *Context) execute_statement(get_next ...interface{}) (variable.User_variable, *Token, error) {
	var state Exec_state
	get_next_token := true
	var t *Token
	alive := true
	if len(get_next) > 0 {
		get_next_token = get_next[0].(bool)
		if len(get_next) == 2 && !get_next_token {
			t = get_next[1].(*Token)
			if t == nil {
				return nil, nil, nil
			}
		} else if !get_next_token {
			return nil, nil, errors.New("INTERNAL ERROR")
		}
	}
	for {
		if get_next_token {
			if context.is_in_if() {
				// If we're in an if statement, firstly check if we're at THEN.
				// If we're at THEN, and the result is true, then simply keep executing.
				// If not, keep yielding tokens until we hit end, else, or eof
				st := context.rss_peek(0).(*Run_state_if)
				if !st.In_condition { // we're at THEN
					if st.Result {
						t, alive = context.next_token_line_aware()
						if !alive {
							err := errors.New("IF: RAN INTO PROGRAM END")
							return nil, nil, err
						}
						// handle special case: GOTO
						if t.token_type == KEYWORD && t.value == "GOTO" {
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
							// drain the channel so we don't leak goros
							for t.token_type != DELIMITER_EOL {
								tkn := <-context.tokeniser_channel
								t = &tkn
							}
							context.tokeniser_channel = make(chan Token)
							go context.tokenise_line(context.program[context.program_counter].line)
							context.rss_pop()
							return nil, nil, nil
						}
						st.In_condition = true
						t, alive = context.next_token_line_aware()
					} else {
						for {
							t, alive = context.next_token_line_aware()
							if !alive || t.token_type == KEYWORD && (t.value == "END" || t.value == "ELSE") {
								break
							}
						}
						if !alive {
							err := errors.New("IF: Expected END or ELSE")
							return nil, nil, err
						}
						if t.token_type == KEYWORD && t.value == "END" {
							context.rss_pop()
						} else {
							st.In_condition = true
							t, alive = context.next_token_line_aware()
						}
					}
				} else {
					t, alive = context.next_token_line_aware()
				}
			} else {
				t, alive = context.next_token_line_aware()
			}
		} else {
			get_next_token = true
		}
		if t == nil || !alive || t.token_type == DELIMITER_EOL {
			if state != nil {
				state_as_user_var, ok := state.(*Exec_state_user_var)
				if ok {
					return state_as_user_var.Variable, t, nil
				} else {
					return nil, t, nil
				}

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
			state = &Exec_state_Fcn{Fcn: functions.Std_fcns[t.value]}
			context.ess_push(state)
			fcn_state := state.(*Exec_state_Fcn)
			// slice for the arguments
			fcn_state.Args = make([]variable.User_variable, 0)
			// Evaluate each argument until ;, ), or line end
			var stopped_token *Token
			got_open_bracket := false
			// Peek the next token to check if it's (
			peek_tkn, _ := context.next_token_line_aware()
			if peek_tkn.token_type == DELIMITER_LBRACKET {
				got_open_bracket = true
			}
			for {
				user_var, stopped_tkn, err := context.execute_statement(got_open_bracket, peek_tkn)
				if err != nil {
					return nil, nil, err
				}
				if user_var != nil {
					fcn_state.Args = append(fcn_state.Args, user_var)
				}
				if stopped_tkn != nil {
					if stopped_tkn.token_type == DELIMITER_SEMICOLON ||
						(context.rss_peek_type(0) == RUN_STATE_IF && stopped_tkn.token_type == KEYWORD && stopped_tkn.value == "END") ||
						(context.rss_peek_type(0) == RUN_STATE_IF && stopped_tkn.token_type == KEYWORD && stopped_tkn.value == "ELSE") ||
						(context.rss_peek_type(0) == RUN_STATE_IF && stopped_tkn.token_type == KEYWORD && stopped_tkn.value == "THEN") ||
						stopped_tkn.token_type == DELIMITER_EOL ||
						(got_open_bracket && stopped_tkn.token_type == DELIMITER_RBRACKET) {
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
			if result != nil {
				s := &Exec_state_user_var{Variable: result}
				context.ess_replace_top(s)
				state = s
				defer context.ess_pop()
			} else {
				t = stopped_token
				context.ess_pop()
			}
			if got_open_bracket && stopped_token.token_type == DELIMITER_RBRACKET {
				// we DO want to get the next token so just continue
				continue
			} else {
				get_next_token = false
				continue
			}
		case USER_VAR:
			// If state hasn't been initialised, then it's the first (or maybe only) user var
			if state == nil {
				state = &Exec_state_user_var{Variable: context.user_variables[t.value]}
			}
			continue
		case USER_VAR_UNASSIGNED:
			// If state hasn't been initialised, then it's the first (or maybe only) user var
			line_no := context.program[context.program_counter].line_number
			err := errors.New("LINE " + fmt.Sprintf("%d", line_no) + " UNEXPECTED IDENTIFIER: " + t.value)
			return nil, nil, err
		case NUMBER:
			// Get the number
			if state == nil {
				v := &variable.VARTYPE_NUMBER{}
				v = v.From_string(t.value)
				state = &Exec_state_user_var{Variable: v}
			}
			continue
		case STRING:
			// Get the string
			if state == nil {
				state = &Exec_state_user_var{Variable: (&variable.VARTYPE_STRING{}).New(t.value)}
			}
			continue
		case DELIMITER_COMMA, DELIMITER_SEMICOLON, DELIMITER_RBRACKET, DELIMITER_RSQUARE:
			if state != nil {
				return state.(*Exec_state_user_var).Variable, t, nil
			}
		case DELIMITER_LBRACKET:
			res, stop, err := context.execute_statement()
			if err != nil {
				return nil, nil, err
			}
			if res == nil {
				e := errors.New("(: EXPECTED EXPRESSION")
				return nil, nil, e
			}
			if stop.token_type != DELIMITER_RBRACKET {
				e := errors.New("(: EXPECTED )")
				return nil, nil, e
			}
			state = &Exec_state_user_var{Variable: res}
		case DELIMITER_LSQUARE:
			if state == nil {
				e := errors.New("[: EXPECTED EXPRESSION")
				return nil, nil, e
			}
			if state.Get_type() == EXEC_STATE_USER_VAR {
				st := state.(*Exec_state_user_var)
				if st.Variable.Type_of() != variable.ARRAY && st.Variable.Type_of() != variable.STRING {
					e := errors.New(st.Variable.To_string() + ": NOT AN ARRAY OR STRING")
					return nil, nil, e
				}
				res, stop, err := context.execute_statement()
				if err != nil {
					return nil, nil, err
				}
				if res == nil {
					e := errors.New("[: EXPECTED EXPRESSION")
					return nil, nil, e
				}
				if res.Type_of() != variable.NUMBER {
					e := errors.New("[: EXPECTED A NUMBER")
					return nil, nil, e
				}
				if stop.token_type != DELIMITER_RSQUARE {
					e := errors.New("[: EXPECTED ]")
					return nil, nil, e
				}
				if st.Variable.Type_of() == variable.STRING {
					if int(res.Value().(float64)) >= len(st.Variable.Value().(string)) {
						e := errors.New("STRING INDEX OUT OF BOUNDS")
						return nil, nil, e
					}
					st.Variable = (&variable.VARTYPE_STRING{}).New(string(st.Variable.Value().(string)[int(res.Value().(float64))]))
					// Get next token
					t, alive = context.next_token_line_aware()
					if !alive || t.token_type == DELIMITER_EOL {
						return st.Variable, t, nil
					}
					get_next_token = false
					continue
				} else {
					if (int(res.Value().(float64))) >= st.Variable.(*variable.VARTYPE_ARRAY).Len() {
						e := errors.New("ARRAY INDEX OUT OF BOUNDS")
						return nil, nil, e
					}
					st.Variable = st.Variable.(*variable.VARTYPE_ARRAY).Get(int(res.Value().(float64)))
					// Get next token
					t, alive = context.next_token_line_aware()
					if !alive || t.token_type == DELIMITER_EOL {
						return st.Variable, t, nil
					}
					get_next_token = false
					continue
				}
			}
		case OPERATOR_MATHEMATICAL:
			if state == nil && t.value == "-" {
				// minus
				// Check if the next token is a number
				next_tkn, _ := context.next_token_line_aware()
				if next_tkn.token_type == NUMBER {
					state = &Exec_state_user_var{Variable: (&variable.VARTYPE_NUMBER{}).From_string("-" + next_tkn.value)}
					continue
				}
			}
			if state == nil {
				e := errors.New("OPERATOR_MATHEMATICAL: Expected a variable")
				return nil, nil, e
			}
			new_state := (&Exec_state_operator_mathematical{}).From_user_var(state.(*Exec_state_user_var), t.value)
			context.ess_push(new_state)
			defer context.ess_pop()
			user_var, stopped_token, err := context.execute_statement()
			if err != nil {
				return nil, nil, err
			}
			// execute the operator
			result, _ := new_state.Operator_func(new_state.Left, user_var)
			o_state := &Exec_state_user_var{Variable: result}
			context.ess_replace_top(o_state)
			state = o_state
			if stopped_token != nil {
				if stopped_token.token_type == OPERATOR {
					get_next_token = false
					t = stopped_token
					continue
				} else {
					return result, stopped_token, nil
				}
			} else {
				return result, nil, nil
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
				st := state.(*Exec_state_user_var)
				// Evaluate RHS
				user_var, stopped_token, err := context.execute_statement()
				if err != nil {
					return nil, nil, err
				}
				if user_var == nil {
					e := errors.New("= Expected an expression")
					return nil, nil, e
				}
				st.Variable.Set(user_var.Value())
				return user_var, stopped_token, nil
			}
			if context.ess_peek(0) != nil && context.ess_peek(0).Get_type() == EXEC_STATE_OPERATOR_MATHEMATICAL {
				// stop here
				if state.Get_type() == EXEC_STATE_USER_VAR {
					return state.(*Exec_state_user_var).Variable, t, nil
				}
				return nil, t, nil
			}

			state = (&Exec_state_operator{}).From_user_var(state.(*Exec_state_user_var), t.value)
			context.ess_push(state)
			defer context.ess_pop()
			// evaluate for the rhs
			user_var, stopped_token, err := context.execute_statement()
			if err != nil {
				return nil, nil, err
			}
			// execute the operator
			op_state := state.(*Exec_state_operator)
			result, _ := op_state.Operator_func(op_state.Left, user_var)
			return result, stopped_token, nil
		case KEYWORD:
			switch t.value {
			case "GOTO":
				if context.is_in_if() {
					e := errors.New("GOTO: NOT ALLOWED IN MULTILINE IF")
					return nil, nil, e
				}
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
				// drain the channel so we don't leak goros
				for t.token_type != DELIMITER_EOL {
					tkn := <-context.tokeniser_channel
					t = &tkn
				}
				context.tokeniser_channel = make(chan Token)
				go context.tokenise_line(context.program[context.program_counter].line)
				return nil, nil, nil
			case "THEN":
				if state != nil {
					return state.(*Exec_state_user_var).Variable, t, nil
				}
				return nil, t, nil
			case "ELSE":
				if context.is_in_if() {
					if context.rss_peek(0).(*Run_state_if).In_condition && context.rss_peek(0).(*Run_state_if).Result {
						// we're in the true part of the if, so skip to the end
						for {
							t, alive = context.next_token_line_aware()
							if !alive || t.token_type == KEYWORD && t.value == "END" {
								break
							}
						}
						if !alive {
							err := errors.New("ELSE: Expected END")
							return nil, nil, err
						}
						context.rss_pop()
						return nil, t, nil
					}
				} else {
					e := errors.New("ELSE: NOT IN IF")
					return nil, nil, e
				}
			case "END":
				if context.is_in_if() {
					context.rss_pop()
					return nil, t, nil
				}
			case "IF":
				stop, err := context.handle_if()
				if err != nil {
					return nil, nil, err
				}
				return nil, stop, nil
			case "LET":
				stop, err := context.handle_let()
				if err != nil {
					return nil, nil, err
				}
				return nil, stop, nil
			case "FOR":
				stop, err := context.handle_for()
				if err != nil {
					return nil, nil, err
				}
				return nil, stop, nil
			case "NEXT":
				if context.rss_peek_type(0) != RUN_STATE_FOR {
					e := errors.New("NEXT: NOT IN FOR")
					return nil, nil, e
				}
				st := context.rss_peek(0).(*Run_state_for)
				// increment the variable
				if st.Post_loop_func(st.Variable) {
					// loop again
					context.program_counter = st.First_pc
					return nil, nil, nil
				} else {
					delete(context.user_variables, st.Variable)
					context.rss_pop()
					return nil, nil, nil
				}
			case "TO", "STEP":
				if state != nil {
					return state.(*Exec_state_user_var).Variable, t, nil
				}
				return nil, t, nil
			case "DIM":
				stop, err := context.handle_dim()
				if err != nil {
					return nil, nil, err
				}
				return nil, stop, nil
			case "REM":
				// Ignore all tokens until delim_eol
				for {
					t, alive = context.next_token_line_aware()
					if !alive {
						return nil, nil, nil
					}
					if t.token_type == DELIMITER_EOL {
						get_next_token = false
						break
					}
				}
				continue
			}
		}
	}
}

func (context *Context) handle_dim() (*Token, error) {
	// Get the next token
	t, _ := context.next_token_line_aware()
	if t.token_type != USER_VAR_UNASSIGNED {
		e := errors.New("DIM: Expected a variable name")
		return nil, e
	}
	var_name := t.value
	// Get the next token
	t, _ = context.next_token_line_aware()
	if t.token_type != DELIMITER_LBRACKET {
		e := errors.New("DIM: Expected (")
		return nil, e
	}
	// Get the next token
	dimensions := make([]int, 0)
	for {
		res, stop, err := context.execute_statement()
		if err != nil {
			return nil, err
		}
		if res == nil {
			e := errors.New("DIM: Expected an expression")
			return nil, e
		}
		if res.Type_of() != variable.NUMBER {
			e := errors.New("DIM: Expected a number")
			return nil, e
		}
		dimensions = append(dimensions, int(res.Value().(float64)))
		if stop.token_type == DELIMITER_COMMA {
			continue
		}
		if stop.token_type == DELIMITER_RBRACKET {
			break
		}
		e := errors.New("DIM: Expected , or )")
		return nil, e
	}
	context.user_variables[var_name] = (&variable.VARTYPE_ARRAY{}).New(dimensions...)
	return t, nil
}

func (context *Context) handle_for() (*Token, error) {
	// Get the next token
	s := &Run_state_for{
		First_pc: context.program_counter,
	}
	context.rss_push(s)
	t, _ := context.next_token_line_aware()
	if t.token_type != USER_VAR_UNASSIGNED {
		e := errors.New("FOR: EXPECTED UNASSIGNED VARIABLE NAME")
		return nil, e
	}
	v_name := t.value
	// Get the next token
	t, _ = context.next_token_line_aware()
	if t.token_type != OPERATOR || t.value != "=" {
		e := errors.New("FOR: EXPECTED =")
		return nil, e
	}
	res, stop, err := context.execute_statement()
	if err != nil {
		return nil, err
	}
	if res == nil {
		e := errors.New("FOR: EXPECTED AN EXPRESSION")
		return nil, e
	}
	if res.Type_of() != variable.NUMBER {
		e := errors.New("FOR: EXPECTED A NUMBER")
		return nil, e
	}
	// add the variable to the user variables
	context.user_variables[v_name] = res
	s.Variable = v_name
	if stop.token_type != KEYWORD || stop.value != "TO" {
		e := errors.New("FOR: EXPECTED TO")
		return nil, e
	}
	stop_val, stop, err := context.execute_statement()
	if err != nil {
		return nil, err
	}
	if stop_val == nil {
		e := errors.New("FOR: EXPECTED AN EXPRESSION")
		return nil, e
	}
	if stop_val.Type_of() != variable.NUMBER {
		e := errors.New("FOR: EXPECTED A NUMBER")
		return nil, e
	}
	post_loop_inc := 1.0
	if stop.token_type == KEYWORD || stop.value == "STEP" {
		step_val, stop, err := context.execute_statement()
		if err != nil {
			return nil, err
		}
		if step_val == nil {
			e := errors.New("FOR TO: EXPECTED AN EXPRESSION")
			return nil, e
		}
		if step_val.Type_of() != variable.NUMBER {
			e := errors.New("FOR TO: EXPECTED A NUMBER")
			return nil, e
		}
		post_loop_inc = step_val.Value().(float64)
		if stop.token_type != DELIMITER_EOL {
			e := errors.New("FOR: EXPECTED NEWLINE AFTER STEP")
			return nil, e
		}
	} else if stop.token_type != DELIMITER_EOL {
		e := errors.New("FOR: EXPECTED NEWLINE")
		return nil, e
	}

	s.Post_loop_func = func(var_name string) bool {
		current := context.user_variables[var_name].Value().(float64)
		context.user_variables[var_name] = (&variable.VARTYPE_NUMBER{}).New(current + post_loop_inc)
		current = context.user_variables[var_name].Value().(float64)
		if post_loop_inc > 0 {
			return current < stop_val.Value().(float64)
		} else {
			return current >= stop_val.Value().(float64)
		}
	}
	return stop, nil
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
	st := &Run_state_if{}
	// eval condition
	condition, stopped_token, err := context.execute_statement()
	if err != nil {
		return nil, err
	}
	if stopped_token.token_type != KEYWORD && stopped_token.value != "THEN" {
		e := errors.New("IF: Expected THEN")
		return nil, e
	}
	context.rss_push(st)
	st.Touch = 0
	cnd_val, ok := condition.(*variable.VARTYPE_NUMBER).Value().(float64)

	if !ok {
		e := errors.New("IF ON LINE " + strconv.Itoa(context.program[context.program_counter].line_number) + ": RESULT NOT BOOL")
		return nil, e
	}
	if cnd_val != 0 && cnd_val != 1 {
		e := errors.New("IF ON LINE " + strconv.Itoa(context.program[context.program_counter].line_number) + ": RESULT NOT BOOL")
		return nil, e
	}
	st.Result = cnd_val == 1
	st.In_condition = false
	return stopped_token, err
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
	context.tokeniser_channel = make(chan Token)
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
