package alerter

import "fmt"

// ContractMetrics is a set of metrics collected about a single contract
type ContractMetrics struct {
}

// BuiltinFunction is any function which is built-in
// and can be called from EMQL expressions
type BuiltinFunction func(*Env, ...Value) (Value, error)

// Env contains all the current environment held by the node
type Env struct {
	Metrics   map[string]ContractMetrics
	Functions map[string]BuiltinFunction
}

// ExecuteFunction retrieves the function from the environment
// and executes it
func (e *Env) ExecuteFunction(name string, args ...Value) (Value, error) {
	function, exists := e.Functions[name]
	if !exists {
		return nil, fmt.Errorf("unkonwn function %s", name)
	}
	return function(e, args...)
}
