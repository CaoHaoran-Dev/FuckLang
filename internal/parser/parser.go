package parser

import (
	"fmt"
	"fucklang/internal/ast"
	"fucklang/internal/lexer"
	"fucklang/internal/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
)

var precedences = map[token.Type]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       EQUALS,
	token.GT:       EQUALS,
	token.LTE:      EQUALS,
	token.GTE:      EQUALS,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.STAR:     PRODUCT,
	token.SLASH:    PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string { return p.errors }

func (p *Parser) addError(line int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if line > 0 {
		p.errors = append(p.errors, fmt.Sprintf("line %d: %s", line, msg))
	} else {
		p.errors = append(p.errors, msg)
	}
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.FUCK:
		return p.parsePrintStatement()
	case token.SHIT:
		return p.parseAssignStatement()
	case token.HELL:
		return p.parseLoopStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.ASS:
		return p.parseReturnStatement()
	case token.BITCH:
		if p.peekToken.Type == token.IDENT {
			return p.parseNamedFunctionStatement()
		}
		return p.parseExpressionStatement()
	case token.LBRACE:
		return p.parseBlockStatement()
	case token.SEMICOLON:
		return nil
	case token.ILLEGAL:
		p.addError(p.curToken.Line, "illegal token: %q", p.curToken.Literal)
		return nil
	default:
		return p.parseExpressionOrAssignStatement()
	}
}

func (p *Parser) parseBreakStatement() ast.Statement {
	return &ast.BreakStatement{Token: p.curToken}
}

func (p *Parser) parseNamedFunctionStatement() ast.Statement {
	bitchToken := p.curToken
	p.nextToken()

	if p.curToken.Type != token.IDENT {
		p.addError(p.curToken.Line, "expected function name after `bitch`, got %q", p.curToken.Literal)
		return nil
	}

	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	if p.curToken.Type != token.LPAREN {
		p.addError(p.curToken.Line, "expected `(` after function name `%s`, got %q", name.Value, p.curToken.Literal)
		return nil
	}

	params := p.parseFunctionParameters()
	if p.peekToken.Type == token.LBRACE {
		p.nextToken()
	}
	if p.curToken.Type != token.LBRACE {
		p.addError(p.curToken.Line, "expected `{` for function body, got %q", p.curToken.Literal)
		return nil
	}
	body := p.parseBlockStatement()

	fn := &ast.FunctionLiteral{Token: bitchToken, Parameters: params, Body: body}

	return &ast.AssignStatement{
		Token: token.Token{Type: token.SHIT, Literal: "shit", Line: bitchToken.Line},
		Name:  name,
		Value: fn,
	}
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	stmt := &ast.PrintStatement{Token: p.curToken}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if stmt.Value == nil {
		p.addError(stmt.Token.Line, "expected expression after `fuck`")
		return nil
	}
	return stmt
}

func (p *Parser) parseAssignStatement() ast.Statement {
	stmt := &ast.AssignStatement{Token: p.curToken}
	p.nextToken()
	if p.curToken.Type != token.IDENT {
		p.addError(p.curToken.Line, "expected identifier after `shit`, got %q", p.curToken.Literal)
		return nil
	}
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	if p.curToken.Type == token.LBRACKET {
		p.nextToken()
		idx := p.parseExpression(LOWEST)
		if idx == nil {
			p.addError(p.curToken.Line, "expected index expression in assignment")
			return nil
		}
		if p.peekToken.Type == token.RBRACKET {
			p.nextToken()
		} else {
			p.addError(p.peekToken.Line, "expected `]` after index expression, got %q", p.peekToken.Literal)
			return nil
		}
		p.nextToken()
		if p.curToken.Type != token.ASSIGN {
			p.addError(p.curToken.Line, "expected `=` in index assignment, got %q", p.curToken.Literal)
			return nil
		}
		p.nextToken()
		val := p.parseExpression(LOWEST)
		if val == nil {
			p.addError(p.curToken.Line, "expected value expression after `=`")
			return nil
		}
		return &ast.IndexAssignStatement{Token: stmt.Token, Left: ident, Index: idx, Value: val}
	}

	if p.curToken.Type != token.ASSIGN {
		p.addError(p.curToken.Line, "expected `=` after identifier `%s`, got %q", ident.Value, p.curToken.Literal)
		return nil
	}
	p.nextToken()
	stmt.Name = ident
	stmt.Value = p.parseExpression(LOWEST)
	if stmt.Value == nil {
		p.addError(stmt.Token.Line, "expected expression after `=`")
		return nil
	}
	return stmt
}

func (p *Parser) parseLoopStatement() *ast.LoopStatement {
	stmt := &ast.LoopStatement{Token: p.curToken}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if stmt.Condition == nil {
		p.addError(stmt.Token.Line, "expected condition after `hell`")
		return nil
	}
	if p.peekToken.Type == token.LBRACE {
		p.nextToken()
	}
	if p.curToken.Type != token.LBRACE {
		p.addError(p.curToken.Line, "expected `{` for loop body, got %q", p.curToken.Literal)
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if stmt.Value == nil {
		p.addError(stmt.Token.Line, "expected expression after `ass`")
		return nil
	}
	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	p.nextToken()
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		if p.curToken.Type == token.ILLEGAL {
			p.addError(p.curToken.Line, "illegal token: %q", p.curToken.Literal)
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	if p.curToken.Type == token.EOF {
		p.addError(block.Token.Line, "unterminated block, expected `}`")
	}
	return block
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if stmt.Expression == nil {
		return nil
	}
	return stmt
}

func (p *Parser) parseExpressionOrAssignStatement() ast.Statement {
	startToken := p.curToken
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}

	if p.peekToken.Type == token.ASSIGN {
		p.nextToken()
		p.nextToken()

		val := p.parseExpression(LOWEST)
		if val == nil {
			p.addError(startToken.Line, "expected expression after `=`")
			return nil
		}

		switch left := expr.(type) {
		case *ast.Identifier:
			return &ast.AssignStatement{
				Token: token.Token{Type: token.SHIT, Literal: "shit", Line: startToken.Line},
				Name:  left,
				Value: val,
			}
		case *ast.IndexExpression:
			return &ast.IndexAssignStatement{
				Token: token.Token{Type: token.SHIT, Literal: "shit", Line: startToken.Line},
				Left:  left.Left,
				Index: left.Index,
				Value: val,
			}
		default:
			p.addError(startToken.Line, "invalid assignment target: %s", expr.TokenLiteral())
			return nil
		}
	}

	return &ast.ExpressionStatement{Token: startToken, Expression: expr}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	var left ast.Expression
	switch p.curToken.Type {
	case token.INT:
		val, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
		if err != nil {
			p.addError(p.curToken.Line, "invalid integer: %q", p.curToken.Literal)
			return nil
		}
		left = &ast.IntegerLiteral{Token: p.curToken, Value: val}
	case token.STRING:
		left = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case token.TRUE:
		left = &ast.Boolean{Token: p.curToken, Value: true}
	case token.FALSE:
		left = &ast.Boolean{Token: p.curToken, Value: false}
	case token.IDENT:
		left = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	case token.BANG, token.MINUS:
		left = p.parsePrefixExpression()
	case token.LPAREN:
		left = p.parseGroupedExpression()
	case token.DAMN:
		return p.parseIfExpression()
	case token.BITCH:
		return p.parseFunctionLiteral()
	case token.LBRACKET:
		left = p.parseArrayLiteral()
	case token.LBRACE:
		left = p.parseHashLiteral()
	case token.SEMICOLON:
		return nil
	case token.ILLEGAL:
		p.addError(p.curToken.Line, "illegal token: %q", p.curToken.Literal)
		return nil
	default:
		p.addError(p.curToken.Line, "unexpected token in expression: %q", p.curToken.Literal)
		return nil
	}

	if left == nil {
		return nil
	}

	for precedence < p.peekPrecedence() {
		p.nextToken()
		left = p.parseInfixExpression(left)
		if left == nil {
			return nil
		}
	}
	return left
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	if expr.Right == nil {
		return nil
	}
	return expr
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}
	if p.peekToken.Type != token.RPAREN {
		p.addError(p.peekToken.Line, "expected `)`, got %q", p.peekToken.Literal)
		return nil
	}
	p.nextToken()
	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	switch p.curToken.Type {
	case token.LPAREN:
		return p.parseCallExpression(left)
	case token.LBRACKET:
		return p.parseIndexExpression(left)
	}
	expr := &ast.InfixExpression{Token: p.curToken, Left: left, Operator: p.curToken.Literal}
	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)
	if expr.Right == nil {
		return nil
	}
	return expr
}

func (p *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{Token: p.curToken}
	p.nextToken()
	expr.Condition = p.parseExpression(LOWEST)
	if expr.Condition == nil {
		p.addError(expr.Token.Line, "expected condition after `damn`")
		return nil
	}
	if p.peekToken.Type == token.LBRACE {
		p.nextToken()
	}
	if p.curToken.Type != token.LBRACE {
		p.addError(p.curToken.Line, "expected `{` after condition, got %q", p.curToken.Literal)
		return nil
	}
	expr.Consequence = p.parseBlockStatement()

	if p.peekToken.Type == token.ELSE {
		p.nextToken()
		p.nextToken()
		if p.curToken.Type == token.LBRACE {
			expr.Alternative = p.parseBlockStatement()
		} else if p.curToken.Type == token.DAMN {
			ifExpr := p.parseIfExpression()
			if ifExpr == nil {
				return nil
			}
			altBlock := &ast.BlockStatement{
				Token:      p.curToken,
				Statements: []ast.Statement{&ast.ExpressionStatement{Token: p.curToken, Expression: ifExpr}},
			}
			expr.Alternative = altBlock
		} else {
			p.addError(p.curToken.Line, "expected `{` or `damn` after `else`, got %q", p.curToken.Literal)
			return nil
		}
	}
	return expr
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curToken}
	p.nextToken()
	if p.curToken.Type != token.LPAREN {
		p.addError(p.curToken.Line, "expected `(` after `bitch`, got %q", p.curToken.Literal)
		return nil
	}
	fn.Parameters = p.parseFunctionParameters()
	if p.peekToken.Type == token.LBRACE {
		p.nextToken()
	}
	if p.curToken.Type != token.LBRACE {
		p.addError(p.curToken.Line, "expected `{` for function body, got %q", p.curToken.Literal)
		return nil
	}
	fn.Body = p.parseBlockStatement()
	return fn
}

func (p *Parser) parseFunctionParameters() []*ast.Parameter {
	params := []*ast.Parameter{}

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return params
	}

	p.nextToken()
	if p.curToken.Type != token.IDENT {
		p.addError(p.curToken.Line, "expected parameter name, got %q", p.curToken.Literal)
		return params
	}
	param := &ast.Parameter{Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}}
	params = append(params, param)

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		if p.curToken.Type != token.IDENT {
			p.addError(p.curToken.Line, "expected parameter name, got %q", p.curToken.Literal)
			return params
		}
		param := &ast.Parameter{Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}}
		params = append(params, param)
	}

	if p.peekToken.Type != token.RPAREN {
		p.addError(p.peekToken.Line, "expected `)` in parameter list, got %q", p.peekToken.Literal)
		return params
	}
	p.nextToken()
	return params
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	expr := &ast.CallExpression{Token: p.curToken, Function: function}
	expr.Arguments = p.parseExpressionList(token.RPAREN)
	if expr.Arguments == nil {
		return nil
	}
	return expr
}

func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	var list []ast.Expression
	if p.peekToken.Type == end {
		p.nextToken()
		return list
	}
	p.nextToken()
	first := p.parseExpression(LOWEST)
	if first == nil {
		return nil
	}
	list = append(list, first)
	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		elem := p.parseExpression(LOWEST)
		if elem == nil {
			return nil
		}
		list = append(list, elem)
	}
	if p.peekToken.Type != end {
		p.addError(p.peekToken.Line, "expected %q in expression list, got %q", end, p.peekToken.Literal)
		return nil
	}
	p.nextToken()
	return list
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLiteral{Token: p.curToken}
	arr.Elements = p.parseExpressionList(token.RBRACKET)
	if arr.Elements == nil {
		return nil
	}
	return arr
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	expr := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	expr.Index = p.parseExpression(LOWEST)
	if expr.Index == nil {
		return nil
	}
	if p.peekToken.Type != token.RBRACKET {
		p.addError(p.peekToken.Line, "expected `]`, got %q", p.peekToken.Literal)
		return nil
	}
	p.nextToken()
	return expr
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	for p.peekToken.Type != token.RBRACE {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if key == nil {
			return nil
		}
		if p.peekToken.Type != token.COLON {
			p.addError(p.peekToken.Line, "expected `:` after hash key, got %q", p.peekToken.Literal)
			return nil
		}
		p.nextToken()
		p.nextToken()
		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}
		hash.Pairs[key] = value
		if p.peekToken.Type != token.RBRACE && p.peekToken.Type != token.COMMA {
			p.addError(p.peekToken.Line, "expected `,` or `}` in hash literal, got %q", p.peekToken.Literal)
			return nil
		}
		if p.peekToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	if p.peekToken.Type != token.RBRACE {
		p.addError(p.peekToken.Line, "expected `}` to close hash literal, got %q", p.peekToken.Literal)
		return nil
	}
	p.nextToken()
	return hash
}