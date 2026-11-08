package lexer

import "fucklang/internal/token"

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
	line    int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	if l.ch == '\n' {
		l.line++
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespaceAndComments()

	line := l.line

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch), Line: line}
		} else {
			tok = token.Token{Type: token.ASSIGN, Literal: string(l.ch), Line: line}
		}
	case '+':
		tok = token.Token{Type: token.PLUS, Literal: string(l.ch), Line: line}
	case '-':
		tok = token.Token{Type: token.MINUS, Literal: string(l.ch), Line: line}
	case '*':
		tok = token.Token{Type: token.STAR, Literal: string(l.ch), Line: line}
	case '/':
		tok = token.Token{Type: token.SLASH, Literal: string(l.ch), Line: line}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch), Line: line}
		} else {
			tok = token.Token{Type: token.BANG, Literal: string(l.ch), Line: line}
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: string(ch) + string(l.ch), Line: line}
		} else {
			tok = token.Token{Type: token.LT, Literal: string(l.ch), Line: line}
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: string(ch) + string(l.ch), Line: line}
		} else {
			tok = token.Token{Type: token.GT, Literal: string(l.ch), Line: line}
		}
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: string(l.ch), Line: line}
	case ':':
		tok = token.Token{Type: token.COLON, Literal: string(l.ch), Line: line}
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Literal: string(l.ch), Line: line}
	case '(':
		tok = token.Token{Type: token.LPAREN, Literal: string(l.ch), Line: line}
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: string(l.ch), Line: line}
	case '{':
		tok = token.Token{Type: token.LBRACE, Literal: string(l.ch), Line: line}
	case '}':
		tok = token.Token{Type: token.RBRACE, Literal: string(l.ch), Line: line}
	case '[':
		tok = token.Token{Type: token.LBRACKET, Literal: string(l.ch), Line: line}
	case ']':
		tok = token.Token{Type: token.RBRACKET, Literal: string(l.ch), Line: line}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = line
	case 0:
		tok = token.Token{Type: token.EOF, Literal: "", Line: line}
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Line = line
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			tok.Line = line
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Line: line}
		}
	}

	l.readChar()
	return tok
}

// 跳过空白和 // 行注释
func (l *Lexer) skipWhitespaceAndComments() {
	for {
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readChar()
		}
		if l.ch == '/' && l.peekChar() == '/' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			continue
		}
		break
	}
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readNumber() string {
	pos := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

// 支持转义：\" \\ \n \t \r
func (l *Lexer) readString() string {
	var out []byte
	l.readChar() // 跳过起始引号
	for {
		if l.ch == '"' || l.ch == 0 {
			break
		}
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				out = append(out, '\n')
			case 't':
				out = append(out, '\t')
			case 'r':
				out = append(out, '\r')
			case '"':
				out = append(out, '"')
			case '\\':
				out = append(out, '\\')
			default:
				out = append(out, '\\', l.ch)
			}
			l.readChar()
			continue
		}
		out = append(out, l.ch)
		l.readChar()
	}
	return string(out)
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func isLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}