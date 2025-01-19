package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hiro-lapis/assembler/parser"
)

const MAX_FILE_SIZE = 10_485_760

func main() {
	fileNames := []string{
		"./asm/Add.asm",
		"./asm/Max.asm",
		"./asm/Pong.asm",
		"./asm/Rect.asm",
	}
	c := &Code{}
	for _, fileName := range fileNames {
		inputs, err := openFile(fileName)
		if err != nil {
			fmt.Println("file open error")
			return
		}

		p := parser.NewParser(inputs)
		s := NewSymbolTable()
		// register second path
		setPath(p, s)
		assemble(p, c, s, fileName)
	}
}

// Set symbol table's label path
func setPath(p *parser.Parser, s *SymbolTable) {
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

func assemble(p *parser.Parser, c *Code, s *SymbolTable, fileName string) {
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

const OP_CODE_A = "0"
const OP_CODE_C = "111"

type Code struct {
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
		return "0101010"
	case "1":
		return "0111111"
	case "-1":
		return "0111010"
	case "D":
		return "0001100"
	case "A":
		return "0110000"
	case "M":
		return "1110000"
	case "!D":
		return "0001101"
	case "!A":
		return "0110001"
	case "!M":
		return "1110001"
	case "-D":
		return "0001111"
	case "-A":
		return "0110011"
	case "-M":
		return "1110011"
	case "D+1":
		return "0011111"
	case "A+1":
		return "0110111"
	case "M+1":
		return "1110111"
	case "D-1":
		return "0001110"
	case "A-1":
		return "0110010"
	case "M-1":
		return "1110010"
	case "D+A":
		return "0000010"
	case "D+M":
		return "1000010"
	case "D-A":
		return "0010011"
	case "D-M":
		return "1010011"
	case "A-D":
		return "0000111"
	case "M-D":
		return "1000111"
	case "D&A":
		return "0000000"
	case "D&M":
		return "1000000"
	case "D|A":
		return "0010101"
	case "D|M":
		return "1010101"
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
	case "DM", "MD":
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

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		// pre-defined symbols
		symbol: map[string]int{
			"R0":     0,
			"R1":     1,
			"R2":     2,
			"R3":     3,
			"R4":     4,
			"R5":     5,
			"R6":     6,
			"R7":     7,
			"R8":     8,
			"R9":     9,
			"R10":    10,
			"R11":    11,
			"R12":    12,
			"R13":    13,
			"R14":    14,
			"R15":    15,
			"SCREEN": 16384,
			"KBD":    24576,
			"SP":     0,
			"LCL":    1,
			"ARG":    2,
			"THIS":   3,
			"THAT":   4,
		},
	}
}

const START_SECOND_PATH_IDX = 16

type SymbolTable struct {
	symbol          map[string]int // symbol:address
	firstPathCount  int
	secondPathCount int
}

// ng: number ahead, pre-defined word, (,),@,;
func (s *SymbolTable) AddLabel(symbol string, currentLine int) (bin string) {
	// if it is assign address, return the decimal number
	isNum, _ := regexp.MatchString(`^[0-9]+$`, symbol[1:])
	if isNum {
		return symbol
	}
	if val, ok := s.symbol[symbol]; ok {
		return string(val)
	}
	// since Hack assembler start line from 0, don't need to increment for next line reference
	s.symbol[symbol] = currentLine - s.firstPathCount
	s.firstPathCount++
	return string(currentLine)
}
func (s *SymbolTable) GetValue(symbol string) (address string) {
	address = symbol
	// if it is assign address, return the decimal number
	isNum, _ := regexp.MatchString(`^[0-9]+$`, symbol)
	if isNum {
		return address
	}
	if val, ok := s.symbol[symbol]; ok {
		address = strconv.Itoa(val)
		return address
	}
	s.symbol[symbol] = START_SECOND_PATH_IDX + s.secondPathCount
	s.secondPathCount++
	address = strconv.Itoa(s.symbol[symbol])
	return address
}

// ファイルを開いて, その中身を返す
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
