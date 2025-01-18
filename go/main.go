package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// fmt.Println("please input compile target")
	// fmt.Scan(&fileName)
	fileNames := []string{
		"./asm/Add.asm",
	}
	fileName := fileNames[0]
	inputs, err := openFile(fileName)
	if err != nil {
		fmt.Println("file open error")
		return
	}

	p := Parser{instructions: inputs}
	p.Next()
	// s := SymbolTable{}
	c := Code{}
	parsedLine := make([]string, 0)
	for {
		if !p.HasMoreLines() {
			break
		}
		binaryStr := ""
		if p.InstructionType() == A_INSTRUCTION {
			binaryStr = c.ExecA(p.Label())
		} else if p.InstructionType() == C_INSTRUCTION {
			binaryStr = c.ExecC(p.Dest(), p.Comp(), p.Jump())
		}
		parsedLine = append(parsedLine, binaryStr)
		p.Next()
	}

	createFile("main.hack", parsedLine)
}

type InstructionType int

const (
	A_INSTRUCTION InstructionType = iota
	C_INSTRUCTION
	L_INSTRUCTION
)

type Parser struct {
	instructions              []string
	currentOriginalFileLineNo int
	currentBinaryLineNo       int // refered line no, can be binary code line. This is counted with skiping line in the case of comment and label expression
}

func (i InstructionType) String() string {
	return [...]string{"A_INSTRUCTION", "C_INSTRUCTION", "L_INSTRUCTION"}[i]
}
func (p *Parser) HasMoreLines() bool {
	if p.currentOriginalFileLineNo-1 >= len(p.instructions) {
		return false
	}
	for _, line := range p.instructions[p.currentOriginalFileLineNo:] {
		if p.isBinarizable(line) {
			return true
		}
	}
	return false
}
func (p *Parser) Next() {
	max := len(p.instructions)
	if p.HasMoreLines() {
		for {
			if p.currentOriginalFileLineNo >= max-1 {
				return
			}
			p.currentOriginalFileLineNo++
			if p.isBinarizable(string(p.instructions[p.currentOriginalFileLineNo])) {
				break
			}
		}
		p.currentBinaryLineNo++
	}
}
func (p *Parser) InstructionType() InstructionType {
	line := p.instructions[p.currentOriginalFileLineNo]
	// Next() always step to binarizable line so that we dont have to comment and space
	if string(line[0]) == "@" {
		isNum, _ := regexp.MatchString(`^[0-9]+$`, line[1:])
		if isNum {
			return A_INSTRUCTION
		}
		return L_INSTRUCTION
	}
	return C_INSTRUCTION
}
func (p *Parser) Label() string {
	line := strings.TrimSpace(p.instructions[p.currentOriginalFileLineNo])
	if p.InstructionType() == A_INSTRUCTION {
		return line[1:]
	}
	return ""
}
func (p *Parser) Dest() string {
	line := strings.TrimSpace(p.instructions[p.currentOriginalFileLineNo])
	if p.InstructionType() == C_INSTRUCTION && strings.Contains(line, "=") {
		return strings.Split(line, "=")[0]
	}
	return ""
}
func (p *Parser) Comp() string {
	line := strings.TrimSpace(p.instructions[p.currentOriginalFileLineNo])
	if p.InstructionType() == C_INSTRUCTION {
		cj := strings.Split(line, "=")[1]
		return strings.Split(cj, ";")[0]
	}
	return ""
}
func (p *Parser) Jump() string {
	line := strings.TrimSpace(p.instructions[p.currentOriginalFileLineNo])
	if p.InstructionType() == C_INSTRUCTION {
		if strings.Contains(line, ";") {
			return strings.Split(line, ";")[1]
		}
	}
	return ""
}

func (p *Parser) isBinarizable(line string) bool {
	// skip white space or comment, but label expression is binaraizale
	str := strings.TrimSpace(line)
	return str != "" && string(str[:2]) != "//"
}

func (p *Parser) Exec(line string) (v string, isA bool) {
	if len(line) == 0 {
		return "", isA
	}
	index := strings.Index(line, "//")
	if index != -1 {
		v = line[:index]
	} else {
		v = line
	}
	v = strings.TrimSpace(v)
	if len(v) == 0 {
		return "", isA
	}
	isA = string(v[0]) == "@"
	return v, isA
}

const OP_CODE_A = "0"
const OP_CODE_C = "1110"

type Code struct {
}

func (c *Code) Exec(v string, isA bool) (result string) {
	if isA {
		num, _ := strconv.Atoi(v[1:])
		binaryStr := strconv.FormatInt(int64(num), 2)
		result = OP_CODE_A + strings.Repeat("0", 15-len(binaryStr)) + binaryStr
		return result
	}
	codes := strings.Split(v, "=")
	codes2 := strings.Split(codes[1], ";")
	var jump, dest, comp string
	dest = codes[0]
	if codes2[0] != "" {
		comp = codes2[0]
	}
	if len(codes2) > 1 && codes2[1] != "" {
		jump = codes2[1]
	}
	cccccc := c.computation(comp)
	ddd := c.destination(dest)
	jjj := c.jump(jump)
	return OP_CODE_C + cccccc + ddd + jjj
}
func (c *Code) ExecA(symbol string) (result string) {
	num, _ := strconv.Atoi(symbol)
	binaryStr := strconv.FormatInt(int64(num), 2)
	result = OP_CODE_A + strings.Repeat("0", 15-len(binaryStr)) + binaryStr
	return result
}
func (c *Code) ExecC(dest string, comp string, jump string) (result string) {
	ddd := c.destination(dest)
	cccccc := c.computation(comp)
	jjj := c.jump(jump)
	return OP_CODE_C + cccccc + ddd + jjj
}

func (c *Code) computation(v string) string {
	var result string
	switch string(v) {
	case "0":
		return "101010"
	case "1":
		return "111111"
	case "-1":
		return "111010"
	case "D":
		return "001100"
	case "A", "M":
		return "110000"
	case "!D":
		return "0001101"
	case "!A", "!M":
		return "110001"
	case "-D":
		return "001111"
	case "-A", "-M":
		return "110011"
	case "D+1":
		return "011111"
	case "A+1", "M+1":
		return "110111"
	case "D-1":
		return "001110"
	case "A-1", "M-1":
		return "110010"
	case "D+A", "D+M":
		return "000010"
	case "D-M", "D-A":
		return "010011"
	case "A-D", "M-D":
		return "000111"
	case "D&A", "D&M":
		return "000000"
	case "D|A", "D|M":
		return "010101"
	}
	return result
}

func (c *Code) destination(v string) string {
	var result string
	switch string(v) {
	case "": // null
		return "000"
	case "M":
		return "001"
	case "D":
		return "010"
	case "DM":
		return "011"
	case "A":
		return "100"
	case "AM":
		return "101"
	case "AD":
		return "110"
	case "ADM":
		return "111"
	}
	return result
}

func (c *Code) jump(v string) string {
	var result string
	switch string(v) {
	case "": // nul
		return "000"
	case "JGT":
		return "001"
	case "JEQ":
		return "010"
	case "JGE":
		return "011"
	case "JLT":
		return "100"
	case "JNE":
		return "101"
	case "JLE":
		return "110"
	case "JMP":
		return "111"
	}
	return result
}

type SymbolTable struct {
}

func (s *SymbolTable) Exec() string {
	return ""
}

// ファイルを開いて, その中身を返す
func openFile(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	data := make([]byte, 100000)
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
