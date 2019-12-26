package alerter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidIdentifier(t *testing.T) {
	testCases := map[string]bool{
		"abc":    true,
		"_abc":   true,
		"1abc":   false,
		"abc1":   true,
		"a_b2_c": true,
		"$abc":   false,
		"abc$":   false,
	}
	for input, expected := range testCases {
		assert.Equal(t, expected, IsValidIdentifier(input))
	}
}

var (
	one         = NewIntValue(big.NewInt(1))
	ten         = NewIntValue(big.NewInt(10))
	msgValue    = NewAttribute([]string{"msg", "value"})
	sumMsgValue = NewFunctionCall("sum", []Expression{msgValue})
)

func TestParseFactor(t *testing.T) {
	testCases := map[string]Expression{
		"1":                 one,
		"0x1":               one,
		"\"string\"":        NewStringValue("string"),
		"simple_attr":       NewAttribute([]string{"simple_attr"}),
		"msg.value":         msgValue,
		"op.call.arg.value": NewAttribute([]string{"op", "call", "arg", "value"}),
		"SUM(msg.value)":    NewFunctionCall("sum", []Expression{msgValue}),
		"(1)":               one,
	}
	for input, expected := range testCases {
		parser, err := NewParser(NewLexer(input))
		assert.Nil(t, err)
		exp, err := parser.parseFactor()
		assert.Nil(t, err)
		assert.True(t, expected.Equals(exp), "%v != %v", expected, exp)
	}
}

func TestParseUnary(t *testing.T) {
	testCases := map[string]Expression{
		"1":               one,
		"-1":              MustNewUnaryApplication(one, "-"),
		"-msg.value":      MustNewUnaryApplication(msgValue, "-"),
		"+SUM(msg.value)": MustNewUnaryApplication(NewFunctionCall("sum", []Expression{msgValue}), "+"),
	}
	for input, expected := range testCases {
		parser, err := NewParser(NewLexer(input))
		assert.Nil(t, err)
		exp, err := parser.parseUnary()
		assert.Nil(t, err)
		assert.True(t, expected.Equals(exp), "%v != %v", expected, exp)
	}
}

func TestParseTerm(t *testing.T) {
	testCases := map[string]Expression{
		"1 * 10":         MustNewBinaryApplication(one, ten, "*"),
		"msg.value / 10": MustNewBinaryApplication(msgValue, ten, "/"),
		"-SUM(msg.value) * 10 / COUNT(tx)": MustNewBinaryApplication(
			MustNewBinaryApplication(MustNewUnaryApplication(sumMsgValue, "-"), ten, "*"),
			NewFunctionCall("count", []Expression{NewAttribute([]string{"tx"})}),
			"/",
		),
	}
	for input, expected := range testCases {
		parser, err := NewParser(NewLexer(input))
		assert.Nil(t, err)
		exp, err := parser.parseTerm()
		assert.Nil(t, err)
		assert.True(t, expected.Equals(exp), "%v != %v", expected, exp)
	}
}

func TestParseExpression(t *testing.T) {
	testCases := map[string]Expression{
		"1 + 10":             MustNewBinaryApplication(one, ten, "+"),
		"msg.value + 1 / 10": MustNewBinaryApplication(msgValue, MustNewBinaryApplication(one, ten, "/"), "+"),
		"1 + -SUM(msg.value) * 10 / COUNT(tx) - 10": MustNewBinaryApplication(
			MustNewBinaryApplication(one,
				MustNewBinaryApplication(
					MustNewBinaryApplication(MustNewUnaryApplication(sumMsgValue, "-"), ten, "*"),
					NewFunctionCall("count", []Expression{NewAttribute([]string{"tx"})}),
					"/",
				),
				"+",
			),
			ten,
			"-",
		),
	}
	for input, expected := range testCases {
		parser, err := NewParser(NewLexer(input))
		assert.Nil(t, err)
		exp, err := parser.parseExpression()
		assert.Nil(t, err)
		assert.True(t, expected.Equals(exp), "%v != %v", expected, exp)
	}
}
