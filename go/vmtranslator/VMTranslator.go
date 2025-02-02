package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	SP    = "@0" // 256
	LCL   = "@1" // 300
	ARG   = "@2" // 400
	THIS  = "@3" // 3000
	THAT  = "@4" // 3010
	TMP   = "@5" // 5-12
	STT   = 16   // 16-255
	POINT = ""
)

type CommandType int

const (
	C_ARITHMETIC CommandType = iota // arithmetic command
	C_PUSH                          // push segment i
	C_POP                           // pop segment i
	C_LABEL                         // define label
	C_GOTO                          // jump to label unconditionally
	C_IF                            // retrive topmost value of stack and jump if it is not 0
	C_FUNCTION                      // function declaration
	C_RETURN                        // return from function
	C_CALL                          // function call
)

const maxFileSize = 10485760

// Auto grader の実行環境v1.13以下での実行を想定
func main() {
	flag.Parse()
	fileName := ""
	if len(flag.Args()) == 0 {
		// fmt.Println("please input file name")
		// return
		// TODO remove following code
		// fileName = "./project8/ProgramFlow/BasicLoop/BasicLoop.vm"
		// fileName = "./project8/ProgramFlow/FibonacciSeries/"
		// fileName = "./project8/FunctionCalls/SimpleFunction"
		// fileName = "./project8/FunctionCalls/NestedCall"
		// fileName = "./project8/FunctionCalls/FibonacciElement"
		// fileName = "./project8/FunctionCalls/StaticsTest"
		// fileName = "StaticTest.vm"
		// fileName := flag.Args()
		// fileName := "SimpleAdd.vm"
		// fileName := "BasicTest.vm"
		// fileName = "PointerTest.vm"
		// fileName := "StackTest.vm"
	} else if len(flag.Args()) > 1 {
		fmt.Println("too many arguments. we use only 1st argument")
	}
	if fileName == "" {
		fileName = flag.Args()[0]
	}
	isDir := false
	outPutPath := ""
	outPutFileName := ""
	if !strings.Contains(fileName, ".vm") {
		isDir = true
		outPutPath = fileName
		outPutFileName = fileName[strings.LastIndex(fileName, "/")+1:]
	} else {
		outPutPath = fileName[:strings.LastIndex(fileName, "/")+1]
		outPutFileName = fileName[strings.LastIndex(fileName, "/")+1:]
		outPutFileName = outPutFileName[:strings.LastIndex(outPutFileName, ".")]
	}
	c := NewCodeWriter()
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
		c.WriteCall("Sys.init", 0)

		// 取得したエントリを表示
		for _, f := range files {
			if !strings.Contains(f.Name(), ".vm") {
				continue
			}
			slash := ""
			if string(fileName[len(fileName)-1]) != "/" {
				slash = "/"
			}
			className := strings.Split(f.Name(), ".")[0]
			if err = compile(c, fileName+slash+f.Name(), className); err != nil {
				fmt.Println(err)
			}
		}
	} else {
		className := strings.Split(fileName, "/")[0]
		if err := compile(c, fileName, className); err != nil {
			fmt.Println(err)
		}
	}
	c.CloseFile(outPutPath, outPutFileName)
}

func compile(c *CodeWriter, filePath, className string) error {
	p, err := NewParser(filePath)
	if err != nil {
		fmt.Println("file open eror")
	}
	c.SetStaticName(className)

	for p.HasMoreCommands() {
		switch p.CommandType() {
		case C_ARITHMETIC:
			c.WriteArithmetic(p.Arg1())
		case C_PUSH:
			arg2, err := p.Arg2()
			if err != nil {
				fmt.Println(err)
				break
			}
			c.WritePushPop(C_PUSH, p.Arg1(), arg2)
		case C_POP:
			arg2, err := p.Arg2()
			if err != nil {
				fmt.Println(err)
				break
			}
			c.WritePushPop(C_POP, p.Arg1(), arg2)
		case C_LABEL:
			c.WriteLabel(p.Arg1())
		case C_IF:
			c.WriteIf(p.Arg1())
		case C_GOTO:
			c.WriteGoto(p.Arg1())
		case C_FUNCTION:
			arg2, err := p.Arg2()
			if err != nil {
				fmt.Println(err)
				break
			}
			c.WriteFunction(p.Arg1(), arg2)
		case C_CALL:
			arg2, err := p.Arg2()
			if err != nil {
				fmt.Println(err)
				break
			}
			c.WriteCall(p.Arg1(), arg2)
		case C_RETURN:
			c.WriteReturn()
		}

		p.advance()
	}
	return nil
}

type Parser struct {
	commands    []string
	currentLine int
}

func NewParser(fileName string) (*Parser, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// grader にて10mb unexpected _485_760 at end of statement とエラー出る(v1.13以下?)ので通常のリテラルに変更
	data := make([]byte, maxFileSize)
	count, err := file.Read(data)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if count == 0 {
		err := fmt.Errorf("error: file is empty")
		return nil, err
	}
	lines := strings.Split(string(data[:count]), "\n")

	list := make([]string, 0)
	for _, line := range lines {
		v := ""
		if len(line) == 0 {
			continue
		}
		index := strings.Index(line, "//")
		// skip comment
		if index == 0 {
			continue
		} else if index != -1 {
			v = line[:index]
		} else {
			v = line
		}
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}
		list = append(list, v)
	}
	return &Parser{commands: list, currentLine: 0}, nil
}

func (p *Parser) HasMoreCommands() bool {
	maxIdx := len(p.commands) - 1
	return maxIdx >= p.currentLine
}

func (p *Parser) advance() {
	if p.HasMoreCommands() {
		p.currentLine++
	}
}

func (p *Parser) CommandType() CommandType {
	cmds := strings.Split(p.commands[p.currentLine], " ")
	cmd := cmds[0]
	switch string(cmd) {
	case "push":
		return C_PUSH
	case "pop":
		return C_POP
	case "label":
		return C_LABEL
	case "if-goto":
		return C_IF
	case "goto":
		return C_GOTO
	case "function":
		return C_FUNCTION
	// TODO: implement call command
	case "call":
		return C_CALL
	case "return":
		return C_RETURN
	default:
		return C_ARITHMETIC
	}
}
func (p *Parser) Arg1() string {
	if p.CommandType() == C_RETURN {
		return ""
	}
	cmds := strings.Split(p.commands[p.currentLine], " ")
	if len(cmds) == 1 {
		return cmds[0]
	}
	return cmds[1]
}
func (p *Parser) Arg2() (int, error) {
	cmds := strings.Split(p.commands[p.currentLine], " ")
	if len(cmds) == 1 {
		return -1, fmt.Errorf("Arg2 should be called only push or pop")
	}
	num, _ := strconv.Atoi(cmds[2])
	return num, nil
}

type CodeWriter struct {
	vmCodes      []string
	staticLabels map[string]string
	className    string
	labelCount   int
	callCount    int
}

func NewCodeWriter() *CodeWriter {
	initAsm := [][]string{
		{"0", "256"},
		// {"1", "300"},
		// {"2", "400"},
		// {"3", "3000"},
		// {"4", "3010"},
		// {"5", "5"},
		// {"6", "16"},
	}
	setPointer := func(setTarget, value string) []string {
		return []string{"@" + value, "D=A", "@" + setTarget, "M=D"}
	}
	asmLines := make([]string, 0)
	for _, l := range initAsm {
		cmds := setPointer(l[0], l[1])
		asmLines = append(asmLines, cmds...)
	}
	// codes順を保つため instance 化後にSys.initを実行
	c := &CodeWriter{
		vmCodes:      asmLines,
		staticLabels: make(map[string]string),
	}
	return c
}
func (c *CodeWriter) SetStaticName(name string) {
	c.className = name
}

func (c *CodeWriter) CloseFile(filePath, fileName string) {
	asmLines := c.AssembledCodes()
	dir := filePath
	if (string)(dir[len(dir)-1]) != "/" {
		dir += "/"
	}
	extentionIdx := strings.LastIndex(fileName, ".")
	outputFileName := ""
	if extentionIdx != -1 {
		outputFileName = fileName[:strings.LastIndex(fileName, ".")] + ".asm"
	} else {
		outputFileName = fileName + ".asm"
	}

	// createFile(dir, outputFileName, asmLines)
	file, err := os.Create(dir + outputFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for i := 0; i < len(asmLines); i++ {
		w.WriteString(asmLines[i] + "\n")
	}
	w.Flush()
	fmt.Println("output file: ", dir+outputFileName)
}

func (c *CodeWriter) WriteArithmetic(cmd string) {
	result := make([]string, 0)
	switch string(cmd) {
	case "add":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "D=M") // D = *SP
		result = append(result, "A=A-1")
		result = append(result, "A=A-1")
		result = append(result, "D=M")
		result = append(result, "A=A+1")
		result = append(result, "D=D+M")
		result = append(result, "A=A-1")
		result = append(result, "M=D")
		result = append(result, SP)
		result = append(result, "M=M-1")
	case "sub":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "A=A-1")
		result = append(result, "D=M")   // x(stack 2つ目の値)
		result = append(result, "A=A+1") // y(stack 1つ目の値)
		result = append(result, "D=D-M") // y - x
		result = append(result, "A=A-1")
		result = append(result, "M=D") // x = y - x(yのメモリは上書き対象)
		result = append(result, SP)
		result = append(result, "M=M-1")
	case "neg":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "D=M") // D = *SP
		result = append(result, "M=-D")
	case "eq":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "A=A-1")
		result = append(result, "D=M") // x
		result = append(result, "A=A+1")
		result = append(result, "D=D-M") // x - y
		result = append(result, c.compare(cmd)...)
	case "gt":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "A=A-1")
		result = append(result, "D=M") // x
		result = append(result, "A=A+1")
		result = append(result, "D=D-M") // x - y
		result = append(result, c.compare(cmd)...)
	case "lt":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "A=A-1")
		result = append(result, "D=M") // x
		result = append(result, "A=A+1")
		result = append(result, "D=D-M") // x - y
		result = append(result, c.compare(cmd)...)
	case "and":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "D=M") // y
		result = append(result, "A=A-1")
		result = append(result, "D=D&M") // x & y
		result = append(result, "M=D")   // x = x & y
		result = append(result, SP)
		result = append(result, "M=M-1")
	case "or":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "D=M") // y
		result = append(result, "A=A-1")
		result = append(result, "D=D|M") // x | y
		result = append(result, "M=D")   // x = x | y
		result = append(result, SP)
		result = append(result, "M=M-1")
	case "not":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "D=M")  // y
		result = append(result, "M=!D") // !y
	}
	c.vmCodes = append(c.vmCodes, result...)
}

func (c *CodeWriter) compare(cmd string) []string {
	l := make([]string, 0)
	// labelはユニークにしないとassembler error 発生するため都度ユニークなappendinx をつける
	// generate unique label
	labelCount := strconv.Itoa(c.labelCount)
	var label, fLabel, jump string
	c.labelCount++
	switch string(cmd) {
	case "eq":
		label = "EQ_TRUE"
		fLabel = "EQ_FALSE"
		jump = "JEQ"
	case "gt":
		label = "GT_TRUE"
		fLabel = "GT_FALSE"
		jump = "JGT"
	case "lt":
		label = "LT_TRUE"
		fLabel = "LT_FALSE"
		jump = "JLT"
	}
	label += "_" + labelCount
	fLabel += "_" + labelCount
	l = append(l, "A=A-1") // y の位置からxへ移動
	// 条件つきjumpの処理は判定の結果ジャンプしなくてもアクセスmemory が変わってしまうためtrue/false両方について-1/0を書く処理が必要
	// trueの場合
	l = append(l, "M=-1") // trueの場合の値は@label設定の前にセットしておく
	l = append(l, "@"+label)
	l = append(l, "D;"+jump)
	// falseの場合
	l = append(l, "@"+fLabel)
	l = append(l, "0;JMP")
	l = append(l, "("+fLabel+")")
	l = append(l, SP)
	l = append(l, "A=M")
	l = append(l, "A=A-1")
	l = append(l, "A=A-1")
	l = append(l, "M=0")
	// trueの場合,jump前のM=-1設定を保持してそのまま進む
	l = append(l, "("+label+")")
	l = append(l, SP)
	l = append(l, "M=M-1")
	return l
}

// get segment pointer from 2nd argument
func (c *CodeWriter) getSecondArgSegment(segment string, index int) string {
	if segment == "pointer" {
		if index == 0 {
			return THIS
		}
		return THAT
	}
	switch segment {
	case "local":
		return LCL
	case "argument":
		return ARG
	case "temp":
		return TMP
	case "this":
		return THIS
	case "that":
		return THAT
	case "static":
		i := strconv.Itoa(index)
		if aInstruction, ok := c.staticLabels[c.className+"."+i]; ok {
			return aInstruction
		}
		nextC := len(c.staticLabels)
		// 16 + exsting count + 1
		ii := strconv.Itoa(STT + nextC)
		c.staticLabels[c.className+"."+i] = "@" + ii
		return "@" + ii
	}

	return ""
}

func (c *CodeWriter) getSecondArgSegmentIndex(segment string, i int) []string {
	l := make([]string, 0)
	if segment != "pointer" && segment != "static" {
		// tempはglobal stack管理の値のため,アドレスeは常に@5~12を使用する
		// staticはtranslator内でマッピング管理しているためindex移動は不要
		if segment != "temp" {
			l = append(l, "A=M")
		}
		idx := i
		for idx > 0 {
			l = append(l, "A=A+1")
			idx--
		}
	}
	return l
}
func (c *CodeWriter) WritePushPop(cmdType CommandType, segment string, i int) {
	result := make([]string, 0)
	index := strconv.Itoa(i)

	switch cmdType {
	case C_PUSH:
		if segment == "constant" {
			result = append(result, "@"+(index))
			result = append(result, "D=A")
		} else {
			result = append(result, c.getSecondArgSegment(segment, i))
			result = append(result, c.getSecondArgSegmentIndex(segment, i)...)
			result = append(result, "D=M")
		}
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "M=D")
		result = append(result, SP)
		result = append(result, "M=M+1")
	case C_POP:
		result = append(result, SP)
		result = append(result, "A=M-1")
		result = append(result, "D=M")
		result = append(result, c.getSecondArgSegment(segment, i))
		result = append(result, c.getSecondArgSegmentIndex(segment, i)...)
		result = append(result, "M=D")
		result = append(result, SP)
		result = append(result, "M=M-1")
	}
	c.vmCodes = append(c.vmCodes, result...)
}

func (c *CodeWriter) WriteLabel(arg2 string) {
	c.vmCodes = append(c.vmCodes, "("+arg2+")")
}
func (c *CodeWriter) WriteIf(arg2 string) {
	result := make([]string, 0)
	result = append(result, SP)
	result = append(result, "M=M-1")
	result = append(result, "A=M")
	result = append(result, "D=M")
	result = append(result, "@"+arg2)
	result = append(result, "D;JNE") // jump if true(D!0)(JMPだとバグる)
	c.vmCodes = append(c.vmCodes, result...)
}
func (c *CodeWriter) WriteGoto(arg2 string) {
	result := make([]string, 0)
	result = append(result, "@"+arg2)
	result = append(result, "0;JMP") // uncoditional jump
	c.vmCodes = append(c.vmCodes, result...)
}
func (c *CodeWriter) WriteFunction(funcName string, localValCount int) {
	c.WriteLabel(funcName)
	result := make([]string, 0)
	for i := 0; i < localValCount; i++ { // initialize local variables
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "M=0")
		result = append(result, SP)
		result = append(result, "M=M+1")
	}
	c.vmCodes = append(c.vmCodes, result...)
}

func (c *CodeWriter) newLabel() string {
	c.callCount++
	return "FUNC_LABEL_" + strconv.Itoa(c.callCount)
}
func (c *CodeWriter) WriteCall(funcName string, argNum int) {
	result := make([]string, 0)
	returnLabel := c.newLabel()
	reposCount := strconv.Itoa(5 + argNum)

	result = append(result, "@"+returnLabel)
	result = append(result, "D=A") // return address
	result = append(result, SP)
	result = append(result, "A=M")
	result = append(result, "M=D") // push return address
	result = append(result, SP)
	result = append(result, "M=M+1")

	callerPointerIndice := []string{LCL, ARG, THIS, THAT}
	for _, seg := range callerPointerIndice {
		result = append(result, seg)
		result = append(result, "D=M")
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "M=D")
		result = append(result, SP)
		result = append(result, "M=M+1")
	}
	// reposition arg(SP-5-nArg)
	result = append(result, SP)
	result = append(result, "D=M")
	result = append(result, "@"+reposCount)
	result = append(result, "D=D-A")
	result = append(result, ARG)
	result = append(result, "M=D")
	// LCL = SP
	result = append(result, SP)
	result = append(result, "D=M")
	result = append(result, LCL)
	result = append(result, "M=D")

	// exec callee function
	result = append(result, "@"+funcName)
	result = append(result, "0;JMP")

	// return label
	result = append(result, "("+returnLabel+")")
	c.callCount++
	c.vmCodes = append(c.vmCodes, result...)
}
func (c *CodeWriter) WriteReturn() {
	result := make([]string, 0)
	result = append(result, LCL)
	result = append(result, "D=M")
	result = append(result, "@R13") // 一時変数だが関数によって汚染されないようにTMP[13]に保存
	result = append(result, "M=D")

	result = append(result, "@5")    // refer to before 5 idx of LCL, the indice of return address
	result = append(result, "A=D-A") // D(endFrame) -5 = returnAddress index
	result = append(result, "D=M")
	result = append(result, "@R14") // 関数によって汚染されないようにTMP[14]に保存
	result = append(result, "M=D")

	// 戻り値を *argument segment に設定
	// WritePushPop を使わない(cmd追加順が狂う)
	result = append(result, SP)
	result = append(result, "A=M-1")
	result = append(result, "D=M")
	result = append(result, ARG)
	result = append(result, "A=M")
	result = append(result, "M=D") // *ARG= pop()
	// result = append(result, SP) // SP は後続で書き換えるため不要
	// result = append(result, "M=M-1")

	// restore caller state
	result = append(result, ARG) // SP restore
	result = append(result, "D=M+1")
	result = append(result, SP)
	result = append(result, "M=D") // SP=ARG+1

	result = append(result, "@R13") // endFrame の参照
	result = append(result, "D=M")  // *(endFrame-1) のcaller THAT参照
	result = append(result, "@1")
	result = append(result, "D=D-A")
	result = append(result, "A=D") // A=*THAT
	result = append(result, "D=M") // D= caller THAT
	result = append(result, THAT)  // *THAT= caller THAT
	result = append(result, "M=D") // THAT=caller THAT

	result = append(result, "@R13") // end frame の参照
	result = append(result, "D=M")
	result = append(result, "@2")
	result = append(result, "D=D-A")
	result = append(result, "A=D") // A=THIS
	result = append(result, "D=M") // D=*THIS
	result = append(result, THIS)  // *THIS= caller THIS
	result = append(result, "M=D") // THIS=caller THIS

	result = append(result, "@R13") // end frame の参照
	result = append(result, "D=M")
	result = append(result, "@3")
	result = append(result, "D=D-A")
	result = append(result, "A=D") // A=ARG
	result = append(result, "D=M") // D=*ARG
	result = append(result, ARG)   // *ARG= caller ARG
	result = append(result, "M=D") // ARG=caller ARG

	result = append(result, "@R13") // end frame の参照
	result = append(result, "D=M")
	result = append(result, "@4")
	result = append(result, "D=D-A")
	result = append(result, "A=D") // A=LCL
	result = append(result, "D=M") // D=*LCL
	result = append(result, LCL)   // *LCL= caller LCL
	result = append(result, "M=D") // LCL=caller LCL

	// goto ret Addr
	result = append(result, "@R14")
	result = append(result, "A=M")
	result = append(result, "0;JMP") // return address を参照
	c.vmCodes = append(c.vmCodes, result...)
}

func (c *CodeWriter) AssembledCodes() []string {
	return c.vmCodes
}
