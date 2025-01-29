package main

import (
	"bufio"
	"fmt"
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
	STT   = "@6" // 16-255
	POINT = ""
)

// Auto grader の実行環境v1.13以下での実行を想定
func main() {
	// OK[add, static, basic, pointer]
	// flag.Parse()
	// fileName := flag.Args()[0]
	// fileName := "SimpleAdd.vm"
	// fileName := "StaticTest.vm"
	// fileName := "BasicTest.vm"
	// fileName := "SimpleTest.vm"
	fileName := "PointerTest.vm"
	outputFileName := strings.Split(fileName, ".")[0] + ".asm"

	inputs, err := openFile(fileName)
	if err != nil {
		fmt.Println("file open error")
		return
	}

	p := NewParser(inputs)
	c := NewCodeWriter()
	// if more line

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
		}
		p.advance()
	}

	asmLines := c.AssembledCodes()
	createFile(outputFileName, asmLines)
}

// return file contents without formating
func openFile(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	// 10mb unexpected _485_760 at end of statement とエラー出る(v1.13以下?)ので通常のリテラルに変更
	maxFileSize := 10485760
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
	list := strings.Split(string(data[:count]), "\n")
	return list, nil
}

func createFile(name string, data []string) {
	file, err := os.Create(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for i := 0; i < len(data); i++ {
		w.WriteString(data[i] + "\n")
	}
	w.Flush()
}

type Parser struct {
	commands    []string
	currentLine int
}

func NewParser(lines []string) *Parser {
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
	return &Parser{commands: list, currentLine: 0}
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

type CommandType int

const (
	C_ARITHMETIC CommandType = iota
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_RETURN
	C_CALL
)

func (p *Parser) CommandType() CommandType {
	cmds := strings.Split(p.commands[p.currentLine], " ")
	if len(cmds) == 1 {
		return C_ARITHMETIC
	}
	cmd := cmds[0]
	switch string(cmd) {
	case "push":
		return C_PUSH
	case "pop":
		return C_POP
	}
	// TODO: implement other command types
	return C_ARITHMETIC
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
	if p.CommandType() != C_PUSH && p.CommandType() != C_POP {
		return -1, fmt.Errorf("Arg2 should be called only push or pop")
	}
	cmds := strings.Split(p.commands[p.currentLine], " ")
	if len(cmds) == 1 {
		return -1, fmt.Errorf("Arg2 should be called only push or pop")
	}
	num, _ := strconv.Atoi(cmds[2])
	return num, nil
}

type CodeWriter struct {
	vmCodes []string
}

func NewCodeWriter() *CodeWriter {
	initAsm := [][]string{
		{"0", "256"},
		{"1", "300"},
		{"2", "400"},
		{"3", "3000"},
		{"4", "3010"},
		{"5", "5"},
		{"6", "16"},
	}
	setPointer := func(setTarget, value string) []string {
		return []string{"@" + value, "D=A", "@" + setTarget, "M=D"}
	}
	asmLines := make([]string, 0)
	for _, l := range initAsm {
		cmds := setPointer(l[0], l[1])
		asmLines = append(asmLines, cmds...)
	}
	// asmLines = append(asmLines,
	// 	"CONDITION_TRUE",
	// 	"M=-1",
	// 	"@13",
	// 	"A=M",
	// 	"CONDITION_TRUE",
	// 	"M=0",
	// 	"@13",
	// 	"A=M",
	// )
	return &CodeWriter{
		vmCodes: asmLines,
	}
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
		result = append(result, "M=-1")
		result = append(result, "@CONDITION_TRUE")
		result = append(result, "D;JEQ")
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
	case "gt":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "A=A-1")
		result = append(result, "D=M") // x
		result = append(result, "A=A+1")

		result = append(result, "D=D-M") // x - y
		result = append(result, "M=-1")
		result = append(result, "@GT_TRUE")
		result = append(result, "D;JGT")
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "A=A-1")
		result = append(result, "M=0")
		result = append(result, "(GT_TRUE)")
		result = append(result, SP)
		result = append(result, "M=M-1")
	case "lt":
		break
	case "and":
		break
	case "or":
		break
	case "not":
		break
	}
	c.vmCodes = append(c.vmCodes, result...)
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
		return STT
	}

	return ""
}

func (c *CodeWriter) getSecondArgSegmentIndex(segment string, i int) []string {
	l := make([]string, 0)
	if segment != "pointer" {
		l = append(l, "A=M")
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

func (c *CodeWriter) AssembledCodes() []string {
	return c.vmCodes
}
