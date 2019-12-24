package alerter

import (
	"bytes"
	"fmt"
	"strconv"
)

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\t'
}

func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

type lexer struct {
	query string
	index int
}

func newLexer(query string) *lexer {
	return &lexer{
		query: query,
		index: 0,
	}
}

func (l *lexer) isDone() bool {
	return l.index >= len(l.query)
}

// peeks the current char
func (l *lexer) peek() byte {
	if l.isDone() {
		return '\000'
	}
	return l.query[l.index]
}

func (l *lexer) advance() {
	if !l.isDone() {
		l.index++
	}
}

func (l *lexer) skipWhitespaces() {
	for isWhitespace(l.peek()) {
		l.advance()
	}
}

func (l *lexer) nextToken() (string, bool) {
	l.skipWhitespaces()
	if l.isDone() {
		return "", false
	}

	c := l.peek()

	// NOTE: we only have symbols of 1 char for now
	if !isAlphaNum(c) {
		l.advance()
		return string([]byte{c}), true
	}

	buffer := bytes.NewBufferString("")
	for isAlphaNum(c) {
		buffer.WriteByte(c)
		l.advance()
		c = l.peek()
	}
	return buffer.String(), true
}

type parser struct {
	Lexer        *lexer
	currentToken string
	isDone       bool
}

func newParser(lexer *lexer) *parser {
	return &parser{
		Lexer:        lexer,
		currentToken: "",
		isDone:       false,
	}
}

func (p *parser) parseStatement() error {
	hasNext := false
	if p.currentToken, hasNext = p.Lexer.nextToken(); hasNext {
		return fmt.Errorf("empty statement")
	}
	p.isDone = false
	return nil
}

func (p *parser) parseInt() (int64, error) {
	return strconv.ParseInt(p.peek(), 10, 64)
}

func (p *parser) advance() {
	hasNext := false
	p.currentToken, hasNext = p.Lexer.nextToken()
	p.isDone = !hasNext
}

func (p *parser) peek() string {
	return p.currentToken
}

func (p *parser) eat(token string) error {
	if p.peek() != token {
		return fmt.Errorf("expected %s but got %s", token, p.peek())
	}
	p.advance()
	return nil
}
