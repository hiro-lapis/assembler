package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	SP    = "@0" // 256
	LCL   = "@1" // 300
	ARG   = "@2" // 400
	THIS  = "@3" // 3000
	THAT  = "@4" // 3010
	TMP   = "@5" // 5-12
	STT   = "@6" // 16-255
	POINT = ""
)

// Auto grader の実行環境v1.13以下での実行を想定
func main() {
	flag.Parse()
	fileName := flag.Args()[0]
	// fileName := "SimpleAdd.vm"
	// fileName := "StaticTest.vm"
	// fileName := "SimpleTest.vm"
	outputFileName := fileName[:len(fileName)-3] + ".asm"
	fmt.Println(outputFileName)

	inputs, err := openFile(fileName)
	if err != nil {
		fmt.Println("file open error")
		return
	}
	lines := make([]string, 0)
	for _, v := range inputs {
		if v == "" || v == "//" {
			continue
		}
		lines = append(lines, strings.TrimSpace(v))
	}

	asmLines := make([]string, 0)
	initAsm := [][]string{
		{"0", "256"},
		{"1", "300"},
		{"2", "400"},
		{"3", "3000"},
		{"4", "3010"},
		{"5", "5"},
		{"6", "16"},
	}
	for _, l := range initAsm {
		cmds := setPointer(l[0], l[1])
		asmLines = append(asmLines, cmds...)
	}

	for _, l := range lines {
		cmd := strings.Split(l, " ")

		if len(cmd) == 3 {
			cmds := assembleArithmetic(cmd[0], cmd[1], cmd[2])
			asmLines = append(asmLines, cmds...)
		} else if len(cmd) == 1 {
			cmds := assembleLogistic(cmd[0])
			asmLines = append(asmLines, cmds...)
		}
	}
	createFile(outputFileName, asmLines)
}

func setPointer(setTarget, value string) []string {
	return []string{"@" + value, "D=A", "@" + setTarget, "M=D"}
}
func assembleArithmetic(cmd, segment, index string) []string {
	result := make([]string, 0)
	switch cmd {
	case "push":
		if segment == "constant" {
			result = append(result, "@"+index)
			result = append(result, "D=A")
			result = append(result, SP)
			result = append(result, "A=M")
			result = append(result, "M=D")
			result = append(result, SP)
			result = append(result, "M=M+1")
		} else if segment == "static" {
			result = append(result, STT)
			result = append(result, "A=M")
			i, _ := strconv.Atoi(index)
			for i > 0 {
				result = append(result, "A=A+1")
				i--
			}
			result = append(result, "D=M")
			result = append(result, SP)
			result = append(result, "A=M")
			result = append(result, "M=D")
			result = append(result, SP)
			result = append(result, "M=M+1")
		}
	case "pop":
		if segment == "static" {
			result = append(result, SP)
			result = append(result, "A=M-1")
			result = append(result, "D=M")
			result = append(result, STT)
			result = append(result, "A=M")
			i, _ := strconv.Atoi(index)
			for i > 0 {
				result = append(result, "A=A+1")
				i--
			}
			result = append(result, "M=D")
			result = append(result, SP)
			result = append(result, "M=M-1")
		}
	}
	return result
}
func assembleLogistic(cmd string) []string {
	result := make([]string, 0)
	switch cmd {
	case "add":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "D=M") // D = *SP
		result = append(result, "A=A-1")
		result = append(result, "A=A-1")
		result = append(result, "D=M")
		result = append(result, "A=A+1")
		result = append(result, "D=D+M")
		result = append(result, "A=A-1")
		result = append(result, "M=D")
		result = append(result, SP)
		result = append(result, "M=M-1")
	case "sub":
		result = append(result, SP)
		result = append(result, "A=M")
		result = append(result, "A=A-1")
		result = append(result, "A=A-1")
		result = append(result, "D=M")   // x(stack 2つ目の値)
		result = append(result, "A=A+1") // y(stack 1つ目の値)
		result = append(result, "D=D-M") // y - x
		result = append(result, "A=A-1")
		result = append(result, "M=D") // x = y - x(yのメモリは上書き対象)
		result = append(result, SP)
		result = append(result, "M=M-1")
	}
	return result
}

// return file contents without formating
func openFile(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	// 10mb unexpected _485_760 at end of statement とエラー出る(v1.13以下?)ので通常のリテラルに変更
	maxFileSize := 10485760
	data := make([]byte, maxFileSize)
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
