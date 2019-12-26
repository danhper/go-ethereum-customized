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
	msgSender   = NewAttribute([]string{"msg", "sender"})
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
		"-1":              MustNewIntUnaryApplication(one, "-"),
		"-msg.value":      MustNewIntUnaryApplication(msgValue, "-"),
		"+SUM(msg.value)": MustNewIntUnaryApplication(NewFunctionCall("sum", []Expression{msgValue}), "+"),
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
		"1 * 10":         MustNewIntBinaryApplication(one, ten, "*"),
		"msg.value / 10": MustNewIntBinaryApplication(msgValue, ten, "/"),
		"-SUM(msg.value) * 10 / COUNT(tx)": MustNewIntBinaryApplication(
			MustNewIntBinaryApplication(MustNewIntUnaryApplication(sumMsgValue, "-"), ten, "*"),
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
		"1 + 10":             MustNewIntBinaryApplication(one, ten, "+"),
		"msg.value + 1 / 10": MustNewIntBinaryApplication(msgValue, MustNewIntBinaryApplication(one, ten, "/"), "+"),
		"1 + -SUM(msg.value) * 10 / COUNT(tx) - 10": MustNewIntBinaryApplication(
			MustNewIntBinaryApplication(one,
				MustNewIntBinaryApplication(
					MustNewIntBinaryApplication(MustNewIntUnaryApplication(sumMsgValue, "-"), ten, "*"),
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

func TestBasicSelect(t *testing.T) {
	query := "select sum(msg.value) / 10 as sum, count(tx) from 0x1234abcd"
	parser, err := NewParser(NewLexer(query))
	assert.Nil(t, err)
	stmt, err := parser.ParseSelect()
	assert.Nil(t, err)
	assert.Len(t, stmt.Selected, 2)
	assert.Len(t, stmt.Aliases, 1)
	firstExp := MustNewIntBinaryApplication(sumMsgValue, ten, "/")
	assert.True(t, firstExp.Equals(stmt.Selected[0]), "%v != %v", firstExp, stmt.Selected[0])
	assert.True(t, firstExp.Equals(stmt.Aliases["sum"]), "%v != %v", firstExp, stmt.Aliases["sum"])
	secondExp := NewFunctionCall("count", []Expression{NewAttribute([]string{"tx"})})
	assert.True(t, secondExp.Equals(stmt.Selected[1]), "%v != %v", secondExp, stmt.Selected[1])
	expectedAddress, _ := big.NewInt(0).SetString("1234abcd", 16)
	assert.Equal(t, expectedAddress, stmt.From.Address)
}

func TestSelectWithWhere(t *testing.T) {
	query := `select tx.hash from 0x1234abcd
	where SUM(msg.value) > 10 AND not (msg.sender is not address OR msg.sender = 0x54321 OR
		msg.sender in (0x123, 0x432))`

	parser, err := NewParser(NewLexer(query))
	assert.Nil(t, err)
	stmt, err := parser.ParseSelect()
	assert.Nil(t, err)
	expected := MustNewBoolBinaryApplication(
		MustNewCompBinaryApplication(sumMsgValue, ten, ">"),
		NegatePredicate(
			MustNewBoolBinaryApplication(
				MustNewBoolBinaryApplication(
					NegatePredicate(NewIsOperator(msgSender, "address")),
					MustNewCompBinaryApplication(msgSender, NewIntValue(big.NewInt(0x54321)), "="),
					"or",
				),
				NewInOperator(msgSender, []Expression{
					NewIntValue(big.NewInt(0x123)),
					NewIntValue(big.NewInt(0x432)),
				}),
				"or",
			),
		),
		"and",
	)
	assert.True(t, expected.Equals(stmt.Where), "expected != actual:\n%v != %v", expected, stmt.Where)
}
