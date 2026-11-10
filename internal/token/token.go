package token

type Type string

const (
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"

	FUCK  Type = "FUCK"
	SHIT  Type = "SHIT"
	DAMN  Type = "DAMN"
	ELSE  Type = "ELSE"
	HELL  Type = "HELL"
	BREAK Type = "BREAK"
	BITCH Type = "BITCH"
	ASS   Type = "ASS"

	WTF        Type = "WTF"
	FK         Type = "FK"
	ORGY       Type = "ORGY"
	BULLSHIT   Type = "BULLSHIT"
	NOBULLSHIT Type = "NOBULLSHIT"

	// 命名空间关键字
	STR  Type = "STR"
	ARR  Type = "ARR"
	FILE Type = "FILE"

	IDENT  Type = "IDENT"
	INT    Type = "INT"
	STRING Type = "STRING"

	ASSIGN Type = "="
	PLUS   Type = "+"
	MINUS  Type = "-"
	STAR   Type = "*"
	SLASH  Type = "/"
	BANG   Type = "!"
	LT     Type = "<"
	GT     Type = ">"
	EQ     Type = "=="
	NOT_EQ Type = "!="
	LTE    Type = "<="
	GTE    Type = ">="

	PLUS_ASSIGN  Type = "+="
	MINUS_ASSIGN Type = "-="
	STAR_ASSIGN  Type = "*="
	SLASH_ASSIGN Type = "/="

	COMMA     Type = ","
	SEMICOLON Type = ";"
	COLON     Type = ":"
	DOT       Type = "."

	LPAREN   Type = "("
	RPAREN   Type = ")"
	LBRACE   Type = "{"
	RBRACE   Type = "}"
	LBRACKET Type = "["
	RBRACKET Type = "]"

	TRUE  Type = "TRUE"
	FALSE Type = "FALSE"
)

type Token struct {
	Type    Type
	Literal string
	Line    int
}

var keywords = map[string]Type{
	"fuck":  FUCK,
	"shit":  SHIT,
	"damn":  DAMN,
	"else":  ELSE,
	"hell":  HELL,
	"break": BREAK,
	"bitch": BITCH,
	"ass":   ASS,
	"true":  TRUE,
	"false": FALSE,

	"wtf":        WTF,
	"fk":         FK,
	"orgy":       ORGY,
	"bullshit":   BULLSHIT,
	"nobullshit": NOBULLSHIT,

	"str":  STR,
	"arr":  ARR,
	"file": FILE,
}

func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}