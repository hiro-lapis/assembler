package tokenizer

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

const (
	KEY_CLASS       = "class"
	KEY_CONSTRUCTOR = "constructor"
	KEY_FUNCTION    = "function"
	KEY_METHOD      = "method"
	KEY_FIELD       = "field"
	KEY_STATIC      = "static"
	KEY_VAR         = "var"
	KEY_INT         = "int"
	KEY_CHAR        = "char"
	KEY_BOOLEAN     = "boolean"
	KEY_VOID        = "void"
	KEY_TRUE        = "true"
	KEY_FALSE       = "false"
	KEY_NULL        = "null"
	KEY_THIS        = "this"
	KEY_LET         = "let"
	KEY_DO          = "do"
	KEY_IF          = "if"
	KEY_ELSE        = "else"
	KEY_WHILE       = "while"
	KEY_RETURN      = "return"
)

// golang はsliceで定数定義できないのでvar定義する
var constKeywords = []string{
	KEY_CLASS,
	KEY_CONSTRUCTOR,
	KEY_FUNCTION,
	KEY_METHOD,
	KEY_FIELD,
	KEY_STATIC,
	KEY_VAR,
	KEY_INT,
	KEY_CHAR,
	KEY_BOOLEAN,
	KEY_VOID,
	KEY_TRUE,
	KEY_FALSE,
	KEY_NULL,
	KEY_THIS,
	KEY_LET,
	KEY_DO,
	KEY_IF,
	KEY_ELSE,
	KEY_WHILE,
	KEY_RETURN,
}

const symbols = "{}()[].,;+-*/&|<>=~"

type TokenType int

const (
	T_KEYWORD    TokenType = iota // keyword
	T_SYMBOL                      // symbol
	T_INT_CONST                   // integer constant
	T_STR_CONST                   // string constant
	T_IDENTIFIER                  // identifier
)

type Token struct {
	tokenType TokenType
	v         string
}
type Tokenizer struct {
	tokens  []Token
	current int
	lexer   *Lexer
}

func NewTokenizer(fileName string) (*Tokenizer, error) {
	// v1.0を想定
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	cReg := regexp.MustCompile(commentReg)
	formatted := cReg.ReplaceAll(file, []byte(""))
	lines := strings.Split(string(formatted), "\n")

	list := make([]string, 0)
	for _, line := range lines {
		v := strings.TrimSpace(line)
		// 複数行コメント含め取り除く
		v = cReg.ReplaceAllString(v, "")
		if len(v) == 0 {
			continue
		}
		list = append(list, v)
	}
	t := &Tokenizer{lexer: NewLexer()}
	t.tokenize(list)
	return t, nil
}

var commentReg = `//.*|/\*[\s\S]*?\*/`

// var commentReg = `//[^\n]*|/\*.*?\*/`

func (t *Tokenizer) tokenize(lines []string) {
	for _, line := range lines {
		// 定数tokenとしてマッチする単位でletterをslice化
		words := t.lexer.Split(line)
		for _, word := range words {
			// letterにidentifier が含まれる場合を考慮してさらに分割
			if t.lexer.IsKeyWord(word) {
				t.tokens = append(t.tokens, Token{T_KEYWORD, word})
			} else if t.lexer.IsSymbol(word) {
				t.tokens = append(t.tokens, Token{T_SYMBOL, word})
			} else if t.lexer.IsNum(word) {
				t.tokens = append(t.tokens, Token{T_INT_CONST, word})
			} else if t.lexer.IsStr(word) {
				t.tokens = append(t.tokens, Token{T_STR_CONST, word})
			} else if t.lexer.IsId(word) {
				t.tokens = append(t.tokens, Token{T_IDENTIFIER, word})
			}
		}
	}
}

func (t *Tokenizer) HasMoreToken() bool {
	maxIdx := len(t.tokens) - 1
	return maxIdx > t.current
}

func (t *Tokenizer) Advance() {
	if t.HasMoreToken() {
		t.current++
	}
}

func (t *Tokenizer) TokenType() TokenType {
	return t.tokens[t.current].tokenType
}

func (t *Tokenizer) CurrentToken() string {
	return t.tokens[t.current].v
}

type Lexer struct {
	keywordReg *regexp.Regexp
	symbolReg  *regexp.Regexp
	numReg     *regexp.Regexp
	strReg     *regexp.Regexp
	idReg      *regexp.Regexp
	wordReg    *regexp.Regexp
}

// https://regex101.com/r/1J9Z8v/1
// https://chatgpt.com/share/67a8b629-ac3c-8011-a679-673eff10c172
func NewLexer() *Lexer {
	// 前方一致で部分ミスマッチ回避
	keywordRe := `^(?:` + strings.Join(constKeywords, `|`) + `)`
	// 1文字づつエスケープ
	var symReList []string
	for _, s := range symbols {
		symReList = append(symReList, regexp.QuoteMeta(string(s)))
	}
	// |区切り
	var symRe = `(?:` + strings.Join(symReList, `|`) + `)`
	// ↓だと1部しかエスケープされないので注意
	// symRe := `(?:` + regexp.QuoteMeta(symbols) + `)`
	numRe := `\d+`
	strRe := `"[^"\n]*"`
	idRe := `[a-zA-Z_][a-zA-Z0-9_]*`
	wordRe := regexp.MustCompile(keywordRe + `|` + symRe + `|` + numRe + `|` + strRe + `|` + idRe)

	return &Lexer{
		keywordReg: regexp.MustCompile(keywordRe),
		symbolReg:  regexp.MustCompile(symRe),
		numReg:     regexp.MustCompile(numRe),
		strReg:     regexp.MustCompile(strRe),
		idReg:      regexp.MustCompile(idRe),
		wordReg:    wordRe,
	}
}

// r=regexp.MustCompile(`p([a-z]+)ch`)
// fmt.Println(r.FindAllString("peach punch pinch", -1))
// ["peach" "punch" "pinch"]
func (l *Lexer) Split(line string) []string {
	return l.wordReg.FindAllString(line, -1)
}
func (l *Lexer) IsKeyWord(line string) bool {
	return l.keywordReg.MatchString(line)
}
func (l *Lexer) IsSymbol(line string) bool {
	return l.symbolReg.MatchString(line)
}
func (l *Lexer) IsNum(line string) bool {
	return l.numReg.MatchString(line)
}
func (l *Lexer) IsStr(line string) bool {
	return l.strReg.MatchString(line)
}
func (l *Lexer) IsId(line string) bool {
	return l.idReg.MatchString(line)
}
