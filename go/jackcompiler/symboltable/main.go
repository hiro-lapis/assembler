package symboltable

import (
	"fmt"

	"github.com/hiro-lapis/jackanalyzer/vmwriter"
)

type TableKind int

const (
	CLASS_LEVEL TableKind = iota
	SUBROUTINE_LEVEL
)

type Var struct {
	i  int
	t  string
	vk vmwriter.Segment // STATIC, FIELD, ARGUMENT, LCL, THIS, THAT
}
type SubVar struct {
	i  int
	t  string
	vk vmwriter.Segment
}

type SymbolTable struct {
	className       string
	classLevel      map[string]*Var         // class level table
	subroutineLevel map[int]map[string]*Var // subroutine level table
	depth           int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		className:       "",
		classLevel:      make(map[string]*Var),
		subroutineLevel: make(map[int]map[string]*Var),
	}
}

// reset subroutine table
func (s *SymbolTable) StartSubroutine() {
	s.subroutineLevel = make(map[int]map[string]*Var)
}

// reset subroutine table
func (s *SymbolTable) SetClassName(n string) {
	s.className = n
}

func (s *SymbolTable) Define(tl TableKind, name string, t string, vk vmwriter.Segment) {
	index := 0
	if tl == CLASS_LEVEL {
		// increment var index besed on exsting same kind vars
		for _, v := range s.classLevel {
			if v.vk == vk {
				index++
			}
		}
		s.classLevel[name] = &Var{i: index, t: t, vk: vk}
	} else {
		// search only current depth
		for _, v := range s.subroutineLevel[s.depth] {
			if v.vk == vk {
				index++
			}
		}
		if s.subroutineLevel[s.depth] == nil {
			s.subroutineLevel[s.depth] = make(map[string]*Var)
		}
		s.subroutineLevel[s.depth][name] = &Var{i: index, t: t, vk: vk}
	}
}

func (s *SymbolTable) IncrementDepth() {
	s.depth = +1
}

func (s *SymbolTable) IndexOf(name string) (int, error) {
	v, err := s.find(name)
	if err != nil {
		return -1, err
	}
	return v.i, err
}

func (s *SymbolTable) TypeOf(name string) (string, error) {
	v, err := s.find(name)
	if err != nil {
		return "", err
	}
	return v.t, err
}

func (s *SymbolTable) KindOf(name string) (vmwriter.Segment, error) {
	v, err := s.find(name)
	if err != nil {
		return vmwriter.STATIC, err
	}
	return v.vk, err
}

func (s *SymbolTable) VarCount(vk vmwriter.Segment) (int, error) {
	count := 0
	for _, v := range s.classLevel {
		if v.vk == vk {
			count++
		}
	}
	for _, v := range s.classLevel {
		if v.vk == vk {
			count++
		}
	}
	return count, nil
}

func (s *SymbolTable) find(name string) (*Var, error) {
	// look current depth first, then upper depth
	for i := s.depth; i >= 0; i-- {
		v, ok := s.subroutineLevel[i][name]
		if ok {
			return v, nil
		}
	}
	// var doesn't exist subroutine level, then check class level
	v, ok := s.classLevel[name]
	if ok {
		return v, nil
	}
	err := fmt.Errorf("error: var is not found")
	return nil, err
}
