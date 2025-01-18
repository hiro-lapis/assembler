package main

import (
	"bufio"
	"fmt"
	"os"
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

	p := Parser{}
	c := Code{}
	parsedLine := make([]string, 0)
	idx := 0
	for i := 0; i < len(inputs); i++ {
		l, isA := p.Exec(inputs[i])
		if l != "" {
			idx++
			binaryStr := c.Exec(l, isA)
			parsedLine = append(parsedLine, binaryStr)
		}
	}

	// createFile("main.hack", []string{"Hello", "World"})
	createFile("main.hack", parsedLine)
}

type Parser struct {
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

	data := make([]byte, 1000)
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
