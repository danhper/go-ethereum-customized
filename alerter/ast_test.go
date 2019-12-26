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
		{MustNewIntUnaryOperator("+"), MustNewIntUnaryOperator("+"), true},
		{MustNewIntUnaryOperator("-"), MustNewIntUnaryOperator("-"), true},
		{MustNewIntUnaryOperator("+"), MustNewIntUnaryOperator("-"), false},
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
		op, err := NewIntBinOperator(testCase.op)
		assert.Nil(t, err)
		other, err := NewIntBinOperator(testCase.other)
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
		{MustNewIntBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "+"),
			MustNewIntBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "+"), true},
		{MustNewIntBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "+"),
			MustNewIntBinaryApplication(NewStringValue("abcc"), NewStringValue("def"), "+"), false},
		{MustNewIntBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "+"),
			MustNewIntBinaryApplication(NewStringValue("def"), NewStringValue("abc"), "+"), false},
		{MustNewIntBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "+"),
			MustNewIntBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "-"), false},
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
		{MustNewIntUnaryApplication(NewStringValue("a"), "+"),
			MustNewIntUnaryApplication(NewStringValue("a"), "+"), true},
		{MustNewIntUnaryApplication(NewIntValue(big.NewInt(1)), "-"),
			MustNewIntUnaryApplication(NewIntValue(big.NewInt(1)), "-"), true},
		{MustNewIntUnaryApplication(NewIntValue(big.NewInt(1)), "-"),
			MustNewIntUnaryApplication(NewIntValue(big.NewInt(2)), "-"), false},
		{MustNewIntUnaryApplication(NewIntValue(big.NewInt(1)), "+"),
			MustNewIntUnaryApplication(NewIntValue(big.NewInt(1)), "-"), false},
	}
	for _, testCase := range testCases {
		actual := testCase.value.Equals(testCase.other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %v %v, expected %v", testCase.value, testCase.other, testCase.expected)
	}
}

func TestInOperatorEquals(t *testing.T) {
	testCases := []struct {
		value    Predicate
		other    Predicate
		expected bool
	}{
		{NewInOperator(NewIntValue(big.NewInt(1)), []Expression{NewIntValue(big.NewInt(1)), NewStringValue("a")}),
			NewInOperator(NewIntValue(big.NewInt(1)), []Expression{NewIntValue(big.NewInt(1)), NewStringValue("a")}),
			true},
		{NewInOperator(NewIntValue(big.NewInt(1)), []Expression{NewIntValue(big.NewInt(1)), NewStringValue("a")}),
			NewInOperator(NewIntValue(big.NewInt(2)), []Expression{NewIntValue(big.NewInt(1)), NewStringValue("a")}),
			false},
		{NewInOperator(NewIntValue(big.NewInt(1)), []Expression{NewIntValue(big.NewInt(1)), NewStringValue("a")}),
			NewInOperator(NewIntValue(big.NewInt(1)), []Expression{NewIntValue(big.NewInt(1))}),
			false},
	}

	for _, testCase := range testCases {
		actual := testCase.value.Equals(testCase.other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %v %v, expected %v", testCase.value, testCase.other, testCase.expected)
	}
}

func TestIsOperatorEquals(t *testing.T) {
	testCases := []struct {
		value    Predicate
		other    Predicate
		expected bool
	}{
		{NewIsOperator(NewIntValue(big.NewInt(0x123)), "address"),
			NewIsOperator(NewIntValue(big.NewInt(0x123)), "address"),
			true},
		{NewIsOperator(NewIntValue(big.NewInt(0x123)), "address"),
			NewIsOperator(NewIntValue(big.NewInt(0x1234)), "address"),
			false},
		{NewIsOperator(NewIntValue(big.NewInt(0x123)), "address"),
			NewIsOperator(NewIntValue(big.NewInt(0x123)), "null"),
			false},
	}

	for _, testCase := range testCases {
		actual := testCase.value.Equals(testCase.other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %v %v, expected %v", testCase.value, testCase.other, testCase.expected)
	}
}

func TestPredBinaryApplicationEquals(t *testing.T) {
	testCases := []struct {
		value    Predicate
		other    Predicate
		expected bool
	}{
		{MustNewCompBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "="),
			MustNewCompBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "="),
			true},
		{MustNewCompBinaryApplication(NewIntValue(big.NewInt(0x123)), NewIntValue(big.NewInt(0x456)), "<"),
			MustNewCompBinaryApplication(NewIntValue(big.NewInt(0x123)), NewIntValue(big.NewInt(0x456)), "<"),
			true},
		{MustNewCompBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "="),
			MustNewCompBinaryApplication(NewStringValue("abc"), NewStringValue("def"), ">"),
			false},
		{MustNewCompBinaryApplication(NewStringValue("abc"), NewStringValue("def"), "="),
			MustNewCompBinaryApplication(NewStringValue("abcd"), NewStringValue("def"), "="),
			false},
		{MustNewBoolBinaryApplication(NewBoolValue(true), NewBoolValue(false), "or"),
			MustNewBoolBinaryApplication(NewBoolValue(true), NewBoolValue(false), "or"),
			true},
		{MustNewBoolBinaryApplication(NewBoolValue(true), NewBoolValue(false), "or"),
			MustNewBoolBinaryApplication(NewBoolValue(true), NewBoolValue(false), "and"),
			false},
	}
	for _, testCase := range testCases {
		actual := testCase.value.Equals(testCase.other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %v %v, expected %v", testCase.value, testCase.other, testCase.expected)
	}
}

func TestPredUnaryApplicationEquals(t *testing.T) {
	testCases := []struct {
		value    Predicate
		other    Predicate
		expected bool
	}{
		{NegatePredicate(NewBoolValue(true)), NegatePredicate(NewBoolValue(true)), true},
		{NegatePredicate(NewBoolValue(false)), NegatePredicate(NewBoolValue(false)), true},
		{NegatePredicate(NewBoolValue(true)), NegatePredicate(NewBoolValue(false)), false},
		{NegatePredicate(NewBoolValue(false)), NegatePredicate(NewBoolValue(true)), false},
	}
	for _, testCase := range testCases {
		actual := testCase.value.Equals(testCase.other)
		assert.Equal(t, testCase.expected, actual,
			"failed with %v %v, expected %v", testCase.value, testCase.other, testCase.expected)
	}
}
