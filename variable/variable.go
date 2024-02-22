package variable

import "strconv"

type VAR_TYPE int8

const (
	NUMBER VAR_TYPE = iota
	STRING
	ARRAY
	NULL
)

type User_variable interface {
	To_string() string
	Type_of() VAR_TYPE
	Value() interface{}
}

type VARTYPE_NUMBER struct {
	value float64
}

func (v VARTYPE_NUMBER) To_string() string {
	return strconv.FormatFloat(v.value, 'f', -1, 32)
}

func (v VARTYPE_NUMBER) Type_of() VAR_TYPE {
	return NUMBER
}

func (v VARTYPE_NUMBER) Value() interface{} {
	return v.value
}

func (v VARTYPE_NUMBER) New(value float64) VARTYPE_NUMBER {
	return VARTYPE_NUMBER{value}
}

func (v VARTYPE_NUMBER) From_string(value string) VARTYPE_NUMBER {
	result, _ := strconv.ParseFloat(value, 64)
	return VARTYPE_NUMBER{result}
}

type VARTYPE_STRING struct {
	value string
}

func (v VARTYPE_STRING) New(value string) VARTYPE_STRING {
	return VARTYPE_STRING{value}
}

func (v VARTYPE_STRING) To_string() string {
	return v.value
}

func (v VARTYPE_STRING) Type_of() VAR_TYPE {
	return STRING
}

func (v VARTYPE_STRING) Value() interface{} {
	return v.value
}

type VARTYPE_ARRAY struct {
	value []User_variable
}

func (v VARTYPE_ARRAY) To_string() string {
	result := "["
	for _, element := range v.value {
		result += element.To_string() + ", "
	}
	result += "]"
	return result
}

func (v VARTYPE_ARRAY) Type_of() VAR_TYPE {
	return ARRAY
}

func (v VARTYPE_ARRAY) Value() interface{} {
	return v.value
}

func (v VARTYPE_ARRAY) Get(index int) User_variable {
	return v.value[index]
}

func (v VARTYPE_ARRAY) Set(index int, value User_variable) {
	v.value[index] = value
}

type VARTYPE_NULL struct {
}

func (v VARTYPE_NULL) To_string() string {
	return "NULL"
}

func (v VARTYPE_NULL) Type_of() VAR_TYPE {
	return NULL
}

func (v VARTYPE_NULL) Value() interface{} {
	return nil
}
