package alerter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	simpleInput string = `select   sum(msg.value)	FROM 0xabcdef  -- comment
	WHERE   msg.value  >= 5 -- other comment
	and msg.sig = "ab\"cd"
	group by BLOCKS(3)`
)

func TestNextToken(t *testing.T) {
	lexer := NewLexer(simpleInput)
	expectedTokens := []string{
		"select", "sum", "(", "msg", ".", "value", ")", "from", "0xabcdef",
		"where", "msg", ".", "value", ">=", "5",
		"and", "msg", ".", "sig", "=", "\"ab\\\"cd\"",
		"group by", "blocks", "(", "3", ")"}
	for _, expected := range expectedTokens {
		nextToken, hasNext, err := lexer.NextToken()
		assert.True(t, hasNext)
		assert.Equal(t, expected, nextToken)
		assert.Nil(t, err)
	}
	_, hasNext, err := lexer.NextToken()
	assert.Nil(t, err)
	assert.False(t, hasNext)
}
