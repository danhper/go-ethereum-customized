package alerter

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

// the poor man's set
var caseInsensitiveTokens = map[string]bool{
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

// Lexer represents a lexer for a single query
type Lexer struct {
	reader       *bufio.Reader
	hasNext      bool
	currentRune  *rune
	currentToken string
}

// NewLexer returns a new lexer initialized with the query
func NewLexer(query string) *Lexer {
	lexer := &Lexer{
		reader:       bufio.NewReader(bytes.NewBufferString(query)),
		hasNext:      false,
		currentToken: "",
		currentRune:  nil,
	}
	lexer.advance()
	lexer.readToken()
	return lexer
}

// IsDone returns true if the lexer does not have any more tokens
func (l *Lexer) IsDone() bool {
	return l.currentRune == nil
}

// peek returns the current rune without advancing
func (l *Lexer) peek() rune {
	return *l.currentRune
}

func (l *Lexer) advance() {
	res, _, err := l.reader.ReadRune()
	if err != nil {
		l.currentRune = nil
		return
	}
	l.currentRune = &res
}

func (l *Lexer) skipWhitespaces() {
	for !l.IsDone() && unicode.IsSpace(l.peek()) {
		l.advance()
	}
}

func (l *Lexer) readString() error {
	l.advance()
	buffer := bytes.NewBufferString("\"")
	c := l.peek()
	escaped := false

	for escaped || c != '"' {
		escaped = c == '\\'
		buffer.WriteRune(c)
		l.advance()
		if l.IsDone() {
			return fmt.Errorf("reached EOS inside string")
		}
		c = l.peek()
	}
	buffer.WriteRune('"')
	l.advance()

	l.currentToken = buffer.String()

	return nil
}

func (l *Lexer) readSymbol(c rune) {
	var token string
	if c == '<' && l.peek() == '>' {
		token = "<>"
		l.advance()
	} else if (c == '>' || c == '<') && l.peek() == '=' {
		token = string([]rune{c, l.peek()})
		l.advance()
	} else {
		token = string([]rune{c})
	}
	l.currentToken = token
}

func (l *Lexer) skipLine() {
	for !l.IsDone() && l.peek() != '\n' {
		l.advance()
	}
	l.advance()
}

func (l *Lexer) readToken() error {
	l.hasNext = true
	l.skipWhitespaces()
	if l.IsDone() {
		l.hasNext = false
		return nil
	}

	c := l.peek()

	if c == '"' {
		return l.readString()
	}

	if !(unicode.IsLetter(c) || unicode.IsDigit(c)) {
		l.advance()
		if c == '-' && l.peek() == '-' {
			l.skipLine()
			return l.readToken()
		}
		l.readSymbol(c)
		return nil
	}

	buffer := bytes.NewBufferString("")
	for unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {
		buffer.WriteRune(c)
		l.advance()
		if l.IsDone() {
			break
		}
		c = l.peek()
	}

	l.currentToken = buffer.String()
	lowered := strings.ToLower(l.currentToken)
	if _, exists := caseInsensitiveTokens[lowered]; exists {
		l.currentToken = lowered
	}
	return nil
}

// NextToken returns the next token in the stream
func (l *Lexer) NextToken() (string, bool, error) {
	var err error

	if !l.hasNext {
		return "", false, nil
	}
	token := l.currentToken
	if err = l.readToken(); err != nil {
		return "", false, err
	}

	// special case to treat group by as a single token
	if token == "group" && l.currentToken == "by" {
		token = "group by"
		err = l.readToken()
	}

	return token, true, err
}
