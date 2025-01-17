package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// fmt.Println("please input compile target")
	var fileName string
	// fmt.Scan(&fileName)
	fileName = "Add.asm"
	inputs, err := openFile(fileName)
	if err != nil {
		fmt.Println("file open error")
		return
	}
	for i := 0; i < len(inputs); i++ {
		fmt.Println(inputs[i])
	}

	createFile("main.hack", []string{"Hello", "World"})
}

type Parser struct {
}

func (p *Parser) Exec(line string) (v string, isA bool) {
	return v, isA
}

type Code struct {
}

func (c *Code) Exec(v string) string {
	return "1110001100001000"
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
