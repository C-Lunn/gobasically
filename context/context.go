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

func (context *Context) execute_line(line string) {
	// Split the line into identifiers
	// If the line is empty, skip it
	if len(line) == 0 {
		context.program_counter++
		return
	} else {
		ch := make(chan Token)
		go context.tokenise_line(line, ch)
		t, alive := context.next_token(ch)

		for alive {
			switch t.token_type {
			case STD_FCN:
				// Get the function
				fcn := functions.Std_fcns[t.value]
				// Get the arguments
				args := make([]variable.User_variable, 0)
				// Keep getting tokens until we see a delimiter or hit the end of the line
				for {
					t, alive = context.next_token(ch)
					if !alive {
						break
					}
					if t.token_type == STRING {
						args = append(args, variable.VARTYPE_STRING{}.New(t.value))
					} else if t.token_type == NUMBER {
						args = append(args, variable.VARTYPE_NUMBER{}.From_string(t.value))
					} else if t.token_type == USER_VAR {
						args = append(args, context.user_variables[t.value])
					} else if t.token_type == DELIMITER_COMMA {
						continue
					} else {
						break
					}
				}
				// Execute the function
				fcn(context.terminal, args)
			default:
				// Do nothing
			}
			if !alive {
				break
			} else {
				t, alive = context.next_token(ch)
			}
		}
	}

}

func (context *Context) Run() {
	context.parse_program()
	context.program_counter = 0
	for context.program_counter < len(context.program) {
		context.execute_line(context.program[context.program_counter].line)
		context.program_counter++
	}
}
