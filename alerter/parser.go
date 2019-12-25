package alerter

import (
	"fmt"
	"strconv"
	"strings"
)

type parser struct {
	Lexer        *Lexer
	currentToken string
	hasNext      bool
}

func newParser(lexer *Lexer) *parser {
	return &parser{
		Lexer:        lexer,
		currentToken: "",
		hasNext:      true,
	}
}

func (p *parser) parseSelect() (*SelectStatement, error) {
	p.advance()
	if err := p.eat("select"); err != nil {
		return nil, err
	}

	return nil, nil
}

// parseSelectList returns the expressions to be selected and
// a mapping of alias to expression
func (p *parser) parseSelectList() ([]Expression, map[string]Expression, error) {
	var expressions []Expression
	mapping := make(map[string]Expression)
	expression, err := p.parseExpression()
	if err != nil {
		return nil, nil, err
	}
	expressions = append(expressions, expression)

	for p.peek() == "," {
	}
	return expressions, mapping, nil
}

func (p *parser) parseExpression() (Expression, error) {
	return nil, nil
}

func (p *parser) parseAs() (string, error) {
	return "", nil
}

func (p *parser) parseInt() (int64, error) {
	return strconv.ParseInt(p.peek(), 10, 64)
}

func (p *parser) advance() {
	p.currentToken, p.hasNext = p.Lexer.NextToken()
}

func (p *parser) isDone() bool {
	return !p.hasNext
}

func (p *parser) peek() string {
	return p.currentToken
}

func (p *parser) peekLower() string {
	return strings.ToLower(p.currentToken)
}

func (p *parser) eat(token string) error {
	if p.peekLower() != strings.ToLower(token) {
		return fmt.Errorf("expected %s but got %s", token, p.peek())
	}
	p.advance()
	return nil
}
