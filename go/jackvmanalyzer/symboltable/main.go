package symboltable

import "fmt"

type VarKind int

const (
	STATIC VarKind = iota
	FIELD
	ARGUMENT
	VAR
)

type TableKind int

const (
	CLASS_LEVEL TableKind = iota
	SUBROUTINE_LEVEL
)

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		classLevel:      make(map[string]*Var),
		subroutineLevel: make(map[int]map[string]*Var),
	}
}

type Var struct {
	i  int
	t  string
	vk VarKind
}
type SubVar struct {
	i  int
	t  string
	vk VarKind
}

type SymbolTable struct {
	classLevel      map[string]*Var         // class level table
	subroutineLevel map[int]map[string]*Var // subroutine level table
	depth           int
}

// class level
// x int 0
// y int 0
//
//

// reset subroutine table
func (s *SymbolTable) StartSubroutine() {
	s.subroutineLevel = make(map[int]map[string]*Var)
}

func (s *SymbolTable) Define(tl TableKind, name string, t string, vk VarKind) {
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
		s.subroutineLevel[0][name] = &Var{i: index, t: t, vk: vk}
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

func (s *SymbolTable) KindOf(name string) (VarKind, error) {
	v, err := s.find(name)
	if err != nil {
		return STATIC, err
	}
	return v.vk, err
}

func (s *SymbolTable) VarCount(vk VarKind) (int, error) {
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
	for i := s.depth; i > 0; i-- {
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
