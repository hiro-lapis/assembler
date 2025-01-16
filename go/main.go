package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// fmt.Println("please input compile target")
	var fileName string
	// fmt.Scan(&fileName)
	fileName = "Add.asm"
	inputs, err := openFile(fileName)
	if err != nil {
		fmt.Println("file open error")
		return
	}
	for i := 0; i < len(inputs); i++ {
		fmt.Println(inputs[i])
	}

}

// ファイルを開いて, その中身を返す
func openFile(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	data := make([]byte, 1000)
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
