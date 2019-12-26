package alerter

import (
	"fmt"
	"math/big"
	"strings"
)

// IntBinaryFunction is a function which operates on two ints
type IntBinaryFunction func(*big.Int, *big.Int) interface{}

var intBinaryOperators = map[string]IntBinaryFunction{
	// arithmetic operators
	"+": func(a *big.Int, b *big.Int) interface{} { return big.NewInt(0).Add(a, b) },
	"-": func(a *big.Int, b *big.Int) interface{} { return big.NewInt(0).Sub(a, b) },
	"*": func(a *big.Int, b *big.Int) interface{} { return big.NewInt(0).Mul(a, b) },
	"/": func(a *big.Int, b *big.Int) interface{} { return big.NewInt(0).Div(a, b) },
	"%": func(a *big.Int, b *big.Int) interface{} { return big.NewInt(0).Mod(a, b) },

	// boolean operators
	">":  func(a *big.Int, b *big.Int) interface{} { return a.Cmp(b) > 0 },
	">=": func(a *big.Int, b *big.Int) interface{} { return a.Cmp(b) >= 0 },
	"<":  func(a *big.Int, b *big.Int) interface{} { return a.Cmp(b) < 0 },
	"<=": func(a *big.Int, b *big.Int) interface{} { return a.Cmp(b) <= 0 },
	"=":  func(a *big.Int, b *big.Int) interface{} { return a.Cmp(b) == 0 },
	"<>": func(a *big.Int, b *big.Int) interface{} { return a.Cmp(b) != 0 },
}

// IntUnaryFunction is a function which operates on an int
type IntUnaryFunction func(*big.Int) interface{}

var intUnaryOperators = map[string]IntUnaryFunction{
	// arithmetic operators
	"+": func(a *big.Int) interface{} { return a },
	"-": func(a *big.Int) interface{} { return big.NewInt(0).Neg(a) },
}

// BinaryOperator is an arbitrary binary operator which can take or
// return any type.
// In practice, it will be int -> int -> int or int -> int -> bool
// Type checking should be done at runtime by each operator :-(
type BinaryOperator interface {
	Apply(left interface{}, right interface{}) (interface{}, error)
	Equals(other interface{}) bool
}

// IntBinaryOperator is an operator which takes two ints as operands
type IntBinaryOperator struct {
	Name     string
	Operator IntBinaryFunction
}

// NewBinaryOperator returns a binary operator from a string
// representation of the operator
func NewBinaryOperator(rawOperator string) (BinaryOperator, error) {
	operator, exists := intBinaryOperators[rawOperator]
	if !exists {
		return nil, fmt.Errorf("operator %s not found", rawOperator)
	}
	return &IntBinaryOperator{Name: rawOperator, Operator: operator}, nil
}

// MustNewBinaryOperator returns a binary operator from a string
// or panics if it fails. This is mostly for tests purposes
func MustNewBinaryOperator(rawOperator string) BinaryOperator {
	op, err := NewBinaryOperator(rawOperator)
	if err != nil {
		panic(err)
	}
	return op
}

// Equals returns true if the operator is equal to other
func (a *IntBinaryOperator) Equals(other interface{}) bool {
	otherIntOp, ok := other.(*IntBinaryOperator)
	if !ok {
		return false
	}
	return a.Name == otherIntOp.Name
}

func (a *IntBinaryOperator) String() string {
	return a.Name
}

// Apply casts the operands and executes the arithmetic operation
func (a *IntBinaryOperator) Apply(left interface{}, right interface{}) (interface{}, error) {
	leftInt, ok := left.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("could not cast %v to int", left)
	}

	rightInt, ok := right.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("could not cast %v to int", right)
	}

	return a.Operator(leftInt, rightInt), nil
}

// UnaryOperator is an arbitrary unary operator which can take or
// return any type.
// In practice, it will be int -> int or bool -> bool
type UnaryOperator interface {
	Apply(operand interface{}) (interface{}, error)
	Equals(other interface{}) bool
}

// IntUnaryOperator is an operator which takes an int as operands
type IntUnaryOperator struct {
	Operator IntUnaryFunction
	Name     string
}

// Equals returns true if the operator equals other
func (i *IntUnaryOperator) Equals(other interface{}) bool {
	if otherOp, ok := other.(*IntUnaryOperator); ok {
		return i.Name == otherOp.Name
	}
	return false
}

func (i *IntUnaryOperator) String() string {
	return i.Name
}

// Apply casts the operand to an int and returns the result
func (i *IntUnaryOperator) Apply(operand interface{}) (interface{}, error) {
	value, ok := operand.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("could not cast %v to int", operand)
	}
	return i.Operator(value), nil
}

// NewUnaryOperator returns a unary operator from a string
// representation of the operator
func NewUnaryOperator(rawOperator string) (UnaryOperator, error) {
	operator, exists := intUnaryOperators[rawOperator]
	if !exists {
		return nil, fmt.Errorf("operator %s not found", rawOperator)
	}
	return &IntUnaryOperator{Name: rawOperator, Operator: operator}, nil
}

// MustNewUnaryOperator returns a unary operator from a string
// or panics on failure
func MustNewUnaryOperator(rawOperator string) UnaryOperator {
	op, err := NewUnaryOperator(rawOperator)
	if err != nil {
		panic(err)
	}
	return op
}

// Expression is an arbitrary expression which returns a value when executed
type Expression interface {
	Execute(env *Env) (interface{}, error)
	Equals(other interface{}) bool
}

// Attribute is an attribute such as tx.origin or msg.value
type Attribute struct {
	Parts []string
}

// NewAttribute returns a new attribute with the given parts
func NewAttribute(parts []string) *Attribute {
	return &Attribute{Parts: parts}
}

func (a *Attribute) String() string {
	return strings.Join(a.Parts, ".")
}

// Equals returns true if the value equals other
func (a *Attribute) Equals(rawOther interface{}) bool {
	other, ok := rawOther.(*Attribute)
	if !ok {
		return false
	}
	if len(a.Parts) != len(other.Parts) {
		return false
	}
	for i := 0; i < len(a.Parts); i++ {
		if a.Parts[i] != other.Parts[i] {
			return false
		}
	}

	return true
}

// Execute retrieves the value of the attribute in the environment
func (a *Attribute) Execute(env *Env) (interface{}, error) {
	// TODO: implement
	return nil, nil
}

// BinaryApplication is a binary application of two expressions
// using the given operator
type BinaryApplication struct {
	Left     Expression
	Right    Expression
	Operator BinaryOperator
}

// NewBinaryApplication returns a new binary application from the left and right
// operands as well as the raw symbol for the operator
func NewBinaryApplication(left, right Expression, rawOperator string) (Expression, error) {
	operator, err := NewBinaryOperator(rawOperator)
	if err != nil {
		return nil, err
	}
	return &BinaryApplication{
		Left:     left,
		Right:    right,
		Operator: operator,
	}, nil
}

// MustNewBinaryApplication does the same thing as NewBinaryApplication
// but panics on error
func MustNewBinaryApplication(left, right Expression, rawOperator string) Expression {
	exp, err := NewBinaryApplication(left, right, rawOperator)
	if err != nil {
		panic(err)
	}
	return exp
}

func (a *BinaryApplication) String() string {
	return fmt.Sprintf("(%v %v %v)", a.Operator, a.Left, a.Right)
}

// Equals returns true if the application equals other
func (a *BinaryApplication) Equals(rawOther interface{}) bool {
	if other, ok := rawOther.(*BinaryApplication); ok {
		return a.Operator.Equals(other.Operator) && a.Left.Equals(other.Left) && a.Right.Equals(other.Right)
	}
	return false
}

// Execute evaluates left and right operands and finally applies the operator
func (a *BinaryApplication) Execute(env *Env) (interface{}, error) {
	left, err := a.Left.Execute(env)
	if err != nil {
		return nil, err
	}

	right, err := a.Right.Execute(env)
	if err != nil {
		return nil, err
	}
	return a.Operator.Apply(left, right)
}

// UnaryApplication is a nary application of an expression
// using the given operator
type UnaryApplication struct {
	Operand  Expression
	Operator UnaryOperator
}

// NewUnaryApplication returns a new unary application from
// the operand and the raw symbol for the operator
func NewUnaryApplication(operand Expression, rawOperator string) (Expression, error) {
	operator, err := NewUnaryOperator(rawOperator)
	if err != nil {
		return nil, err
	}
	return &UnaryApplication{
		Operand:  operand,
		Operator: operator,
	}, nil
}

// MustNewUnaryApplication is similar to NewUnaryApplication but panics on failure
func MustNewUnaryApplication(operand Expression, rawOperator string) Expression {
	app, err := NewUnaryApplication(operand, rawOperator)
	if err != nil {
		panic(err)
	}
	return app
}

func (a *UnaryApplication) String() string {
	return fmt.Sprintf("(%v %v)", a.Operator, a.Operand)
}

// Equals returns true if the value equals other
func (a *UnaryApplication) Equals(rawOther interface{}) bool {
	if other, ok := rawOther.(*UnaryApplication); ok {
		return a.Operator.Equals(other.Operator) && a.Operand.Equals(other.Operand)
	}
	return false
}

// Execute evaluates the operand and applies the operator
func (a *UnaryApplication) Execute(env *Env) (interface{}, error) {
	operand, err := a.Operand.Execute(env)
	if err != nil {
		return nil, err
	}
	return a.Operator.Apply(operand)
}

// IntValue is a int wrapper implementing the Expression interface
type IntValue struct {
	Value *big.Int
}

// NewIntValue constructs a new strings value
func NewIntValue(value *big.Int) Expression {
	return &IntValue{Value: value}
}

func (i *IntValue) String() string {
	return i.Value.Text(10)
}

// Equals returns true if the value equals other
func (i *IntValue) Equals(rawOther interface{}) bool {
	if other, ok := rawOther.(*IntValue); ok {
		return i.Value.Cmp(other.Value) == 0
	}
	return false
}

// Execute return the wrapped value
func (i *IntValue) Execute(env *Env) (interface{}, error) {
	return i.Value, nil
}

// StringValue is a string wrapper implementing the Expression interface
type StringValue struct {
	Value string
}

// NewStringValue constructs a new strings value
func NewStringValue(value string) Expression {
	return &StringValue{Value: value}
}

func (s *StringValue) String() string {
	return s.Value
}

// Equals returns true if the value equals other
func (s *StringValue) Equals(rawOther interface{}) bool {
	if other, ok := rawOther.(*StringValue); ok {
		return s.Value == other.Value
	}
	return false
}

// Execute return the wrapped value
func (s *StringValue) Execute(env *Env) (interface{}, error) {
	return s.Value, nil
}

// FunctionCall represents a function call and implements Expression
type FunctionCall struct {
	FunctionName string
	Arguments    []Expression
}

// NewFunctionCall returns a new function call
func NewFunctionCall(name string, arguments []Expression) *FunctionCall {
	return &FunctionCall{
		FunctionName: strings.ToLower(name),
		Arguments:    arguments,
	}
}

func (f *FunctionCall) String() string {
	args := []string{}
	for _, arg := range f.Arguments {
		args = append(args, fmt.Sprintf("%v", arg))
	}
	return fmt.Sprintf("(%s %s)", f.FunctionName, strings.Join(args, " "))
}

// Equals evaluates returns true the value is equal to other
func (f *FunctionCall) Equals(rawOther interface{}) bool {
	other, ok := rawOther.(*FunctionCall)
	if !ok {
		return false
	}
	if f.FunctionName != other.FunctionName {
		return false
	}
	if len(f.Arguments) != len(other.Arguments) {
		return false
	}
	for i := 0; i < len(f.Arguments); i++ {
		if !f.Arguments[i].Equals(other.Arguments[i]) {
			return false
		}
	}
	return true
}

// Execute evaluates all the arguments of the function
// and calls the function
func (f *FunctionCall) Execute(env *Env) (interface{}, error) {
	var evaluatedArguments []interface{}
	for _, argument := range f.Arguments {
		result, err := argument.Execute(env)
		if err != nil {
			return nil, err
		}
		evaluatedArguments = append(evaluatedArguments, result)
	}
	return env.ExecuteFunction(f.FunctionName, evaluatedArguments...)
}

// FromClause is a from clause of a statement
type FromClause struct {
	// NOTE: can currently only be an address
	Address *big.Int
}

// Predicate is a node of the AST which should return a boolean when executed
type Predicate interface {
	Execute(env *Env) (bool, error)
}

// LimitClause is a limit clause
type LimitClause struct {
	Limit int64
}

// SinceClause is a since clause
type SinceClause struct {
	Since int64
}

// UntilClause is an until clause
type UntilClause struct {
	Until int64
}

// GroupByClause is a group by clause
type GroupByClause struct {
}

// SelectStatement is a full EMQL select statement
type SelectStatement struct {
	Selected []Expression
	From     *FromClause
	Where    Predicate
	Limit    LimitClause
	Since    SinceClause
	Until    UntilClause
	GroupBy  GroupByClause
	Aliases  map[string]Expression
}
