package vmwriter

import (
	"bufio"
	"os"
	"strconv"
)

type VmWriter struct {
	f *os.File
	w *bufio.Writer
}

type Segment int

const (
	CONSTANT Segment = iota
	LCL
	ARG
	TEMP
	STATIC
	THIS
	THAT
)

var ArtCmds = map[string]string{
	"+": "add",
	"-": "sub",
	"*": "call Math.multiply 2",
	"/": "call Math.divide 2",
	"&": "and",
	"|": "or",
	"=": "eq",
	">": "gt",
	"<": "lt",
}

var UnaryCmds = map[string]string{
	"-": "neg",
	"~": "not",
}

// receive pointer of os.File which is made by os.Create(fileName)
func NewVmWriter(file *os.File) *VmWriter {
	return &VmWriter{
		f: file,
		w: bufio.NewWriter(file),
	}
}

func (v *VmWriter) toString(segment Segment) string {
	if segment == CONSTANT {
		return "constant"
	}
	if segment == LCL {
		return "local"
	}
	if segment == ARG {
		return "argument"
	}
	if segment == TEMP {
		return "temp"
	}
	if segment == THIS {
		return "pointer" // 0
	}
	if segment == THAT { // 1
		return "pointer"
	}
	return "static"
}
func (v *VmWriter) WritePush(segment Segment, index int) {
	v.w.WriteString("push " + v.toString(segment) + " " + strconv.Itoa(index) + "\n")
}

func (v *VmWriter) WritePop(segment Segment, index int) {
	v.w.WriteString("pop " + v.toString(segment) + " " + strconv.Itoa(index) + "\n")
}

func (v *VmWriter) WriteArithmetic(command string) {
	v.w.WriteString(command + "\n")
}

func (v *VmWriter) WriteLabel(label string) {
	v.w.WriteString("label " + label + "\n")
}

func (v *VmWriter) WriteGoto(label string) {
	v.w.WriteString("goto " + label + "\n")
}

func (v *VmWriter) WriteIf(label string) {
	v.w.WriteString("if-goto " + label + "\n")
}

func (v *VmWriter) WriteCall(name string, nArgs int) {
	v.w.WriteString("call " + name + " " + strconv.Itoa(nArgs) + "\n")
}

// constructor 含めてこれでコンパイル
func (v *VmWriter) WriteFunction(name string, nLocals int) {
	v.w.WriteString("function " + name + " " + strconv.Itoa(nLocals) + "\n")
}

func (v *VmWriter) WriteReturn() {
	v.w.WriteString("return" + "\n")
}

func (v *VmWriter) Close() {
	v.w.Flush()
	v.f.Close()
}
