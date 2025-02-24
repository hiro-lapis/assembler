package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/hiro-lapis/jackanalyzer/symboltable"
	"github.com/hiro-lapis/jackanalyzer/tokenizer"
	"github.com/hiro-lapis/jackanalyzer/vmwriter"
)

var constTokens = []string{ // keyが上記定数の値と対応してるため移動禁止
	"keyword",
	"symbol",
	"integerConstant",
	"stringConstant",
	"identifier",
}

// - var declaration in subroutine
// - pass  multiple argument  when call function
// - register argument var into symbol table
// - let statement
// - unary operation
// - while statement
// - if and else statement
// - handle &
// - handle >

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
		// fileName = "./project11/Seven"
		fileName = "./project11/ConvertToBin"
		// fileName = "./project11/ComplexArrays"
		// fileName = "./project11/Average"
		// fileName = "./project11/Pong"
		// fileName = "./project11/Square"
	} else if len(flag.Args()) > 2 {
		fmt.Println("too many arguments. we use only 1st argument")
	}
	if fileName == "" {
		fileName = flag.Args()[0]
	}
	isDir := false
	outPutBasePath := ""
	outPutFileName := ""
	if strings.Contains(fileName, ".jack") {
		// dir/fileName
		outPutFileName = fileName[:strings.LastIndex(fileName, ".jack")]
	} else {
		isDir = true
		outPutBasePath = fileName
	}
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
			outPutFileName = outPutBasePath + "/" + f.Name()[:strings.LastIndex(f.Name(), ".")]
			if string(fileName[len(fileName)-1]) != "/" {
				fileName += "/"
			}
			if err = compile(fileName+f.Name(), outPutFileName); err != nil {
				// if err = compile(c, fileName+slash+f.Name(), className); err != nil {
				fmt.Println(err)
			}
		}
	} else {
		if err := compile(fileName, outPutFileName); err != nil {
			fmt.Println(err)
		}
	}
}

func compile(readPath, outPutFileName string) error {
	c, err := NewCompilationEngine(readPath, outPutFileName)
	if err != nil {
		fmt.Println("file open eror")
	}
	c.Compile()
	c.writer.Close()
	return nil
}

type CompilationEngine struct {
	t        *tokenizer.Tokenizer
	xmlLines []string
	// 解析中のルールをスタック形式で保持
	parseStack []string
	writer     *vmwriter.VmWriter
	st         *symboltable.SymbolTable
	// class name of compiling file
	cName string
	// function name of compiling file
	fName string
	// return type of compiling subroutine
	rType string
}

func (c *CompilationEngine) setClassName(n string) {
	c.cName = n
}

func NewCompilationEngine(readPath, outputFileName string) (*CompilationEngine, error) {
	t, err := tokenizer.NewTokenizer(readPath)
	if err != nil {
		return nil, err
	}
	file, _ := os.Create(outputFileName + ".vm")
	return &CompilationEngine{
		t:          t,
		xmlLines:   make([]string, 0),
		parseStack: make([]string, 0),
		writer:     vmwriter.NewVmWriter(file),
		st:         symboltable.NewSymbolTable(),
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
	// term
	c.CompileTerm()
	// 2つ目以降のtermがある場合
	for c.isOp() {
		// op term op term...
		op, _ := c.compileCT()
		c.CompileTerm()

		// VM: VM stack での演算は post fix なのでterm push の後にコンパイル
		c.writer.WriteArithmetic(vmwriter.ArtCmds[op])
	}

	c.CompileNonTerminalCloseTag()
	return nil
}

// 0. entry of compile jack file, class
func (c *CompilationEngine) CompileClass() error {
	c.CompileNonTerminalOpenTag("class")
	c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD, tokenizer.KEY_CLASS) // class
	cName := c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER)          // className
	// 現在コンパイル中のクラス名を登録
	c.setClassName(cName)
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "{") // {

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
	if tt == tokenizer.T_INT_CONST {
		c.compileCT()
		// VM use constant to write number on VM file
		num, _ := strconv.Atoi(t)
		c.writer.WritePush(vmwriter.CONSTANT, num)
	} else if c.isKeyWordConst() || tt == tokenizer.T_STR_CONST {
		c.compileCT()
	} else if t == "-" || t == "~" {
		// VM: compile  unary term postfix order.
		//-1 => ex. push const 1 , neg
		uop, _ := c.compileCT()
		// op によって演算される項を先にpush. 再帰から戻ったら uop の演算を出力
		c.CompileTerm()
		c.writer.WriteArithmetic(vmwriter.UnaryCmds[uop])
	} else if tt == tokenizer.T_SYMBOL && t == "(" { // (2 * 3)
		// (expression)
		c.compileCT()         // (
		c.CompileExpression() // 2, 3 will be compiled in `if tt == tokenizer.T_INT_CONST`
		c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ")")
	} else if tt == tokenizer.T_IDENTIFIER {
		// varName | varName '[' expression ']' | subroutineCall
		nIdentifier, _ := c.compileCT()
		if c.t.CurrentToken() == "[" {
			// varName '[' expression ']'
			c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "[")
			c.CompileExpression()
			c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "]")
		} else if c.t.CurrentToken() == "(" || c.t.CurrentToken() == "." {
			// subroutineCall
			fName, count := c.CompileSubroutineCall()
			// VM: CompileSubroutineCall で引数を stack にpush した後に function call をコンパイル
			c.writer.WriteCall(nIdentifier+"."+fName, count)
		} else {
			// VM: varName単体
			k, _ := c.st.KindOf(nIdentifier)
			i, _ := c.st.IndexOf(nIdentifier)
			c.writer.WritePush(k, i)
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
// This func generate any VM code but register local vars on subroutine level symbol table to RAM virtual mapping
func (c *CompilationEngine) CompileVarDec() (count int) {
	// compiles a var declaration
	c.CompileNonTerminalOpenTag("varDec")
	// var
	// TODO:
	c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD, tokenizer.KEY_VAR)
	for c.isType() {
		// type
		vType, _ := c.compileCT()
		// name
		vName, _ := c.compileCT()
		sep := c.isDelimiter()
		// VM: register variable in subroutine level
		c.st.Define(symboltable.SUBROUTINE_LEVEL, vName, vType, vmwriter.LCL)
		count++
		c.compileCT()
		for sep {
			vName = c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER) // name
			sep = c.isDelimiter()
			// VM: register variable in subroutine level
			c.st.Define(symboltable.SUBROUTINE_LEVEL, vName, vType, vmwriter.LCL)
			count++
			c.compileCT() // , ;
		}
	}

	c.CompileNonTerminalCloseTag()
	// VM: return registerd vars' count
	return count
}

func (c *CompilationEngine) CompileReturnDec() {
	// 型宣言なのでidnetifier含めkeywordとしてコンパイル
	if c.isType() {
		rType, _ := c.compileCT()
		c.rType = rType
		return
	}
	if c.t.CurrentToken() == tokenizer.KEY_VOID {
		c.rType = "void"
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
		// VM: in the func, compiler engine save return type for the last compile process of subroutine
		c.CompileReturnDec()
		// method name
		// VM: 関数名を保存し、var dec をよみこんだタイミングでfunc dec コンパイル.return のタイミングで初期化する
		c.fName = c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER)
		// (
		c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "(")
		// parameter list
		c.CompileParameterList()
		// )
		c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ")")

		// subroutine body
		c.CompileSubroutineBody()
		c.CompileNonTerminalCloseTag()
		// VM: reset return type and fName
		c.rType = ""
		c.fName = ""
	}
	return nil
}

func (c *CompilationEngine) CompileSubroutineBody() error {
	c.CompileNonTerminalOpenTag("subroutineBody")
	// TODO: より上層のCompileSubroutineDecでリセットする方がいいかも。要確認
	// VM: reset subroutine level symbol table to compile the function
	c.st.StartSubroutine()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "{")
	// jackでは関数初期に変数定義あり
	aCount := 0
	for c.isVarDec() {
		aCount += c.CompileVarDec()
	}
	// VM: write method declaration
	c.writer.WriteFunction(c.cName+"."+c.fName, aCount)
	// 関数本体はstatementで構成される
	c.CompileStatements()

	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "}")
	c.CompileNonTerminalCloseTag()

	return nil
}

// parameter list をコンパイルしcount を返す
func (c *CompilationEngine) CompileParameterList() (count int) {
	// compiles a (possibly empty) parameter list
	c.CompileNonTerminalOpenTag("parameterList")
	// ()はparameter Listの外側で処理

	for c.isType() {
		// 型宣言なのでidnetifier含めkeywordとしてコンパイル
		vType, _ := c.compileCT()
		vName := c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER) // name
		c.st.Define(symboltable.SUBROUTINE_LEVEL, vName, vType, vmwriter.ARG)
		if c.isDelimiter() {
			c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ",") // 列挙型field separator
		}
		count++
	}
	c.CompileNonTerminalCloseTag()
	return count
}

func (c *CompilationEngine) CompileLetStatement() error {
	// compiles a let statement
	c.CompileNonTerminalOpenTag("letStatement")

	c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD, tokenizer.KEY_LET)
	vName := c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER)
	vSegment, _ := c.st.KindOf(vName)
	vIndex, _ := c.st.IndexOf(vName)

	// 配列の場合を考慮して次のtokenを見てコンパイルを分岐
	if c.t.CurrentToken() == "[" {
		// [expression]
		c.compileCT()
		c.CompileExpression()
		c.compileCT()
	}
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "=")
	c.CompileExpression()
	// VM: assign right side expression's value to the left side var
	c.writer.WritePush(vmwriter.Segment(vSegment), vIndex)
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
	cName := c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER)
	// VM: 引数渡しのpush stash を先にコンパイルする
	fName, count := c.CompileSubroutineCall()
	// VM: 関数呼び出しをコンパイル
	c.writer.WriteCall(cName+"."+fName, count)
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ";")
	// VM: do statement doesn't receive return value, pop return value
	c.writer.WritePop(vmwriter.TEMP, 0)
	c.CompileNonTerminalCloseTag()
}
func (c *CompilationEngine) CompileReturnStatement() {
	// compiles a do statement
	c.CompileNonTerminalOpenTag("returnStatement")
	c.processGrammaticallyExpectedToken(tokenizer.T_KEYWORD, tokenizer.KEY_RETURN)
	// VM: if returnning value is void, push empty value
	if c.rType == "void" && c.t.CurrentToken() == ";" {
		c.writer.WritePush(vmwriter.CONSTANT, 0)
	} else {
		c.writer.WriteReturn()
	}
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ";")
	// VM: add return after pushing return value
	c.writer.WriteReturn()
	c.CompileNonTerminalCloseTag()
}

// compile term op term
// return count of compiled term
func (c *CompilationEngine) CompileExpressionList() (count int) {
	// compiles a  comma-separated list of expressions
	c.CompileNonTerminalOpenTag("expressionList")

	c.CompileExpression()
	count++
	for c.isDelimiter() {
		c.compileCT()
		c.CompileExpression()
		count++
	}
	c.CompileNonTerminalCloseTag()
	return count
}

// hoge'.fuga(xxx)' or 'hoge(xxx)' のように、関数実行するidentifierは呼び出し元でコンパイルする
//
//	return function name and count of argument in the part of xxx
func (c *CompilationEngine) CompileSubroutineCall() (fName string, count int) {
	// compiles a subroutine call expression
	// subroutineCall のパターンを網羅的にコンパイル
	// subroutineCall というnon terminal tagはない
	if c.t.CurrentToken() == "." {
		c.compileCT()
		fName = c.processGrammaticallyExpectedToken(tokenizer.T_IDENTIFIER)
	}
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, "(")
	count = c.CompileExpressionList()
	c.processGrammaticallyExpectedToken(tokenizer.T_SYMBOL, ")")
	return fName, count
}

// 4. compileロジックのための補助関数
// tokenizerの現在参照中のtokenをtokentypeそのままXMLに書き込み, 次のtokenへ参照を移す
// 戻り値で書き込みに使った token value, typeを返す
func (c *CompilationEngine) compileCT() (string, tokenizer.TokenType) {
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
	return c.t.TokenType() == tokenizer.T_KEYWORD && (c.t.CurrentToken() == tokenizer.KEY_FUNCTION || c.t.CurrentToken() == tokenizer.KEY_CONSTRUCTOR || c.t.CurrentToken() == tokenizer.KEY_METHOD)
}
