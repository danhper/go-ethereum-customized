package alerter

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode"
)

const (
	// LookAhead is the number of values to allow to look ahead
	// Should be k for an LL(k) grammar
	// Grammar is currently LL(2)
	LookAhead int = 2
)

// Parser is a stateful parser to parse a single statement
type Parser struct {
	Lexer   *Lexer
	buffer  []string
	hasNext bool
}

// IsValidIdentifier returns true if value is a valid identifier:
// [a-zA-Z_][a-zA-Z0-9_]*
func IsValidIdentifier(value string) bool {
	if len(value) == 0 {
		return false
	}
	for i, r := range value {
		if !unicode.IsLetter(r) && r != '_' && (i == 0 || !unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

// StartsWithDigit returns true if the string starts with a digit
func StartsWithDigit(value string) bool {
	for _, r := range value {
		return unicode.IsDigit(r)
	}
	return false
}

// NewParser creates a new parser
func NewParser(lexer *Lexer) (*Parser, error) {
	parser := &Parser{
		Lexer:   lexer,
		buffer:  []string{},
		hasNext: true,
	}
	return parser, parser.advance()
}

// ParseSelect parses an EMQL select statement
func (p *Parser) ParseSelect() (*SelectStatement, error) {
	if err := p.eat("select"); err != nil {
		return nil, err
	}

	selected, aliases, err := p.parseSelectList()
	if err != nil {
		return nil, err
	}

	fromClause, err := p.parseFrom()
	if err != nil {
		return nil, err
	}

	return &SelectStatement{
		Selected: selected,
		From:     fromClause,
		Aliases:  aliases,
	}, nil
}

func (p *Parser) parseFrom() (*FromClause, error) {
	if err := p.eat("from"); err != nil {
		return nil, err
	}
	from, err := p.parseHex()
	if err != nil {
		return nil, err
	}
	return &FromClause{Address: from}, nil
}

// parseSelectList returns the expressions to be selected and
// a mapping of alias to expression
func (p *Parser) parseSelectList() (expressions []Expression, aliases map[string]Expression, err error) {
	aliases = make(map[string]Expression)
	if err = p.parseSelectElem(&expressions, aliases); err != nil {
		return
	}

	for p.peek() == "," {
		if err = p.advance(); err != nil {
			return
		}
		if err = p.parseSelectElem(&expressions, aliases); err != nil {
			return
		}
	}
	return
}

// parseSelectElem parses an element from the select list of the form
// expresion [as alias]
func (p *Parser) parseSelectElem(expressions *[]Expression, aliases map[string]Expression) error {
	expression, err := p.parseExpression()
	if err != nil {
		return err
	}

	*expressions = append(*expressions, expression)
	alias, err := p.parseAs()
	if err != nil {
		return err
	}
	if alias != "" {
		aliases[alias] = expression
	}
	return nil
}

func (p *Parser) parseAs() (string, error) {
	if p.peek() == "as" {
		if err := p.advance(); err != nil {
			return "", err
		}
		return p.eatIdentifier()
	}
	return "", nil
}

func (p *Parser) parseExpression() (Expression, error) {
	term, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	return p.parseRecExpression(term)
}

func (p *Parser) parseRecExpression(left Expression) (Expression, error) {
	token := p.peek()
	if token == "+" || token == "-" {
		if err := p.advance(); err != nil {
			return nil, err
		}

		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		app, err := NewBinaryApplication(left, right, token)
		if err != nil {
			return nil, err
		}
		return p.parseRecExpression(app)
	}
	return left, nil
}

func (p *Parser) parseTerm() (Expression, error) {
	unary, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	return p.parseRecTerm(unary)
}

func (p *Parser) parseRecTerm(left Expression) (Expression, error) {
	token := p.peek()
	if token == "*" || token == "/" || token == "%" {
		if err := p.advance(); err != nil {
			return nil, err
		}

		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		app, err := NewBinaryApplication(left, right, token)
		if err != nil {
			return nil, err
		}
		return p.parseRecTerm(app)
	}
	return left, nil
}

func (p *Parser) parseUnary() (Expression, error) {
	token := p.peek()
	if token == "+" || token == "-" {
		if err := p.advance(); err != nil {
			return nil, err
		}

		unary, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return NewUnaryApplication(unary, token)
	}
	return p.parseFactor()
}

func (p *Parser) parseFactor() (Expression, error) {
	token := p.peek()

	if strings.HasPrefix(token, "\"") { // string literal
		str, err := strconv.Unquote(token)
		if err != nil {
			return nil, err
		}
		return NewStringValue(str), p.advance()
	} else if strings.HasPrefix(token, "0x") { // base 16 int literal
		value, err := p.parseHex()
		return NewIntValue(value), err
	} else if StartsWithDigit(token) { // base 10 literal
		value, success := big.NewInt(0).SetString(token, 10)
		if !success {
			return nil, fmt.Errorf("failed to parse int literal %s", token)
		}
		return NewIntValue(value), p.advance()
	} else if token == "(" {
		p.advance()
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return exp, p.eat(")")
	} else if IsValidIdentifier(token) {
		if p.peekN(1) == "(" {
			return p.parseFuncCall()
		}
		return p.parseAttribute()
	}
	return nil, fmt.Errorf("expected factor, got %s", token)
}

func (p *Parser) parseHex() (*big.Int, error) {
	token := p.peek()
	if !strings.HasPrefix(token, "0x") {
		return nil, fmt.Errorf("expected hex number")
	}
	value, success := big.NewInt(0).SetString(token[2:], 16)
	if !success {
		return nil, fmt.Errorf("failed to parse int literal %s", token)
	}
	return value, p.advance()
}

func (p *Parser) parseFuncCall() (Expression, error) {
	funcName, err := p.eatIdentifier()
	if err != nil {
		return nil, err
	}
	args, err := p.parseArgsList()
	if err != nil {
		return nil, err
	}
	return NewFunctionCall(funcName, args), nil
}

func (p *Parser) parseArgsList() ([]Expression, error) {
	var arguments []Expression
	for i := 0; (i == 0 && p.peek() == "(") || (i > 0 && p.peek() == ","); i++ {
		p.advance()
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, exp)
	}
	return arguments, p.eat(")")
}

func (p *Parser) parseAttribute() (Expression, error) {
	id, err := p.eatIdentifier()
	if err != nil {
		return nil, err
	}
	return p.parseAttributeParts([]string{id})
}

func (p *Parser) parseAttributeParts(parts []string) (Expression, error) {
	for p.peek() == "." {
		p.advance()
		part, err := p.eatIdentifier()
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return NewAttribute(parts), nil
}

func (p *Parser) advance() error {
	if len(p.buffer) > 0 {
		p.buffer = p.buffer[1:]
	}
	for len(p.buffer) < LookAhead {
		hasNext, err := p.readToken()
		if err != nil {
			return err
		}
		if !hasNext {
			break
		}
	}
	return nil
}

func (p *Parser) isDone() bool {
	return !p.hasNext && len(p.buffer) == 0
}

func (p *Parser) readToken() (bool, error) {
	token, hasNext, err := p.Lexer.NextToken()
	if err != nil {
		return false, err
	}
	p.hasNext = hasNext
	if hasNext {
		p.buffer = append(p.buffer, token)
	}
	return hasNext, nil
}

func (p *Parser) peek() string {
	if len(p.buffer) == 0 {
		return ""
	}
	return p.buffer[0]
}

func (p *Parser) peekN(n int) string {
	if len(p.buffer) <= n {
		return ""
	}
	return p.buffer[n]
}

func (p *Parser) eatIdentifier() (string, error) {
	token := p.peek()
	if !IsValidIdentifier(token) {
		return "", fmt.Errorf("%s is not a valid identifier", token)
	}
	return token, p.advance()
}

func (p *Parser) eat(token string) error {
	if p.peek() != token {
		return fmt.Errorf("expected %s but got %s", token, p.peek())
	}
	return p.advance()
}
