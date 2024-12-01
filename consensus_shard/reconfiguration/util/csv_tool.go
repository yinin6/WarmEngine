package util

import (
	"blockEmulator/params"
	"log"
	"os"
	"strings"
)

func GetFilePath(name string) string {
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, ".", "_")
	dirpath := params.DataWrite_path
	return dirpath + name + ".csv"
}

func MakeCsv(columns []string, name string) {
	// Construct directory path
	dirpath := params.DataWrite_path
	if err := os.MkdirAll(dirpath, os.ModePerm); err != nil {
		log.Panic(err)
	}
	// Construct target file path
	targetPath := GetFilePath(name)

	// Open file, create if it does not exist
	file, err := os.OpenFile(targetPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	str := ""
	for i := 0; i < len(columns)-1; i++ {
		str += columns[i] + ","
	}
	// Write column names
	if _, err := file.WriteString(str + columns[len(columns)-1] + "\n"); err != nil {
		log.Panic(err)
	}

}
