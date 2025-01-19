package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hiro-lapis/assembler/code"
	"github.com/hiro-lapis/assembler/parser"
	"github.com/hiro-lapis/assembler/symboltable"
)

const MAX_FILE_SIZE = 10_485_760

func main() {
	fileNames := []string{
		"./asm/Add.asm",
		"./asm/Max.asm",
		"./asm/Pong.asm",
		"./asm/Rect.asm",
	}
	c := &code.Code{}
	for _, fileName := range fileNames {
		inputs, err := openFile(fileName)
		if err != nil {
			fmt.Println("file open error")
			return
		}

		p := parser.NewParser(inputs)
		s := symboltable.NewSymbolTable()
		// register second path
		setPath(p, s)
		assemble(p, c, s, fileName)
	}
}

// Set symbol table's label path
func setPath(p *parser.Parser, s *symboltable.SymbolTable) {
	for {
		if p.InstructionType() == parser.L_INSTRUCTION {
			s.AddLabel(p.Label(), p.CurrentLine())
		}
		if !p.HasMoreLines() {
			break
		}
		p.Next()
	}
	p.Reset()
}

func spliceFileName(path string) string {
	slashIndex := strings.LastIndex(path, "/")
	if slashIndex == -1 {
		return ""
	}
	dotIndex := strings.LastIndex(path, ".")
	if dotIndex == -1 || dotIndex < slashIndex {
		return ""
	}
	return path[slashIndex+1 : dotIndex]
}

func assemble(p *parser.Parser, c *code.Code, s *symboltable.SymbolTable, fileName string) {
	parsedLine := make([]string, 0)
	for {
		binaryStr := ""
		if p.InstructionType() == parser.A_INSTRUCTION {
			symol := s.GetValue(p.Symbol())
			binaryStr = c.ExecA(symol)
		} else if p.InstructionType() == parser.C_INSTRUCTION {
			binaryStr = c.ExecC(p.Dest(), p.Comp(), p.Jump())
		}

		if binaryStr != "" {
			parsedLine = append(parsedLine, binaryStr)
		}
		if !p.HasMoreLines() {
			break
		}
		p.Next()
	}
	newFileName := spliceFileName(fileName)
	createFile(newFileName+".hack", parsedLine)
}

// return file contents without formating
func openFile(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	// 10mb
	data := make([]byte, MAX_FILE_SIZE)
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
