package vmwriter

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type VmWriter struct {
	w *bufio.Writer
}

// receive pointer of os.File which is made by os.Create(fileName)
func NewVmWriter(file *os.File) *VmWriter {
	return &VmWriter{
		w: bufio.NewWriter(file),
	}
}

func (v *VmWriter) WritePush(segment string, index int) {
	v.w.WriteString("push " + segment + " " + strconv.Itoa(index) + "\n")
}

func (v *VmWriter) WritePop(segment string, index int) {
	v.w.WriteString("pop " + segment + " " + strconv.Itoa(index) + "\n")
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

func (v *VmWriter) WriteFunction(name string, nLocals int) {
	v.w.WriteString("function " + name + " " + strconv.Itoa(nLocals) + "\n")
}

func (v *VmWriter) WriteReturn() {
	v.w.WriteString("return \n")
}

func (v *VmWriter) Close() {
	v.w.Flush()
}
