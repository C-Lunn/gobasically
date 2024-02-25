package context

import "strings"

var KEYWORDS = []string{
	"DATA",
	"DIM",
	"END",
	"ELSE",
	"FN",
	"FOR",
	"GOSUB",
	"GOTO",
	"IF",
	"INPUT",
	"LET",
	"NEXT",
	"NOT",
	"ON",
	"REM",
	"RETURN",
	"STEP",
	"STOP",
	"THEN",
	"TO",
}

func Is_keyword(word string) bool {
	word = strings.ToUpper(word)
	for _, keyword := range KEYWORDS {
		if keyword == word {
			return true
		}
	}
	return false
}
