package parser

import (
	"fucklang/internal/ast"
	"fucklang/internal/lexer"
	"fucklang/internal/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // == != < > <= >=
	SUM         // + -
	PRODUCT     // * /
	PREFIX      // -x !x
	CALL        // fn(x)
	INDEX       // arr[i]
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
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
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
	case token.ASS:
		return p.parseReturnStatement()
	case token.LBRACE:
		return p.parseBlockStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	stmt := &ast.PrintStatement{Token: p.curToken}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseAssignStatement() ast.Statement {
	stmt := &ast.AssignStatement{Token: p.curToken}
	p.nextToken()
	if p.curToken.Type != token.IDENT {
		return nil
	}
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	// shit arr[0] = ... 形式
	if p.curToken.Type == token.LBRACKET {
		p.nextToken()
		idx := p.parseExpression(LOWEST)
		if p.peekToken.Type == token.RBRACKET {
			p.nextToken()
		}
		p.nextToken()
		if p.curToken.Type != token.ASSIGN {
			return nil
		}
		p.nextToken()
		val := p.parseExpression(LOWEST)
		return &ast.IndexAssignStatement{
			Token: stmt.Token,
			Left:  ident,
			Index: idx,
			Value: val,
		}
	}

	if p.curToken.Type != token.ASSIGN {
		return nil
	}
	p.nextToken()
	stmt.Name = ident
	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseLoopStatement() *ast.LoopStatement {
	stmt := &ast.LoopStatement{Token: p.curToken}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if p.peekToken.Type == token.LBRACE {
		p.nextToken()
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	p.nextToken()
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	var left ast.Expression

	switch p.curToken.Type {
	case token.INT:
		val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
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
	default:
		return nil
	}

	for precedence < p.peekPrecedence() {
		p.nextToken()
		left = p.parseInfixExpression(left)
	}
	return left
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	expr := p.parseExpression(LOWEST)
	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
	}
	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	switch p.curToken.Type {
	case token.LPAREN:
		return p.parseCallExpression(left)
	case token.LBRACKET:
		return p.parseIndexExpression(left)
	}
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)
	return expr
}

func (p *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{Token: p.curToken}
	p.nextToken()
	expr.Condition = p.parseExpression(LOWEST)
	if p.peekToken.Type == token.LBRACE {
		p.nextToken()
	}
	expr.Consequence = p.parseBlockStatement()
	if p.peekToken.Type == token.ELSE {
		p.nextToken()
		p.nextToken()
		if p.curToken.Type == token.LBRACE {
			expr.Alternative = p.parseBlockStatement()
		}
	}
	return expr
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curToken}
	p.nextToken()
	if p.curToken.Type != token.LPAREN {
		return nil
	}
	fn.Parameters = p.parseFunctionParameters()
	if p.peekToken.Type == token.LBRACE {
		p.nextToken()
	}
	fn.Body = p.parseBlockStatement()
	return fn
}

func (p *Parser) parseFunctionParameters() []*ast.Parameter {
	var params []*ast.Parameter
	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return params
	}
	p.nextToken()
	param := &ast.Parameter{Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}}
	params = append(params, param)
	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		param := &ast.Parameter{Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}}
		params = append(params, param)
	}
	if p.peekToken.Type != token.RPAREN {
		return nil
	}
	p.nextToken()
	return params
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	expr := &ast.CallExpression{Token: p.curToken, Function: function}
	expr.Arguments = p.parseExpressionList(token.RPAREN)
	return expr
}

func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	var list []ast.Expression
	if p.peekToken.Type == end {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if p.peekToken.Type != end {
		return nil
	}
	p.nextToken()
	return list
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLiteral{Token: p.curToken}
	arr.Elements = p.parseExpressionList(token.RBRACKET)
	return arr
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	expr := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	expr.Index = p.parseExpression(LOWEST)
	if p.peekToken.Type == token.RBRACKET {
		p.nextToken()
	}
	return expr
}