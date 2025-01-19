package symboltable

import (
	"regexp"
	"strconv"
)

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
