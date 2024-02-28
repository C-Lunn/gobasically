package context

type TOKEN_TYPE int8

const (
	NUMBER TOKEN_TYPE = iota
	STRING
	USER_VAR
	USER_VAR_UNASSIGNED
	STD_FCN
	USER_FCN
	OPERATOR
	OPERATOR_MATHEMATICAL
	DELIMITER_COMMA
	DELIMITER_SEMICOLON
	DELIMITER_LBRACKET
	DELIMITER_RBRACKET
	DELIMITER_LSQUARE
	DELIMITER_RSQUARE
	DELIMITER_LBRACE
	DELIMITER_RBRACE
	DELIMITER_COLON
	DELIMITER_DOT
	DELIMITER_QUOTE
	DELIMITER_BACKSLASH
	DELIMITER_EOL
	KEYWORD
)

type Token struct {
	token_type TOKEN_TYPE
	value      string
}

var Token_type_lookup = map[string]TOKEN_TYPE{
	"+":   OPERATOR_MATHEMATICAL,
	"-":   OPERATOR_MATHEMATICAL,
	"*":   OPERATOR_MATHEMATICAL,
	"/":   OPERATOR_MATHEMATICAL,
	"%":   OPERATOR_MATHEMATICAL,
	"^":   OPERATOR_MATHEMATICAL,
	"=":   OPERATOR,
	"==":  OPERATOR,
	"<>":  OPERATOR,
	"AND": OPERATOR,
	"OR":  OPERATOR,
	"<":   OPERATOR,
	">":   OPERATOR,
	"<=":  OPERATOR,
	">=":  OPERATOR,
	",":   DELIMITER_COMMA,
	";":   DELIMITER_SEMICOLON,
	"(":   DELIMITER_LBRACKET,
	")":   DELIMITER_RBRACKET,
	"[":   DELIMITER_LSQUARE,
	"]":   DELIMITER_RSQUARE,
	"{":   DELIMITER_LBRACE,
	"}":   DELIMITER_RBRACE,
	":":   DELIMITER_COLON,
	"\\":  DELIMITER_BACKSLASH,
}

func read_float(s string, offset int) (string, int) {
	if s[offset] == '-' {
		offset++
	}
	for i := offset; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' || s[i] == '.' {
			return s[offset:i], i
		}
	}
	return s[offset:], len(s)
}

func read_rest_of_string(s string, offset int) (string, int) {
	for i := offset; i < len(s); i++ {
		if s[i] == '"' {
			return s[offset:i], i
		}
	}
	return s[offset:], len(s)
}

func read_until_next_non_alpha(s string, offset int) (string, int) {
	for i := offset; i < len(s); i++ {
		if s[i] < 'A' || s[i] > 'z' {
			return s[offset:i], i
		}
	}
	return s[offset:], len(s)
}

func is_digit(c byte) bool {
	return c >= '0' && c <= '9'
}
