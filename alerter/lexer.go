package alerter

import (
	"bytes"
	"strings"
)

var caseInsensitiveTokens map[string]bool = map[string]bool{
	"select":       true,
	"from":         true,
	"as":           true,
	"where":        true,
	"since":        true,
	"until":        true,
	"group by":     true,
	"null":         true,
	"contract":     true,
	"and":          true,
	"or":           true,
	"not":          true,
	"blocks":       true,
	"transactions": true,
	"is":           true,
	"in":           true,
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\t'
}

func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// Lexer represents a lexer for a single query
type Lexer struct {
	query        string
	index        int
	hasNext      bool
	currentToken string
}

// NewLexer returns a new lexer initialized with the query
func NewLexer(query string) *Lexer {
	lexer := &Lexer{
		query:        query,
		index:        0,
		hasNext:      false,
		currentToken: "",
	}
	lexer.readToken()
	return lexer
}

// IsDone returns true if the lexer does not have any more tokens
func (l *Lexer) IsDone() bool {
	return l.index >= len(l.query)
}

// peek returns the current char without advancing
func (l *Lexer) peek() byte {
	if l.IsDone() {
		return '\000'
	}
	return l.query[l.index]
}

func (l *Lexer) advance() {
	if !l.IsDone() {
		l.index++
	}
}

func (l *Lexer) skipWhitespaces() {
	for isWhitespace(l.peek()) {
		l.advance()
	}
}

func (l *Lexer) readToken() {
	l.hasNext = true
	l.skipWhitespaces()
	if l.IsDone() {
		l.hasNext = false
		return
	}

	c := l.peek()

	// NOTE: we only have symbols of 1 char for now
	if !isAlphaNum(c) {
		l.advance()
		var token string
		if c == '<' && l.peek() == '>' {
			token = "<>"
			l.advance()
		} else if (c == '>' || c == '<') && l.peek() == '=' {
			token = string([]byte{c, l.peek()})
			l.advance()
		} else {
			token = string([]byte{c})
		}
		l.currentToken = token
		return
	}

	buffer := bytes.NewBufferString("")
	for isAlphaNum(c) {
		buffer.WriteByte(c)
		l.advance()
		c = l.peek()
	}

	l.currentToken = buffer.String()
	lowered := strings.ToLower(l.currentToken)
	if _, exists := caseInsensitiveTokens[lowered]; exists {
		l.currentToken = lowered
	}
}

// NextToken returns the next token in the stream
func (l *Lexer) NextToken() (string, bool) {
	if !l.hasNext {
		return "", false
	}
	token := l.currentToken
	l.readToken()

	// special case to treat group by as a single token
	if token == "group" && l.currentToken == "by" {
		token = "group by"
		l.readToken()
	}

	return token, true
}
