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
const OP_CODE_C = "111"

type Code struct {
}

func (c *Code) Exec(v string, isA bool) (result string) {
	if isA {
		num, _ := strconv.Atoi(v[1:])
		binaryStr := strconv.FormatInt(int64(num), 2)
		result = OP_CODE_A + strings.Repeat("0", 15-len(binaryStr)) + binaryStr
		return result
	}
	return OP_CODE_C + "accccccdddjjj"
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
