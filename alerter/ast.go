package alerter

// Expression is an arbitrary expression which returns a value when executed
type Expression interface {
	Execute(env *Env) (interface{}, error)
}

// Attribute is an attribute such as tx.origin or msg.value
type Attribute struct {
	Parts []string
}

// Execute retrieves the value of the attribute in the environment
func (a *Attribute) Execute(env *Env) (interface{}, error) {
	return nil, nil
}

// IntValue is a int wrapper implementing the Expression interface
type IntValue struct {
	Value int
}

// Execute return the wrapped value
func (i IntValue) Execute(env *Env) (interface{}, error) {
	return i.Value, nil
}

// StringValue is a string wrapper implementing the Expression interface
type StringValue struct {
	Value string
}

// Execute return the wrapped value
func (s StringValue) Execute(env *Env) (interface{}, error) {
	return s.Value, nil
}

// FunctionCall represents a function call and implements Expression
type FunctionCall struct {
	Arguments    []Expression
	FunctionName string
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
	Address string
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
	From     FromClause
	Where    Predicate
	Limit    LimitClause
	Since    SinceClause
	Until    UntilClause
	GroupBy  GroupByClause
}
