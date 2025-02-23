package vmwriter

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type VmWriter struct {
	f *os.File
	w *bufio.Writer
}

type Segment int

const (
	CONSTANT Segment = iota
	LOCAL
	ARG
	TEMP
	STATIC
	THIS
	THAT
)

// receive pointer of os.File which is made by os.Create(fileName)
func NewVmWriter(file *os.File) *VmWriter {
	return &VmWriter{
		f: file,
		w: bufio.NewWriter(file),
	}
}

func (v *VmWriter) toString(segment Segment) string {
	if segment == CONSTANT {
		return "constatnt"
	}
	if segment == LOCAL {
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
	// ADD, SUB, NEG, EQ, GT
	v.w.WriteString(strings.ToUpper(command) + "\n")
}

func (v *VmWriter) WriteLabel(label string) {
	v.w.WriteString("(" + label + ")" + "\n")
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
	v.w.WriteString("return \n")
}

func (v *VmWriter) Close() {
	v.w.Flush()
	v.f.Close()
}
