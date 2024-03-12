package variable

import "strconv"

type VAR_TYPE int8

const (
	NUMBER VAR_TYPE = iota
	STRING
	ARRAY
	NULL
	VARIANT
)

type User_variable interface {
	To_string() string
	Type_of() VAR_TYPE
	Value() interface{}
	Set(v interface{})
}

type VARTYPE_NUMBER struct {
	value float64
}

func (v *VARTYPE_NUMBER) To_string() string {
	return strconv.FormatFloat(v.value, 'f', -1, 32)
}

func (v *VARTYPE_NUMBER) Type_of() VAR_TYPE {
	return NUMBER
}

func (v *VARTYPE_NUMBER) Value() interface{} {
	return v.value
}

func (v *VARTYPE_NUMBER) Set(value interface{}) {
	v.value = value.(float64)
}

func (v *VARTYPE_NUMBER) New(value float64) *VARTYPE_NUMBER {
	v.value = value
	return v
}

func (v *VARTYPE_NUMBER) From_string(value string) *VARTYPE_NUMBER {
	result, _ := strconv.ParseFloat(value, 64)
	return &VARTYPE_NUMBER{result}
}

type VARTYPE_STRING struct {
	value string
}

func (v *VARTYPE_STRING) New(value string) *VARTYPE_STRING {
	v.value = value
	return v
}

func (v *VARTYPE_STRING) To_string() string {
	return v.value
}

func (v *VARTYPE_STRING) Type_of() VAR_TYPE {
	return STRING
}

func (v *VARTYPE_STRING) Value() interface{} {
	return v.value
}

func (v *VARTYPE_STRING) Set(value interface{}) {
	v.value = value.(string)
}

type VARTYPE_ARRAY struct {
	value  []User_variable
	length int
}

func (v *VARTYPE_ARRAY) To_string() string {
	result := "["
	for _, element := range v.value {
		result += element.To_string() + ", "
	}
	result = result[:len(result)-2]
	result += "]"
	return result
}

func (v *VARTYPE_ARRAY) Len() int {
	return v.length
}

func (v *VARTYPE_ARRAY) Type_of() VAR_TYPE {
	return ARRAY
}

func (v *VARTYPE_ARRAY) Value() interface{} {
	return v.value
}

func (v *VARTYPE_ARRAY) Get(index int) User_variable {
	return v.value[index]
}

func (v *VARTYPE_ARRAY) Set(_ interface{}) {
	//
}

func (v *VARTYPE_ARRAY) New(dimensions ...int) *VARTYPE_ARRAY {
	//recursively create arrays
	result := make([]User_variable, dimensions[0])
	if len(dimensions) == 1 {
		for i := 0; i < dimensions[0]; i++ {
			result[i] = &VARTYPE_VARIANT{&VARTYPE_NULL{}}
		}
	} else {
		for i := 0; i < dimensions[0]; i++ {
			a := &VARTYPE_ARRAY{}
			result[i] = a.New(dimensions[1:]...)
		}
	}
	v.length = dimensions[0]
	v.value = result
	return v
}

type VARTYPE_NULL struct {
}

func (v *VARTYPE_NULL) To_string() string {
	return "NULL"
}

func (v *VARTYPE_NULL) Type_of() VAR_TYPE {
	return NULL
}

func (v *VARTYPE_NULL) Value() interface{} {
	return nil
}

func (v *VARTYPE_NULL) Set(_ interface{}) {
	//
}

type VARTYPE_VARIANT struct {
	value User_variable
}

func (v *VARTYPE_VARIANT) To_string() string {
	return v.value.To_string()
}

func (v *VARTYPE_VARIANT) Type_of() VAR_TYPE {
	return VARIANT
}

func (v *VARTYPE_VARIANT) Value() interface{} {
	return v.value
}

func (v *VARTYPE_VARIANT) Set(value interface{}) {
	switch v.value.Type_of() {
	case NUMBER:
		switch value.(type) {
		case float64, int:
			v.value.Set(value)
			return
		}
	case STRING:
		switch value.(type) {
		case string:
			v.value.Set(value)
			return
		}
	case ARRAY:
		switch value.(type) {
		case VARTYPE_ARRAY:
			v.value.Set(value)
			return
		}
	}

	switch value.(type) {
	case float64, int:
		v.value = &VARTYPE_NUMBER{value.(float64)}
	case string:
		v.value = &VARTYPE_STRING{value.(string)}
	case VARTYPE_ARRAY:
		v.value = value.(*VARTYPE_ARRAY)
	}
}
