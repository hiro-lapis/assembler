package parser

import (
	"strings"
)

type InstructionType int

const (
	A_INSTRUCTION InstructionType = iota
	C_INSTRUCTION
	L_INSTRUCTION
)

type Parser struct {
	instructions []string
	currentLine  int
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
	return &Parser{instructions: list}
}

func (i InstructionType) String() string {
	return [...]string{"A_INSTRUCTION", "C_INSTRUCTION", "L_INSTRUCTION"}[i]
}
func (p *Parser) HasMoreLines() bool {
	maxIdx := len(p.instructions) - 1
	return maxIdx > p.currentLine
}

func (p *Parser) CurrentLine() int {
	return p.currentLine
}
func (p *Parser) Reset() {
	p.currentLine = 0
}
func (p *Parser) Next() {
	if p.HasMoreLines() {
		p.currentLine++
	}
}
func (p *Parser) InstructionType() InstructionType {
	line := p.instructions[p.currentLine]
	// Next() always step to binarizable line so that we dont have to comment and space
	if string(line[0]) == "@" {
		return A_INSTRUCTION
	}
	if string(line[0]) == "(" && string(line[len(line)-1]) == ")" {
		return L_INSTRUCTION
	}
	return C_INSTRUCTION
}
func (p *Parser) Label() string {
	if p.InstructionType() == L_INSTRUCTION {
		return p.instructions[p.currentLine][1 : len(p.instructions[p.currentLine])-1]
	}
	return ""
}
func (p *Parser) Symbol() string {
	if p.InstructionType() == A_INSTRUCTION {
		return p.instructions[p.currentLine][1:]
	}
	return ""
}
func (p *Parser) Dest() string {
	line := p.instructions[p.currentLine]
	if p.InstructionType() == C_INSTRUCTION && strings.Contains(line, "=") {
		return strings.Split(p.instructions[p.currentLine], "=")[0]
	}
	return ""
}
func (p *Parser) Comp() string {
	if p.InstructionType() == C_INSTRUCTION {
		cj := strings.Split(p.instructions[p.currentLine], "=")
		if len(cj) > 1 {
			return strings.Split(cj[1], ";")[0]
		}
		return strings.Split(cj[0], ";")[0]
	}
	return ""
}
func (p *Parser) Jump() string {
	if p.InstructionType() == C_INSTRUCTION {
		if strings.Contains(p.instructions[p.currentLine], ";") {
			return strings.Split(p.instructions[p.currentLine], ";")[1]
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
