package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/hiro-lapis/jackanalyzer/tokenizer"
)

var constTokens = []string{ // keyが上記定数の値と対応してるため移動禁止
	"keyword",
	"symbol",
	"integerConstant",
	"stringConstant",
	"identifier",
}

var versionParts int

const maxFileSize = 10485760

// analyzer: top-most module
// Tokenizer: load file and tokenize
// CompilationEngine: compile tokenized data
// Auto grader の実行環境v1.13以下での実行を想定
func main() {
	versionParts, _ = strconv.Atoi(strings.Split(strings.TrimPrefix(runtime.Version(), "go"), ".")[1]) // 1."20".3

	flag.Parse()
	fileName := ""
	if len(flag.Args()) == 0 {
		// TODO remove following code
		// fileName = "./project10/sample.jack"
		fileName = "./project10/ExpressionLessSquare"
		// fileName = "./project10/Square"
		// fileName = "./project10/ArrayTest/Main.jack"
		// fileName = "./project8/ProgramFlow/BasicLoop/BasicLoop.vm"
		// fileName = "StaticTest.vm"
		// fileName = "StaticTest.vm"
		// fileName := "StackTest.vm"
	} else if len(flag.Args()) > 2 {
		fmt.Println("too many arguments. we use only 1st argument")
	}
	if fileName == "" {
		fileName = flag.Args()[0]
	}
	isDir := false
	outPutPath := ""
	outPutFileName := ""
	if !strings.Contains(fileName, ".jack") {
		isDir = true
		outPutPath = fileName
		outPutFileName = fileName[strings.LastIndex(fileName, "/")+1:]
	} else {
		outPutPath = fileName[:strings.LastIndex(fileName, "/")+1]
		outPutFileName = fileName[strings.LastIndex(fileName, "/")+1:]
		outPutFileName = outPutFileName[:strings.LastIndex(outPutFileName, ".")]
	}
	fmt.Println(outPutPath, outPutFileName)
	// コマンドでディレクトリを指定された時は、ディレクトリ内のvmファイルを全て変換する
	if isDir {
		// v1.16前の実行環境を考慮して読み込み
		dir, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer dir.Close()
		files, err := dir.Readdir(-1) // 全取得
		if err != nil {
			log.Fatal(err)
		}

		// golag 1.16以降の実行可能コード
		// files, err := os.ReadDir(fileName)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		// following regulation, call Sys.Init only when reading directory

		// 取得したエントリを表示
		for _, f := range files {
			if !strings.Contains(f.Name(), ".jack") {
				continue
			}
			// project10ではファイル名に応じたxmlファイルを出力
			outPutFileName = f.Name()[:strings.LastIndex(f.Name(), ".")]
			slash := ""
			if string(fileName[len(fileName)-1]) != "/" {
				slash = "/"
			}
			// className := strings.Split(f.Name(), ".")[0]
			if err = compile(fileName+slash+f.Name(), outPutFileName); err != nil {
				// if err = compile(c, fileName+slash+f.Name(), className); err != nil {
				fmt.Println(err)
			}
		}
	} else {
		// className := strings.Split(fileName, "/")[0]
		if err := compile(fileName, outPutFileName); err != nil {
			// if err := compile(c, fileName, className); err != nil {
			fmt.Println(err)
		}
	}
	// c.CloseFile(outPutPath, outPutFileName)
}

func compile(readPath, outPutFileName string) error {
	c, err := NewCompilationEngine(readPath)
	if err != nil {
		fmt.Println("file open eror")
	}
	c.Compile()
	outputDir := readPath[:strings.LastIndex(readPath, "/")]
	// TODO: テストファイル上書き回避のためのoutputDir設定を改修
	c.CloseFile(outputDir, outPutFileName)
	// c.CloseFile(outputDir+"/output", outPutFileName)
	return nil
}

type CompilationEngine struct {
	t        *tokenizer.Tokenizer
	xmlLines []string
	// 解析中のルールをスタック形式で保持
	parseStack []string
}

func NewCompilationEngine(fileName string) (*CompilationEngine, error) {
	t, err := tokenizer.NewTokenizer(fileName)
	if err != nil {
		return nil, err
	}
	return &CompilationEngine{
		t:          t,
		xmlLines:   make([]string, 0),
		parseStack: make([]string, 0),
	}, nil
}

func (c *CompilationEngine) Compile() error {
	c.CompileClass()
	return nil
}

func (c *CompilationEngine) CompileStatements() error {
	c.CompileNonTerminalOpenTag("statements")
	for {
		if c.t.CurrentToken() == tokenizer.KEY_CLASS {
			err := c.CompileClass()
			if err != nil {
				return err
			}
			continue
		}
		if c.t.CurrentToken() == tokenizer.KEY_IF {
			err := c.CompileIfStatement()
			if err != nil {
				return err
			}
			continue
		}
		if c.t.CurrentToken() == tokenizer.KEY_LET {
			err := c.CompileLetStatement()
			if err != nil {
				return err
			}
			continue
		}
		if c.t.CurrentToken() == tokenizer.KEY_WHILE {
			c.CompileWhileStatement()
			continue
		}
		if c.t.CurrentToken() == tokenizer.KEY_DO {
			c.CompileDoStatement()
			continue
		}
		if c.t.CurrentToken() == tokenizer.KEY_RETURN {
			c.CompileReturnStatement()
			continue
		}
		break
	}
	c.CompileNonTerminalCloseTag()
	return nil
}

// term (op term)*
func (c *CompilationEngine) CompileExpression() error {
	if !c.IsTerminalToken() {
		return nil
	}
	c.CompileNonTerminalOpenTag("expression")
	c.CompileTerm()
	// 2つ目以降のtermがある場合
	for c.isOp() {
		c.compileCT()
		c.CompileTerm()
	}

	c.CompileNonTerminalCloseTag()
	return nil
}

// fileNane
func (c *CompilationEngine) CloseFile(filePath, fileName string) {
	xml := c.xmlLines
	dir := filePath
	if (string)(dir[len(dir)-1]) != "/" {
		dir += "/"
	}
	extentionIdx := strings.LastIndex(fileName, ".")
	outputFileName := fileName
	if extentionIdx != -1 {
		outputFileName = fileName[:strings.LastIndex(fileName, ".")] + ".xml"
	} else {
		outputFileName = fileName + ".xml"
	}

	file, err := os.Create(dir + outputFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	// 書き出しと末尾に <tokens> </tokens> タグを追加
	// w.WriteString("<tokens>\n")
	for i := 0; i < len(xml); i++ {
		w.WriteString(xml[i] + "\n")
	}
	// w.WriteString("</tokens>\n")
	w.Flush()
	fmt.Println("output file: ", dir+outputFileName)
}

// 0. entry of compile jack file, class
func (c *CompilationEngine) CompileClass() error {
	c.CompileNonTerminalOpenTag("class")
	c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD, tokenizer.KEY_CLASS) // class
	c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER)                   // className
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "{")                  // {

	if c.isClassVarDec() {
		c.CompileClassVarDec()
	}
	if c.isSubRoutineDec() {
		c.CompileSubroutineDec()
	}
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "}") // {
	c.CompileNonTerminalCloseTag()
	return nil
}

// 1. five base token compile functions

func (c *CompilationEngine) CompileStringConst() {
	if c.t.TokenType() != tokenizer.T_STR_CONST {
		return
	}
	s := c.t.CurrentToken()
	// 先頭と末尾の"を除去
	str := s[1 : len(s)-1]
	// 一部の文字はエスケープ
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, "\"", "&quot;", -1)

	// v1.12以降ならこっち
	// str = strings.ReplaceAll(str, "<", "&lt;")
	// str = strings.ReplaceAll(str, ">", "&gt;")
	// str = strings.ReplaceAll(str, "&", "&amp;")
	// str = strings.ReplaceAll(str, "\"", "&quot;")

	c.xmlLines = append(c.xmlLines, "<stringConstant> "+str+" </stringConstant>")
}
func (c *CompilationEngine) CompileSymbol() {
	if c.t.TokenType() != tokenizer.T_SYMBOL {
		return
	}
	tt := c.t.TokenType()
	sym := c.t.CurrentToken()
	// 一部の文字はエスケープ
	var escapes = map[string]string{
		"<":  "&lt;",
		">":  "&gt;",
		"&":  "&amp;",
		"\"": "&quot;",
	}
	for k, v := range escapes {
		if k == sym {
			sym = v
			break
		}
	}

	c.xmlLines = append(c.xmlLines, "<"+constTokens[tt]+"> "+sym+" </"+constTokens[tt]+">")
}

// ~1. base token compile function

// 3. non-terminal tag
func (c *CompilationEngine) CompileNonTerminalOpenTag(tag string) {
	c.parseStack = append(c.parseStack, tag)
	c.xmlLines = append(c.xmlLines, "<"+tag+">")
}
func (c *CompilationEngine) CompileNonTerminalCloseTag() {
	endTag := c.parseStack[len(c.parseStack)-1]
	// stackからtagをpop
	c.parseStack = c.parseStack[:len(c.parseStack)-1]
	c.xmlLines = append(c.xmlLines, "</"+endTag+">")
}

// 2. compile non-terminal token
// term = integerConstant | stringConstant | keywordConstant | varName | varName '[' expression ']' | subroutineCall | '(' expression ')' | unaryOp term
func (c *CompilationEngine) CompileTerm() {
	t := c.t.CurrentToken()
	tt := c.t.TokenType()
	c.CompileNonTerminalOpenTag("term")
	// token種別に素直にコンパイルしていいパターン
	if c.isKeyWordConst() || tt == tokenizer.T_STR_CONST || tt == tokenizer.T_INT_CONST {
		c.compileCT()
	} else if t == "-" || t == "~" {
		// unary operator term
		c.compileCT()
		// 再帰
		c.CompileTerm()
	} else if tt == tokenizer.T_SYMBOL && t == "(" {
		// (expression)
		c.compileCT()
		c.CompileExpression()
		c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ")")
	} else if tt == tokenizer.T_IDENTIFIER {
		// varName | varName '[' expression ']' | subroutineCall
		c.compileCT()
		if c.t.CurrentToken() == "[" {
			// varName '[' expression ']'
			c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "[")
			c.CompileExpression()
			c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "]")
		} else if c.t.CurrentToken() == "(" || c.t.CurrentToken() == "." {
			// subroutineCall
			c.CompileSubroutineCall()
		}
	} else {
		fmt.Println("error: token type is not matched")
	}
	c.CompileNonTerminalCloseTag()
}

func (c *CompilationEngine) CompileClassVarDec() {
	// compiles a static declaration or a field declaration
	// static or field => type => varName => ;
	for c.isClassVarDec() {
		c.CompileNonTerminalOpenTag("classVarDec")
		c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD)    // field, static                         // type
		c.compileCT()                                               // type
		c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER) // name
		sep := c.isDelimiter()
		c.compileCT() // symbol( , ; )
		// 同型列挙の処理
		for sep {
			c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER) // name
			c.compileCT()                                               // symbol( , ; )
			sep = c.isDelimiter()
			// c.compileCT() // symbol( , ; )
		}

		c.CompileNonTerminalCloseTag()
	}
}

// type varName (, varName)*; 変数/プロパティ型定義のコンパイル
func (c *CompilationEngine) CompileVarDec() {
	// compiles a var declaration
	c.CompileNonTerminalOpenTag("varDec")
	// var
	// TODO:
	c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD, tokenizer.KEY_VAR)
	for c.isType() {
		// type
		c.compileCT()
		// name
		c.compileCT()
		sep := c.isDelimiter()
		c.compileCT()
		for sep {
			c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER) // name
			sep = c.isDelimiter()
			c.compileCT() // , ;
		}
	}

	c.CompileNonTerminalCloseTag()
}

func (c *CompilationEngine) CompileReturnDec() {
	// 型宣言なのでidnetifier含めkeywordとしてコンパイル
	if c.isType() {
		c.compileCT()
		return
	}
	if c.t.CurrentToken() == tokenizer.KEY_VOID {
		c.compileCT()
		return
	}
	fmt.Println("error: return type is not matched")
}
func (c *CompilationEngine) CompileSubroutineDec() error {
	// compiles a complete method, function, or constructor
	for c.isSubRoutineDec() {
		c.CompileNonTerminalOpenTag("subroutineDec")

		// constructor, function, method
		c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD)
		// return type (void:keyword or type:identifier)
		// jackは関数名の前に戻り値を記載. 戻り値ない時も常にvoidを付ける必要がある
		c.CompileReturnDec()
		// method name
		c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER)
		// (
		c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "(")
		// parameter list
		c.CompileParameterList()
		// )
		c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ")")
		// subroutine body
		c.CompileSubroutineBody()
		c.CompileNonTerminalCloseTag()
	}
	return nil
}

func (c *CompilationEngine) CompileSubroutineBody() error {
	c.CompileNonTerminalOpenTag("subroutineBody")
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "{")
	// jackでは関数初期に変数定義あり
	for c.isVarDec() {
		c.CompileVarDec()
	}
	// 関数本体はstatementで構成される
	c.CompileStatements()

	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "}")
	c.CompileNonTerminalCloseTag()
	return nil
}
func (c *CompilationEngine) CompileParameterList() error {
	// compiles a (possibly empty) parameter list
	c.CompileNonTerminalOpenTag("parameterList")
	// ()はparameter Listの外側で処理

	for c.isType() {
		// 型宣言なのでidnetifier含めkeywordとしてコンパイル
		c.compileCT()
		c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER) // name
		if c.isDelimiter() {
			c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ",") // 列挙型field separator
		}
	}
	c.CompileNonTerminalCloseTag()
	return nil
}

func (c *CompilationEngine) CompileLetStatement() error {
	// compiles a let statement
	c.CompileNonTerminalOpenTag("letStatement")

	c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD, tokenizer.KEY_LET)
	c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER)

	// 配列の場合を考慮して次のtokenを見てコンパイルを分岐
	if c.t.CurrentToken() == "[" {
		// [expression]
		c.compileCT()
		c.CompileExpression()
		c.compileCT()
	}
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "=")
	c.CompileExpression()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ";")

	c.CompileNonTerminalCloseTag()
	return nil
}

// while (expression) { statements }
// ex. while (x < 5) { do Output.printInt(x); let x = x + 1; }
func (c *CompilationEngine) CompileWhileStatement() {
	// compiles a while statement
	c.CompileNonTerminalOpenTag("whileStatement")
	c.compileCT()

	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "(")
	// 条件部分はexpression
	c.CompileExpression()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ")")
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "{")
	c.CompileStatements()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "}")
	c.CompileNonTerminalCloseTag()

}
func (c *CompilationEngine) CompileIfStatement() error {
	// compiles an if statement, possibly with a trailing else clause
	c.CompileNonTerminalOpenTag("ifStatement")
	c.compileCT()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "(")
	c.CompileExpression()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ")")
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "{")
	c.CompileStatements()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "}")
	if c.t.CurrentToken() == tokenizer.KEY_ELSE {
		c.compileCT()
		c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "{")
		c.CompileStatements()
		c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "}")
	}
	c.CompileNonTerminalCloseTag()
	return nil
}

// do subroutineCall;
// ex. do Output.printString("Hello, World!");
func (c *CompilationEngine) CompileDoStatement() {
	// compiles a do statement
	c.CompileNonTerminalOpenTag("doStatement")
	c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD, tokenizer.KEY_DO)
	c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER)
	c.CompileSubroutineCall()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ";")
	c.CompileNonTerminalCloseTag()
}
func (c *CompilationEngine) CompileReturnStatement() {
	// compiles a do statement
	c.CompileNonTerminalOpenTag("returnStatement")
	c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD, tokenizer.KEY_RETURN)
	c.CompileExpression()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ";")
	c.CompileNonTerminalCloseTag()
}

func (c *CompilationEngine) CompileExpressionList() {
	// compiles a  comma-separated list of expressions
	c.CompileNonTerminalOpenTag("expressionList")
	c.CompileExpression()
	for c.isDelimiter() {
		c.compileCT()
		c.CompileExpression()
	}
	c.CompileNonTerminalCloseTag()

}

// hoge'.fuga()' or 'hoge()' のように、関数実行するidentifierは呼び出し元でコンパイルする
func (c *CompilationEngine) CompileSubroutineCall() {
	// compiles a subroutine call expression
	// subroutineCall のパターンを網羅的にコンパイル
	// subroutineCall というnon terminal tagはない
	if c.t.CurrentToken() == "." {
		c.compileCT()
		c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER)
	}
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "(")
	c.CompileExpressionList()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ")")
}

// 4. compileロジックのための補助関数
// tokenizerの現在参照中のtokenをtokentypeそのままXMLに書き込み, 次のtokenへ参照を移す
// 戻り値で書き込みに使った token value, typeを返す
func (c *CompilationEngine) compileCT() (string, TokenType) {
	t := c.t.CurrentToken()
	tokenType := c.t.TokenType()
	if tokenType == tokenizer.T_STR_CONST {
		c.CompileStringConst()
	} else if tokenType == tokenizer.T_SYMBOL {
		c.CompileSymbol()
	} else {
		c.xmlLines = append(c.xmlLines, "<"+constTokens[tokenType]+"> "+t+" </"+constTokens[tokenType]+">")
	}
	c.t.Advance()
	return t, tokenType
}

// 期待値を引数に取り、compileCT で書き込んだ値と比較
func (c *CompilationEngine) processGrammaticallyExpectedToken(t tokenizer.TokenType, v ...string) string {
	token, tt := c.compileCT()
	if len(v) > 0 && token != v[0] {
		fmt.Println("error:The value of ", v[0], " is expected")
	}
	if tt != t {
		fmt.Println("error:The token type of ", t, " is expected")
	}
	return token
}

// 期待されるtoken type, valueと一致するか確認
func (c *CompilationEngine) isToken(t tokenizer.TokenType, s ...string) bool {
	token := c.t.CurrentToken()
	tt := c.t.TokenType()
	if s != nil {
		return (t == tt && token == s[0])
	}
	return t == tt
}

// terminal token判定関数
// int_const, str_const, keyword, 一部symbol((,-,~), identifier(varName,subroutineCall)
// keyword は trueなど値を持つtokenのみtrue
// symbol は(, -, ~ など値を表現するtokenのみtrue
func (c *CompilationEngine) IsTerminalToken() bool {
	t := c.t.CurrentToken()
	tt := c.t.TokenType()
	if tt == tokenizer.T_INT_CONST || tt == tokenizer.T_STR_CONST || c.isKeyWordConst() || tt == tokenizer.T_IDENTIFIER {
		return true
	}
	if t == "(" || t == "-" || t == "~" {
		return true
	}
	return false
}

func (c *CompilationEngine) isVarDec() bool {
	// classのproperty定義先頭keywordを元に判定
	return c.t.CurrentToken() == tokenizer.KEY_VAR
}
func (c *CompilationEngine) isClassVarDec() bool {
	// classのproperty定義先頭keywordを元に判定
	return c.t.TokenType() == tokenizer.T_KEYWORD && (c.t.CurrentToken() == tokenizer.KEY_STATIC || c.t.CurrentToken() == tokenizer.KEY_FIELD)
}
func (c *CompilationEngine) isType() bool {
	// identifier = class名(String, Array含)による型定義と判定
	if c.t.TokenType() == tokenizer.T_IDENTIFIER {
		return true
	}
	t := c.t.CurrentToken()
	// primitive type
	return c.t.TokenType() == tokenizer.T_KEYWORD && (t == tokenizer.KEY_INT || t == tokenizer.KEY_CHAR || t == tokenizer.KEY_BOOLEAN)
}
func (c *CompilationEngine) isDelimiter() bool {
	return c.t.CurrentToken() == ","
}
func (c *CompilationEngine) isOp() bool {
	var ops = []string{"+", "-", "*", "/", "&", "|", "<", ">", "="}
	for _, op := range ops {
		if c.t.CurrentToken() == op {
			return true
		}
	}
	return false
}

func (c *CompilationEngine) isKeyWordConst() bool {
	var keyWordConsts = []string{tokenizer.KEY_TRUE, tokenizer.KEY_FALSE, tokenizer.KEY_NULL, tokenizer.KEY_THIS}
	for _, keyWordConst := range keyWordConsts {
		if c.t.CurrentToken() == keyWordConst {
			return true
		}
	}
	return false
}
func (c *CompilationEngine) isSubRoutineDec() bool {
	// classのsubRoutine(関数)の定義先頭keywordを元に判定
	return c.t.TokenType() == tokenizer.T_KEYWORD && (c.t.CurrentToken() == tokenizer.KEY_CONSTRUCTOR || c.t.CurrentToken() == tokenizer.KEY_CONSTRUCTOR || c.t.CurrentToken() == tokenizer.KEY_METHOD)
}
