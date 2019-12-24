package alerter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	simpleInput string = "select   sum(msg.value) \tFROM\n0xabcdef group by blocks(3)"
)

func TestNextToken(t *testing.T) {
	lexer := newLexer(simpleInput)
	expectedTokens := []string{
		"select", "sum", "(", "msg", ".", "value", ")", "FROM", "0xabcdef",
		"group", "by", "blocks", "(", "3", ")"}
	for _, expected := range expectedTokens {
		nextToken, hasNext := lexer.nextToken()
		assert.True(t, hasNext)
		assert.Equal(t, expected, nextToken)
	}
	_, hasNext := lexer.nextToken()
	assert.False(t, hasNext)
}
