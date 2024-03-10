package context

import (
	"basicallygo/functions"
	"basicallygo/variable"
)

type Exec_state_type uint8

type Exec_state_user_var struct {
	Variable variable.User_variable
}

func (e *Exec_state_user_var) Get_type() Exec_state_type {
	return EXEC_STATE_USER_VAR
}

type Exec_state_unassigned struct {
	Variable string
}

func (e *Exec_state_unassigned) Get_type() Exec_state_type {
	return EXEC_STATE_UNASSIGNED
}

const (
	EXEC_STATE_FCN Exec_state_type = iota
	EXEC_STATE_OPERATOR
	EXEC_STATE_OPERATOR_MATHEMATICAL
	EXEC_STATE_USER_VAR
	EXEC_STATE_UNASSIGNED
)

type Exec_state interface {
	Get_type() Exec_state_type
}

type Exec_state_Fcn struct {
	Fcn  functions.Function
	Args []variable.User_variable
}

func (e *Exec_state_Fcn) Get_type() Exec_state_type {
	return EXEC_STATE_FCN
}

type Exec_state_operator struct {
	Operator      Token
	Operator_func functions.Operator_fcn
	Left          variable.User_variable
	Right         variable.User_variable
}

func (e *Exec_state_operator) Get_type() Exec_state_type {
	return EXEC_STATE_OPERATOR
}

func (e *Exec_state_operator) From_user_var(uv *Exec_state_user_var, op string) *Exec_state_operator {
	e.Left = uv.Variable
	e.Operator_func = functions.Ops[op]
	return e
}

type Exec_state_operator_mathematical struct {
	Operator      Token
	Operator_func functions.Operator_fcn
	Left          variable.User_variable
	Right         variable.User_variable
}

func (e *Exec_state_operator_mathematical) Get_type() Exec_state_type {
	return EXEC_STATE_OPERATOR_MATHEMATICAL
}

func (e *Exec_state_operator_mathematical) From_user_var(uv *Exec_state_user_var, op string) *Exec_state_operator_mathematical {
	e.Left = uv.Variable
	e.Operator_func = functions.Ops[op]
	return e
}
