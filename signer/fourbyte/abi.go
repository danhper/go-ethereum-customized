// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package fourbyte

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// decodedCallData is an internal type to represent a method call parsed according
// to an ABI method signature.
type decodedCallData struct {
	signature string
	name      string
	inputs    []decodedArgument
}

// decodedArgument is an internal type to represent an argument parsed according
// to an ABI method signature.
type decodedArgument struct {
	soltype abi.Argument
	value   interface{}
}

// String implements stringer interface, tries to use the underlying value-type
func (arg decodedArgument) String() string {
	var value string
	switch val := arg.value.(type) {
	case fmt.Stringer:
		value = val.String()
	default:
		value = fmt.Sprintf("%v", val)
	}
	return fmt.Sprintf("%v: %v", arg.soltype.Type.String(), value)
}

// String implements stringer interface for decodedCallData
func (cd decodedCallData) String() string {
	args := make([]string, len(cd.inputs))
	for i, arg := range cd.inputs {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", cd.name, strings.Join(args, ","))
}

// verifySelector checks whether the ABI encoded data blob matches the requested
// function signature.
func verifySelector(selector string, calldata []byte) (*decodedCallData, error) {
	// Parse the selector into an ABI JSON spec
	abidata, err := parseSelector(selector)
	if err != nil {
		return nil, err
	}
	// Parse the call data according to the requested selector
	return parseCallData(calldata, string(abidata))
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentifierSymbol(c byte) bool {
	return c == '$' || c == '_'
}

func parseToken(unescapedSelector string, isIdent bool) (string, string, error) {
	if len(unescapedSelector) == 0 {
		return "", "", fmt.Errorf("empty token")
	}
	firstChar := unescapedSelector[0]
	position := 1
	if !(isAlpha(firstChar) || (isIdent && isIdentifierSymbol(firstChar))) {
		return "", "", fmt.Errorf("invalid token start: %c", firstChar)
	}
	for position < len(unescapedSelector) {
		char := unescapedSelector[position]
		if !(isAlpha(char) || isDigit(char) || (isIdent && isIdentifierSymbol(char))) {
			break
		}
		position++
	}
	return unescapedSelector[:position], unescapedSelector[position:], nil
}

func parseIdentifier(unescapedSelector string) (string, string, error) {
	return parseToken(unescapedSelector, true)
}

func parseElementaryType(unescapedSelector string) (string, string, error) {
	parsedType, rest, err := parseToken(unescapedSelector, false)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse elementary type: %v", err)
	}
	// handle arrays
	for len(rest) > 0 && rest[0] == '[' {
		parsedType = parsedType + string(rest[0])
		rest = rest[1:]
		for len(rest) > 0 && isDigit(rest[0]) {
			parsedType = parsedType + string(rest[0])
			rest = rest[1:]
		}
		if len(rest) == 0 || rest[0] != ']' {
			return "", "", fmt.Errorf("failed to parse array: expected ']', got %c", unescapedSelector[0])
		}
		parsedType = parsedType + string(rest[0])
		rest = rest[1:]
	}
	return parsedType, rest, nil
}

func parseCompositeType(unescapedSelector string) ([]interface{}, string, error) {
	if len(unescapedSelector) == 0 || unescapedSelector[0] != '(' {
		return nil, "", fmt.Errorf("expected '(', got %c", unescapedSelector[0])
	}
	parsedType, rest, err := parseType(unescapedSelector[1:])
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse type: %v", err)
	}
	result := []interface{}{parsedType}
	for len(rest) > 0 && rest[0] != ')' {
		parsedType, rest, err = parseType(rest[1:])
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse type: %v", err)
		}
		result = append(result, parsedType)
	}
	if len(rest) == 0 || rest[0] != ')' {
		return nil, "", fmt.Errorf("expected ')', got '%s'", rest)
	}
	return result, rest[1:], nil
}

func parseType(unescapedSelector string) (interface{}, string, error) {
	if len(unescapedSelector) == 0 {
		return nil, "", fmt.Errorf("empty type")
	}
	if unescapedSelector[0] == '(' {
		return parseCompositeType(unescapedSelector)
	} else {
		return parseElementaryType(unescapedSelector)
	}
}

// parseSelector converts a method selector into an ABI JSON spec. The returned
// data is a valid JSON string which can be consumed by the standard abi package.
// Note, although uppercase letters are not part of the ABI spec, this function
// still accepts it as the general format is valid. It will be rejected later
// by the type checker.
func parseSelector(unescapedSelector string) ([]byte, error) {
	// Define a tiny fake ABI struct for JSON marshalling
	type fakeArg struct {
		Name       string    `json:"name"`
		Type       string    `json:"type"`
		Components []fakeArg `json:"components"`
	}
	type fakeABI struct {
		Name   string    `json:"name"`
		Type   string    `json:"type"`
		Inputs []fakeArg `json:"inputs"`
	}

	name, rest, err := parseIdentifier(unescapedSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse selector '%s': %v", unescapedSelector, err)
	}
	args := []interface{}{}
	if len(rest) >= 2 && rest[0] == '(' && rest[1] == ')' {
		rest = rest[2:]
	} else {
		args, rest, err = parseCompositeType(rest)
		if err != nil {
			return nil, fmt.Errorf("failed to parse selector '%s': %v", unescapedSelector, err)
		}
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("failed to parse selector '%s': unexpected string '%s'", unescapedSelector, rest)
	}

	var assembleArgs func([]interface{}) ([]fakeArg, error)
	assembleArgs = func(args []interface{}) ([]fakeArg, error) {
		arguments := make([]fakeArg, 0)
		for i, arg := range args {
			// generate dummy name to avoid unmarshal issues
			name := fmt.Sprintf("name%d", i)
			if s, ok := arg.(string); ok {
				arguments = append(arguments, fakeArg{name, s, nil})
			} else if components, ok := arg.([]interface{}); ok {
				subArgs, err := assembleArgs(components)
				if err != nil {
					return nil, fmt.Errorf("failed to assemble components: %v", err)
				}
				arguments = append(arguments, fakeArg{name, "tuple", subArgs})
			} else {
				return nil, fmt.Errorf("failed to assemble args: unexpected type %T", arg)
			}
		}
		return arguments, nil
	}

	// Reassemble the fake ABI and constuct the JSON
	fakeArgs, err := assembleArgs(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse selector: %v", err)
	}

	return json.Marshal([]fakeABI{{name, "function", fakeArgs}})
}

// parseCallData matches the provided call data against the ABI definition and
// returns a struct containing the actual go-typed values.
func parseCallData(calldata []byte, unescapedAbidata string) (*decodedCallData, error) {
	// Validate the call data that it has the 4byte prefix and the rest divisible by 32 bytes
	if len(calldata) < 4 {
		return nil, fmt.Errorf("invalid call data, incomplete method signature (%d bytes < 4)", len(calldata))
	}
	sigdata := calldata[:4]

	argdata := calldata[4:]
	if len(argdata)%32 != 0 {
		return nil, fmt.Errorf("invalid call data; length should be a multiple of 32 bytes (was %d)", len(argdata))
	}
	// Validate the called method and upack the call data accordingly
	abispec, err := abi.JSON(strings.NewReader(unescapedAbidata))
	if err != nil {
		return nil, fmt.Errorf("invalid method signature (%q): %v", unescapedAbidata, err)
	}
	method, err := abispec.MethodById(sigdata)
	if err != nil {
		return nil, err
	}
	values, err := method.Inputs.UnpackValues(argdata)
	if err != nil {
		return nil, fmt.Errorf("signature %q matches, but arguments mismatch: %v", method.String(), err)
	}
	// Everything valid, assemble the call infos for the signer
	decoded := decodedCallData{signature: method.Sig, name: method.RawName}
	for i := 0; i < len(method.Inputs); i++ {
		decoded.inputs = append(decoded.inputs, decodedArgument{
			soltype: method.Inputs[i],
			value:   values[i],
		})
	}
	// We're finished decoding the data. At this point, we encode the decoded data
	// to see if it matches with the original data. If we didn't do that, it would
	// be possible to stuff extra data into the arguments, which is not detected
	// by merely decoding the data.
	encoded, err := method.Inputs.PackValues(values)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(encoded, argdata) {
		was := common.Bytes2Hex(encoded)
		exp := common.Bytes2Hex(argdata)
		return nil, fmt.Errorf("WARNING: Supplied data is stuffed with extra data. \nWant %s\nHave %s\nfor method %v", exp, was, method.Sig)
	}
	return &decoded, nil
}
