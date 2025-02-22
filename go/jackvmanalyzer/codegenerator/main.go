package codegenerator

import "github.com/hiro-lapis/jackanalyzer/symboltable"

type CodeGenerator struct {
	st *SymbolTable
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		st: symboltable.NewSymbolTable(),
	}
}

// Test programs
// Seven
// ConvertToBin
// Square
// Averate
// Pong
// ComplexArrays

func (c *CodeGenerator) Generate() {
}

func (c *CodeGenerator) Generate() {
}
