package parser

import "github.com/darwin1224/saphire/token"

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	UNARY
	CALL
	INDEX
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.MOD:      PRODUCT,
	token.ASTERISK: PRODUCT,
	token.POWER:    PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}
