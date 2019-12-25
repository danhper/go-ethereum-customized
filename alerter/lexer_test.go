package alerter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	simpleInput string = `select   sum(msg.value)	FROM 0xabcdef 
	WHERE   msg.value  >= 5
	group by BLOCKS(3)`
)

func TestNextToken(t *testing.T) {
	lexer := NewLexer(simpleInput)
	expectedTokens := []string{
		"select", "sum", "(", "msg", ".", "value", ")", "from", "0xabcdef",
		"where", "msg", ".", "value", ">=", "5",
		"group by", "blocks", "(", "3", ")"}
	for _, expected := range expectedTokens {
		nextToken, hasNext := lexer.NextToken()
		assert.True(t, hasNext)
		assert.Equal(t, expected, nextToken)
	}
	_, hasNext := lexer.NextToken()
	assert.False(t, hasNext)
}
