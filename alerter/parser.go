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
	var predicate Predicate = nil
	var since *BlockRef = nil
	var until *BlockRef = nil
	var limit *int64 = nil
	var offset *int64 = nil
	var groupBy *GroupByClause = nil

	if err := p.eat("select"); err != nil {
		return nil, err
	}

	selected, aliases, err := p.parseSelectList()
	if err != nil {
		return nil, err
	}

	if err := p.eat("from"); err != nil {
		return nil, err
	}
	fromClause, err := p.parseFrom()
	if err != nil {
		return nil, err
	}

	token := p.peek()
	if token == "where" {
		if predicate, err = p.parseWhere(); err != nil {
			return nil, err
		}
	}

	token = p.peek()
	if token == "since" {
		if err = p.advance(); err != nil {
			return nil, err
		}
		if since, err = p.parseBlockRef(); err != nil {
			return nil, err
		}
	}

	token = p.peek()
	if token == "until" {
		if err = p.advance(); err != nil {
			return nil, err
		}
		if until, err = p.parseBlockRef(); err != nil {
			return nil, err
		}
	}

	token = p.peek()
	if token == "limit" {
		if err = p.advance(); err != nil {
			return nil, err
		}
		limitValue, err := p.eatIntLiteral()
		if err != nil {
			return nil, err
		}
		limit = &limitValue
	}

	token = p.peek()
	if token == "offset" {
		if err = p.advance(); err != nil {
			return nil, err
		}
		offsetValue, err := p.eatIntLiteral()
		if err != nil {
			return nil, err
		}
		offset = &offsetValue
	}

	token = p.peek()
	if token == "group by" {
		if err = p.advance(); err != nil {
			return nil, err
		}
		if groupBy, err = p.parseGroupBy(); err != nil {
			return nil, err
		}
	}

	return &SelectStatement{
		Selected: selected,
		From:     fromClause,
		Where:    predicate,
		Since:    since,
		Until:    until,
		Limit:    limit,
		Offset:   offset,
		GroupBy:  groupBy,
		Aliases:  aliases,
	}, nil
}

func (p *Parser) parseFrom() (*FromClause, error) {
	from, err := p.parseHex()
	if err != nil {
		return nil, err
	}
	return &FromClause{Address: from}, nil
}

func (p *Parser) parseWhere() (Predicate, error) {
	if err := p.eat("where"); err != nil {
		return nil, err
	}
	return p.parseOrCondition()
}

func (p *Parser) parseBlockRef() (*BlockRef, error) {
	blockNum, err := p.eatIntLiteral()
	if err != nil {
		return nil, err
	}
	return NewBlockRef(blockNum), nil
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

func (p *Parser) parseGroupBy() (*GroupByClause, error) {
	groupByClause := NewGroupByClause()
	if err := p.parseGroupByElem(groupByClause); err != nil {
		return nil, err
	}
	token := p.peek()
	for token == "," {
		if err := p.advance(); err != nil {
			return nil, err
		}

		if err := p.parseGroupByElem(groupByClause); err != nil {
			return nil, err
		}
		token = p.peek()
	}
	return groupByClause, nil
}

func (p *Parser) parseGroupByElem(groupBy *GroupByClause) error {
	token := p.peek()

	if token == "blocks" || token == "transactions" {
		if err := p.advance(); err != nil {
			return err
		}
		if err := p.eat("("); err != nil {
			return err
		}

		value, err := p.eatIntLiteral()
		if err != nil {
			return err
		}
		if err := p.eat(")"); err != nil {
			return err
		}

		if token == "blocks" {
			if groupBy.BlocksCount != nil {
				return fmt.Errorf("can only group once by blocks")
			}
			groupBy.BlocksCount = &value
		}

		if token == "transactions" {
			if groupBy.TransactionsCount != nil {
				return fmt.Errorf("can only group once by transactions")
			}
			groupBy.TransactionsCount = &value
		}
	} else { // group using attribute
		attribute, err := p.parseAttribute()
		if err != nil {
			return err
		}
		groupBy.Attributes = append(groupBy.Attributes, attribute)
	}

	return nil
}

func (p *Parser) parseOrCondition() (Predicate, error) {
	predicate, err := p.parseAndCondition()
	if err != nil {
		return nil, err
	}
	return p.parseOrConditionRec(predicate)
}

func (p *Parser) parseOrConditionRec(left Predicate) (Predicate, error) {
	token := p.peek()
	if token == "or" {
		if err := p.advance(); err != nil {
			return nil, err
		}

		right, err := p.parseAndCondition()
		if err != nil {
			return nil, err
		}
		app, err := NewBoolBinaryApplication(left, right, token)
		if err != nil {
			return nil, err
		}
		return p.parseOrConditionRec(app)
	}
	return left, nil
}

func (p *Parser) parseAndCondition() (Predicate, error) {
	predicate, err := p.parseNegatablePredicate()
	if err != nil {
		return nil, err
	}
	return p.parseAndConditionRec(predicate)
}

func (p *Parser) parseAndConditionRec(left Predicate) (Predicate, error) {
	token := p.peek()
	if token == "and" {
		if err := p.advance(); err != nil {
			return nil, err
		}

		right, err := p.parseNegatablePredicate()
		if err != nil {
			return nil, err
		}
		app, err := NewBoolBinaryApplication(left, right, token)
		if err != nil {
			return nil, err
		}
		return p.parseAndConditionRec(app)
	}
	return left, nil
}

func (p *Parser) parseNegatablePredicate() (Predicate, error) {
	token := p.peek()
	if token == "not" {
		if err := p.advance(); err != nil {
			return nil, err
		}

		unary, err := p.parseNegatablePredicate()
		if err != nil {
			return nil, err
		}
		return NewPredUnaryApplication(unary, token)
	}
	return p.parseSimplePredicate()
}

func (p *Parser) parseSimplePredicate() (Predicate, error) {
	token := p.peek()

	if token == "(" {
		p.advance()
		predicate, err := p.parseOrCondition()
		if err != nil {
			return nil, err
		}
		p.eat(")")
		return predicate, nil
	}

	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	token = p.peek()

	// handle IN and NOT IN syntaxes
	if token == "not" || token == "in" {
		negate := false
		if token == "not" {
			negate = true
			if err := p.advance(); err != nil {
				return nil, err
			}
		}

		predicate, err := p.parseIn(exp)
		if err != nil {
			return nil, err
		}
		if negate {
			predicate = NegatePredicate(predicate)
		}
		return predicate, nil
	}

	if token == "is" {
		if err := p.advance(); err != nil {
			return nil, err
		}
		negate := false
		if p.peek() == "not" {
			p.advance()
			negate = true
		}
		target, err := p.eatIdentifier()
		if err != nil {
			return nil, err
		}
		predicate := NewIsOperator(exp, target)
		if negate {
			predicate = NegatePredicate(predicate)
		}
		return predicate, nil
	}

	if IsComparisonOperator(token) {
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return NewCompBinaryApplication(exp, right, token)
	}

	return nil, fmt.Errorf("expected predicate, token was %s", token)
}

func (p *Parser) parseIn(needle Expression) (Predicate, error) {
	if err := p.eat("in"); err != nil {
		return nil, err
	}
	haystack, err := p.parseExpList()
	if err != nil {
		return nil, err
	}
	if len(haystack) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	return NewInOperator(needle, haystack), nil
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
		app, err := NewIntBinaryApplication(left, right, token)
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
		app, err := NewIntBinaryApplication(left, right, token)
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
		return NewIntUnaryApplication(unary, token)
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
	args, err := p.parseExpList()
	if err != nil {
		return nil, err
	}
	return NewFunctionCall(funcName, args), nil
}

func (p *Parser) parseExpList() ([]Expression, error) {
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

func (p *Parser) parseAttribute() (*Attribute, error) {
	id, err := p.eatIdentifier()
	if err != nil {
		return nil, err
	}
	return p.parseAttributeParts([]string{id})
}

func (p *Parser) parseAttributeParts(parts []string) (*Attribute, error) {
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

func (p *Parser) eatIntLiteral() (res int64, err error) {
	token := p.peek()
	if strings.HasPrefix(token, "0x") {
		res, err = strconv.ParseInt(token[2:], 16, 64)
	} else {
		res, err = strconv.ParseInt(token, 10, 64)
	}
	if err != nil {
		return
	}
	return res, p.advance()
}
