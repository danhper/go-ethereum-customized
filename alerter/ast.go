package alerter

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// BinaryFunction is a generic (and very type unsafe) binary function
type BinaryFunction func(Value, Value) Value

// IntBinaryFunction is a function which operates on two ints
type IntBinaryFunction func(*big.Int, *big.Int) *big.Int

// CompBinaryFunction is a function which operates on two ints and returns a bool
type CompBinaryFunction func(*big.Int, *big.Int) bool

// BoolBinaryFunction is a function which operates on two bools
type BoolBinaryFunction func(bool, bool) bool

// arithmetic binary operators
var arithBinaryOperators = map[string]IntBinaryFunction{
	"+": func(a, b *big.Int) *big.Int { return big.NewInt(0).Add(a, b) },
	"-": func(a, b *big.Int) *big.Int { return big.NewInt(0).Sub(a, b) },
	"*": func(a, b *big.Int) *big.Int { return big.NewInt(0).Mul(a, b) },
	"/": func(a, b *big.Int) *big.Int { return big.NewInt(0).Div(a, b) },
	"%": func(a, b *big.Int) *big.Int { return big.NewInt(0).Mod(a, b) },
}

// comparison binary operators
var comparisonBinaryOperators = map[string]CompBinaryFunction{
	">":  func(a, b *big.Int) bool { return a.Cmp(b) > 0 },
	">=": func(a, b *big.Int) bool { return a.Cmp(b) >= 0 },
	"<":  func(a, b *big.Int) bool { return a.Cmp(b) < 0 },
	"<=": func(a, b *big.Int) bool { return a.Cmp(b) <= 0 },
	"=":  func(a, b *big.Int) bool { return a.Cmp(b) == 0 },
	"<>": func(a, b *big.Int) bool { return a.Cmp(b) != 0 },
}

// comparison binary operators
var boolBinaryOperators = map[string]BoolBinaryFunction{
	"and": func(a, b bool) bool { return a && b },
	"or":  func(a, b bool) bool { return a || b },
}

// UnaryFunction is a generic unary function
type UnaryFunction func(Value) Value

// IntUnaryFunction is a function which takes and returns an int
type IntUnaryFunction func(*big.Int) *big.Int

// BoolUnaryFunction is a function which takes and returns an int
type BoolUnaryFunction func(bool) bool

var intUnaryOperators = map[string]IntUnaryFunction{
	// arithmetic operators
	"+": func(a *big.Int) *big.Int { return a },
	"-": func(a *big.Int) *big.Int { return big.NewInt(0).Neg(a) },
}

// comparison binary operators
var boolUnaryOperators = map[string]BoolUnaryFunction{
	"not": func(a bool) bool { return !a },
}

func wrapIntBinary(binaryFunc IntBinaryFunction) BinaryFunction {
	return func(left, right Value) Value {
		return NewIntValue(binaryFunc(left.ToInt(), right.ToInt()))
	}
}

func wrapCompBinary(binaryFunc CompBinaryFunction) BinaryFunction {
	return func(left, right Value) Value {
		return NewBoolValue(binaryFunc(left.ToInt(), right.ToInt()))
	}
}

func wrapBoolBinary(binaryFunc BoolBinaryFunction) BinaryFunction {
	return func(left, right Value) Value {
		return NewBoolValue(binaryFunc(left.ToBool(), right.ToBool()))
	}
}

func wrapIntUnary(unaryFunc IntUnaryFunction) UnaryFunction {
	return func(operand Value) Value {
		return NewIntValue(unaryFunc(operand.ToInt()))
	}
}

func wrapBoolUnary(unaryFunc BoolUnaryFunction) UnaryFunction {
	return func(operand Value) Value {
		return NewBoolValue(unaryFunc(operand.ToBool()))
	}
}

// BinaryOperator is an arbitrary binary operator which can take or
// return any type.
// In practice, it will be int -> int -> int or int -> int -> bool
// Type checking should be done at runtime by each operator :-(
type BinaryOperator interface {
	Apply(left Value, right Value) (Value, error)
	Equals(other interface{}) bool
}

// GenericBinaryOperator is a generic operator which takes two ints as operands
type GenericBinaryOperator struct {
	Name     string
	Operator BinaryFunction
}

// NewIntBinOperator returns a binary operator which operates on ints from a string
// representation of the operator
func NewIntBinOperator(rawOperator string) (BinaryOperator, error) {
	arithOperator, exists := arithBinaryOperators[rawOperator]
	if exists {
		return &GenericBinaryOperator{Name: rawOperator, Operator: wrapIntBinary(arithOperator)}, nil
	}

	compOperator, exists := comparisonBinaryOperators[rawOperator]
	if exists {
		return &GenericBinaryOperator{Name: rawOperator, Operator: wrapCompBinary(compOperator)}, nil
	}
	return nil, fmt.Errorf("expected a binary operator on ints, got: %s", rawOperator)
}

// IsComparisonOperator returns true if the given operator can be used as
// a comparison operator, i.e. int -> int -> bool
func IsComparisonOperator(rawOperator string) bool {
	_, exists := comparisonBinaryOperators[rawOperator]
	return exists
}

// NewCompOperator returns a binary operator which operates on ints and returns a bool
func NewCompOperator(rawOperator string) (BinaryOperator, error) {
	compOperator, exists := comparisonBinaryOperators[rawOperator]
	if exists {
		return &GenericBinaryOperator{Name: rawOperator, Operator: wrapCompBinary(compOperator)}, nil
	}
	return nil, fmt.Errorf("expected a comparison operator, got: %s", rawOperator)
}

// NewBoolBinOperator returns a binary operator which operates on ints and returns a bool
func NewBoolBinOperator(rawOperator string) (BinaryOperator, error) {
	compOperator, exists := boolBinaryOperators[rawOperator]
	if exists {
		return &GenericBinaryOperator{Name: rawOperator, Operator: wrapBoolBinary(compOperator)}, nil
	}
	return nil, fmt.Errorf("expected a boolean operator, got: %s", rawOperator)
}

// MustNewIntBinOperator returns a binary operator from a string
// or panics if it fails. This is mostly for tests purposes
func MustNewIntBinOperator(rawOperator string) BinaryOperator {
	op, err := NewIntBinOperator(rawOperator)
	if err != nil {
		panic(err)
	}
	return op
}

// Equals returns true if the operator is equal to other
func (a *GenericBinaryOperator) Equals(other interface{}) bool {
	otherIntOp, ok := other.(*GenericBinaryOperator)
	if !ok {
		return false
	}
	return a.Name == otherIntOp.Name
}

func (a *GenericBinaryOperator) String() string {
	return a.Name
}

// Apply casts the operands and executes the arithmetic operation
func (a *GenericBinaryOperator) Apply(left Value, right Value) (Value, error) {
	if !left.IsInt() {
		return nil, fmt.Errorf("cannot cast %v to int", left)
	}
	if !right.IsInt() {
		return nil, fmt.Errorf("cannot cast %v to int", right)
	}

	return a.Operator(left, right), nil
}

// UnaryOperator is an arbitrary unary operator which can take or
// return any type.
// In practice, it will be int -> int or bool -> bool
type UnaryOperator interface {
	Apply(operand Value) (Value, error)
	Equals(other interface{}) bool
}

// IntUnaryOperator is an operator which takes an int as operands
type IntUnaryOperator struct {
	Operator UnaryFunction
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
func (i *IntUnaryOperator) Apply(operand Value) (Value, error) {
	if !operand.IsInt() {
		return nil, fmt.Errorf("could not cast %v to int", operand)
	}
	return i.Operator(operand), nil
}

// NewIntUnaryOperator returns a unary operator from a string
// representation of the operator
func NewIntUnaryOperator(rawOperator string) (UnaryOperator, error) {
	operator, exists := intUnaryOperators[rawOperator]
	if !exists {
		return nil, fmt.Errorf("operator %s not found", rawOperator)
	}
	return &IntUnaryOperator{Name: rawOperator, Operator: wrapIntUnary(operator)}, nil
}

// MustNewIntUnaryOperator returns a unary operator from a string
// or panics on failure
func MustNewIntUnaryOperator(rawOperator string) UnaryOperator {
	op, err := NewIntUnaryOperator(rawOperator)
	if err != nil {
		panic(err)
	}
	return op
}

// NewBoolUnaryOperator returns a unary boolean operator from a string
func NewBoolUnaryOperator(rawOperator string) (UnaryOperator, error) {
	operator, exists := boolUnaryOperators[rawOperator]
	if !exists {
		return nil, fmt.Errorf("operator %s not found", rawOperator)
	}
	return &IntUnaryOperator{Name: rawOperator, Operator: wrapBoolUnary(operator)}, nil
}

// Expression is an arbitrary expression which returns a value when executed
type Expression interface {
	Execute(env *Env) (Value, error)
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
func (a *Attribute) Execute(env *Env) (Value, error) {
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

// NewIntBinaryApplication returns a new binary application which takes
// from the left and right, which should both evaluate to ints
// and the operands raw symbol for the operator
func NewIntBinaryApplication(left, right Expression, rawOperator string) (Expression, error) {
	operator, err := NewIntBinOperator(rawOperator)
	if err != nil {
		return nil, err
	}
	return &BinaryApplication{
		Left:     left,
		Right:    right,
		Operator: operator,
	}, nil
}

// MustNewIntBinaryApplication does the same thing as NewBinaryApplication
// but panics on error
func MustNewIntBinaryApplication(left, right Expression, rawOperator string) Expression {
	exp, err := NewIntBinaryApplication(left, right, rawOperator)
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
func (a *BinaryApplication) Execute(env *Env) (Value, error) {
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

// PredBinaryApplication is a binary application which evaluates to a bool
type PredBinaryApplication struct {
	*BinaryApplication
}

// NewCompBinaryApplication returns a new comparison application
func NewCompBinaryApplication(left, right Expression, rawOperator string) (Predicate, error) {
	operator, err := NewCompOperator(rawOperator)
	if err != nil {
		return nil, err
	}
	return &PredBinaryApplication{
		BinaryApplication: &BinaryApplication{
			Left:     left,
			Right:    right,
			Operator: operator,
		},
	}, nil
}

// MustNewCompBinaryApplication wraps NewCompBinaryApplication but panics on failure
func MustNewCompBinaryApplication(left, right Expression, rawOperator string) Predicate {
	pred, err := NewCompBinaryApplication(left, right, rawOperator)
	if err != nil {
		panic(err)
	}
	return pred
}

// Equals returns true if the app equals other
func (app *PredBinaryApplication) Equals(rawOther interface{}) bool {
	if other, ok := rawOther.(*PredBinaryApplication); ok {
		return app.BinaryApplication.Equals(other.BinaryApplication)
	}
	return false
}

// NewBoolBinaryApplication returns a new comparison application
func NewBoolBinaryApplication(left, right Predicate, rawOperator string) (Predicate, error) {
	operator, err := NewBoolBinOperator(rawOperator)
	if err != nil {
		return nil, err
	}
	return &PredBinaryApplication{
		BinaryApplication: &BinaryApplication{
			Left:     left,
			Right:    right,
			Operator: operator,
		},
	}, nil
}

// MustNewBoolBinaryApplication wraps NewBoolBinaryApplication but panics on failure
func MustNewBoolBinaryApplication(left, right Predicate, rawOperator string) Predicate {
	pred, err := NewBoolBinaryApplication(left, right, rawOperator)
	if err != nil {
		panic(err)
	}
	return pred
}

// ExecuteBool evaluates the value and converts the result to a bool
func (app *PredBinaryApplication) ExecuteBool(env *Env) (bool, error) {
	resValue, err := app.BinaryApplication.Execute(env)
	if err != nil {
		return false, err
	}
	if resValue.IsBool() {
		return resValue.ToBool(), nil
	}
	return false, fmt.Errorf("expected bool but returned %v", resValue)
}

// UnaryApplication is a nary application of an expression
// using the given operator
type UnaryApplication struct {
	Operand  Expression
	Operator UnaryOperator
}

// NewIntUnaryApplication returns a new unary application from
// the operand and the raw symbol for the operator
func NewIntUnaryApplication(operand Expression, rawOperator string) (Expression, error) {
	operator, err := NewIntUnaryOperator(rawOperator)
	if err != nil {
		return nil, err
	}
	return &UnaryApplication{
		Operand:  operand,
		Operator: operator,
	}, nil
}

// MustNewIntUnaryApplication is similar to NewUnaryApplication but panics on failure
func MustNewIntUnaryApplication(operand Expression, rawOperator string) Expression {
	app, err := NewIntUnaryApplication(operand, rawOperator)
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
func (a *UnaryApplication) Execute(env *Env) (Value, error) {
	operand, err := a.Operand.Execute(env)
	if err != nil {
		return nil, err
	}
	return a.Operator.Apply(operand)
}

// PredUnaryApplication is a unary application which evaluates to a bool
type PredUnaryApplication struct {
	*UnaryApplication
}

// NewPredUnaryApplication returns a new application evaluating to bool
func NewPredUnaryApplication(operand Predicate, rawOperator string) (Predicate, error) {
	operator, err := NewBoolUnaryOperator(rawOperator)
	if err != nil {
		return nil, err
	}
	return &PredUnaryApplication{
		UnaryApplication: &UnaryApplication{
			Operand:  operand,
			Operator: operator,
		},
	}, nil
}

// NegatePredicate is a helper to create a unary application with NOT as an operator
func NegatePredicate(operand Predicate) Predicate {
	pred, err := NewPredUnaryApplication(operand, "not")
	if err != nil {
		panic(err)
	}
	return pred
}

// Equals returns true if the value equals other
func (a *PredUnaryApplication) Equals(rawOther interface{}) bool {
	if other, ok := rawOther.(*PredUnaryApplication); ok {
		return a.UnaryApplication.Equals(other.UnaryApplication)
	}
	return false
}

// ExecuteBool evaluates the value and converts the result to a bool
func (a *PredUnaryApplication) ExecuteBool(env *Env) (bool, error) {
	resValue, err := a.UnaryApplication.Execute(env)
	if err != nil {
		return false, err
	}
	if resValue.IsBool() {
		return resValue.ToBool(), nil
	}
	return false, fmt.Errorf("expected bool but returned %v", resValue.Raw())
}

// InOperator is EMQL IN operator: exp in (e1, e2, e3)
type InOperator struct {
	Needle   Expression
	Haystack []Expression
}

// NewInOperator returns an IN operator
func NewInOperator(needle Expression, haystack []Expression) Predicate {
	return &InOperator{
		Needle:   needle,
		Haystack: haystack,
	}
}

// Equals returns true if the two values are equal
func (op *InOperator) Equals(rawOther interface{}) bool {
	other, ok := rawOther.(*InOperator)
	if !ok {
		return false
	}
	if !op.Needle.Equals(other.Needle) {
		return false
	}
	if len(op.Haystack) != len(other.Haystack) {
		return false
	}
	for i := 0; i < len(op.Haystack); i++ {
		if !op.Haystack[i].Equals(other.Haystack[i]) {
			return false
		}
	}
	return true
}

func (op *InOperator) String() string {
	args := []string{fmt.Sprintf("%v", op.Needle)}
	for _, exp := range op.Haystack {
		args = append(args, fmt.Sprintf("%v", exp))
	}
	return fmt.Sprintf("(in %v)", strings.Join(args, " "))
}

// ExecuteBool executes all the expressions and checks if lhs
// is inlcuded in rhs
func (op *InOperator) ExecuteBool(env *Env) (bool, error) {
	lhs, err := op.Needle.Execute(env)
	if err != nil {
		return false, err
	}
	var rhs []interface{}
	for _, exp := range op.Haystack {
		val, err := exp.Execute(env)
		if err != nil {
			return false, err
		}
		rhs = append(rhs, val)
	}
	for _, val := range rhs {
		if lhs.Equals(val) {
			return true, nil
		}
	}
	return false, nil
}

// Execute wraps ExecuteBool to make the function compatible with Expression
func (op *InOperator) Execute(env *Env) (Value, error) {
	res, err := op.ExecuteBool(env)
	return NewBoolValue(res), err
}

// IsOperator is the IS operator: exp IS ADDRESS
type IsOperator struct {
	Operand Expression
	Target  string
}

// NewIsOperator returns a new IS operator
func NewIsOperator(operand Expression, target string) Predicate {
	return &IsOperator{
		Operand: operand,
		Target:  target,
	}
}

func (op *IsOperator) String() string {
	return fmt.Sprintf("(is %s %v)", op.Target, op.Operand)
}

// Equals returns true if the is operator equals other
func (op *IsOperator) Equals(rawOther interface{}) bool {
	if other, ok := rawOther.(*IsOperator); ok {
		return op.Target == other.Target && op.Operand.Equals(other.Operand)
	}
	return false
}

// ExecuteBool returns true if target is fulfilled
func (op *IsOperator) ExecuteBool(env *Env) (bool, error) {
	_, err := op.Operand.Execute(env)
	if err != nil {
		return false, err
	}
	// TODO: implement me
	return false, nil
}

// Execute wraps ExecuteBool
func (op *IsOperator) Execute(env *Env) (Value, error) {
	res, err := op.ExecuteBool(env)
	return NewBoolValue(res), err
}

// Value is a generic value which can be returned by an expression
type Value interface {
	Equals(other interface{}) bool
	Raw() interface{}
	ToBool() bool
	ToString() string
	ToInt() *big.Int
	IsBool() bool
	IsString() bool
	IsInt() bool
}

// IntValue is a int wrapper implementing the Expression interface
type IntValue struct {
	Value *big.Int
}

// NewIntValue constructs a new strings value
func NewIntValue(value *big.Int) *IntValue {
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

// Execute return the value as is
func (i *IntValue) Execute(env *Env) (Value, error) {
	return i, nil
}

// Raw return the raw wrapped value
func (i *IntValue) Raw() interface{} {
	return i.Value
}

// ToBool will panic on IntValue
func (i *IntValue) ToBool() bool {
	panic(fmt.Errorf("cannot convert int to bool"))
}

// ToString will panic on IntValue
func (i *IntValue) ToString() string {
	panic(fmt.Errorf("cannot convert int to string"))
}

// ToInt will return the underlying int
func (i *IntValue) ToInt() *big.Int {
	return i.Value
}

// IsInt is true for IntValue
func (i *IntValue) IsInt() bool {
	return true
}

// IsBool is false for IntValue
func (i *IntValue) IsBool() bool {
	return false
}

// IsString is false for IntValue
func (i *IntValue) IsString() bool {
	return false
}

// StringValue is a string wrapper implementing the Expression interface
type StringValue struct {
	Value string
}

// NewStringValue constructs a new strings value
func NewStringValue(value string) *StringValue {
	return &StringValue{Value: value}
}

func (s *StringValue) String() string {
	return s.Value
}

// Execute return the wrapped value
func (s *StringValue) Execute(env *Env) (Value, error) {
	return s, nil
}

// Equals returns true if the value equals other
func (s *StringValue) Equals(rawOther interface{}) bool {
	if other, ok := rawOther.(*StringValue); ok {
		return s.Value == other.Value
	}
	return false
}

// Raw return the raw wrapped value
func (s *StringValue) Raw() interface{} {
	return s.Value
}

// ToBool will panic on StringValue
func (s *StringValue) ToBool() bool {
	panic(fmt.Errorf("cannot convert string to bool"))
}

// ToString will return the underlying string
func (s *StringValue) ToString() string {
	return s.Value
}

// ToInt will panic on StringValue
func (s *StringValue) ToInt() *big.Int {
	panic(fmt.Errorf("cannot convert string to int"))
}

// IsInt is false for StringValue
func (s *StringValue) IsInt() bool {
	return false
}

// IsBool is false for StringValue
func (s *StringValue) IsBool() bool {
	return false
}

// IsString is true for StringValue
func (s *StringValue) IsString() bool {
	return true
}

// BoolValue is a bool wrapper implementing the Expression and Value interfaces
type BoolValue struct {
	Value bool
}

// NewBoolValue constructs a new strings value
func NewBoolValue(value bool) *BoolValue {
	return &BoolValue{Value: value}
}

func (b *BoolValue) String() string {
	return strconv.FormatBool(b.Value)
}

// Execute return the wrapped value
func (b *BoolValue) Execute(env *Env) (Value, error) {
	return b, nil
}

// ExecuteBool makes bool implement the predicate interface
func (b *BoolValue) ExecuteBool(env *Env) (bool, error) {
	return b.Value, nil
}

// Equals returns true if the value equals other
func (b *BoolValue) Equals(rawOther interface{}) bool {
	if other, ok := rawOther.(*BoolValue); ok {
		return b.Value == other.Value
	}
	return false
}

// Raw return the raw wrapped value
func (b *BoolValue) Raw() interface{} {
	return b.Value
}

// ToBool will return the underlying bool
func (b *BoolValue) ToBool() bool {
	return b.Value
}

// ToString will panic on BoolValue
func (b *BoolValue) ToString() string {
	panic(fmt.Errorf("cannot convert bool to string"))
}

// ToInt will panic on BoolValue
func (b *BoolValue) ToInt() *big.Int {
	panic(fmt.Errorf("cannot convert bool to int"))
}

// IsInt is false for BoolValue
func (b *BoolValue) IsInt() bool {
	return false
}

// IsBool is false for BoolValue
func (b *BoolValue) IsBool() bool {
	return true
}

// IsString is true for BoolValue
func (b *BoolValue) IsString() bool {
	return false
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
func (f *FunctionCall) Execute(env *Env) (Value, error) {
	var evaluatedArguments []Value
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
	Expression
	ExecuteBool(env *Env) (bool, error)
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
