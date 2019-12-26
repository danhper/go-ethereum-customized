package alerter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnaryOperatorEquals(t *testing.T) {
	testCases := []struct {
		op       UnaryOperator
		other    UnaryOperator
		expected bool
	}{
		{MustNewUnaryOperator("+"), MustNewUnaryOperator("+"), true},
		{MustNewUnaryOperator("-"), MustNewUnaryOperator("-"), true},
		{MustNewUnaryOperator("+"), MustNewUnaryOperator("-"), false},
	}
	for _, testCase := range testCases {
		actual := testCase.op.Equals(testCase.other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %v %v, expected %v", testCase.op, testCase.other, testCase.expected)
	}
}

func TestBinaryOperatorEquals(t *testing.T) {
	testCases := []struct {
		op       string
		other    string
		expected bool
	}{
		{"+", "+", true},
		{"-", "+", false},
		{"-", "-", true},
		{"+", ">", false},
		{"<", "+", false},
		{">", ">", true},
	}
	for _, testCase := range testCases {
		op, err := NewBinaryOperator(testCase.op)
		assert.Nil(t, err)
		other, err := NewBinaryOperator(testCase.other)
		assert.Nil(t, err)
		actual := op.Equals(other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %s %s, expected %v", testCase.op, testCase.other, testCase.expected)
	}
}

func TestIntValueEquals(t *testing.T) {
	testCases := []struct {
		value    Expression
		other    Expression
		expected bool
	}{
		{NewIntValue(big.NewInt(1)), NewIntValue(big.NewInt(1)), true},
		{NewIntValue(big.NewInt(1)), NewIntValue(big.NewInt(2)), false},
		{NewIntValue(big.NewInt(0x123)), NewIntValue(big.NewInt(0x123)), true},
		{NewIntValue(big.NewInt(-1)), NewIntValue(big.NewInt(1)), false},
		{NewIntValue(big.NewInt(-1)), NewStringValue("foo"), false},
	}
	for _, testCase := range testCases {
		actual := testCase.value.Equals(testCase.other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %v %v, expected %v", testCase.value, testCase.other, testCase.expected)
	}

}

func TestStringValueEquals(t *testing.T) {
	testCases := []struct {
		value    Expression
		other    Expression
		expected bool
	}{
		{NewStringValue(""), NewStringValue(""), true},
		{NewStringValue(""), NewStringValue("foo"), false},
		{NewStringValue("foo"), NewStringValue("foo"), true},
		{NewStringValue("fooo"), NewStringValue("foo"), false},
		{NewStringValue("fooo"), NewIntValue(big.NewInt(1)), false},
	}
	for _, testCase := range testCases {
		actual := testCase.value.Equals(testCase.other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %v %v, expected %v", testCase.value, testCase.other, testCase.expected)
	}
}

func TestBinaryApplicationEquals(t *testing.T) {
	testCases := []struct {
		value    Expression
		other    Expression
		expected bool
	}{
		{MustNewBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "+"),
			MustNewBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "+"), true},
		{MustNewBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "+"),
			MustNewBinaryApplication(NewStringValue("abcc"), NewStringValue("def"), "+"), false},
		{MustNewBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "+"),
			MustNewBinaryApplication(NewStringValue("def"), NewStringValue("abc"), "+"), false},
		{MustNewBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "+"),
			MustNewBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "-"), false},
	}
	for _, testCase := range testCases {
		actual := testCase.value.Equals(testCase.other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %v %v, expected %v", testCase.value, testCase.other, testCase.expected)
	}
}

func TestUnaryApplicationEquals(t *testing.T) {
	testCases := []struct {
		value    Expression
		other    Expression
		expected bool
	}{
		{MustNewUnaryApplication(NewStringValue("a"), "+"),
			MustNewUnaryApplication(NewStringValue("a"), "+"), true},
		{MustNewUnaryApplication(NewIntValue(big.NewInt(1)), "-"),
			MustNewUnaryApplication(NewIntValue(big.NewInt(1)), "-"), true},
		{MustNewUnaryApplication(NewIntValue(big.NewInt(1)), "-"),
			MustNewUnaryApplication(NewIntValue(big.NewInt(2)), "-"), false},
		{MustNewUnaryApplication(NewIntValue(big.NewInt(1)), "+"),
			MustNewUnaryApplication(NewIntValue(big.NewInt(1)), "-"), false},
	}
	for _, testCase := range testCases {
		actual := testCase.value.Equals(testCase.other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %v %v, expected %v", testCase.value, testCase.other, testCase.expected)
	}
}
