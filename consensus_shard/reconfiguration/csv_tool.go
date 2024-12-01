package reconfiguration

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// 定义一个写入函数
func writeDataToCSV(header []string, filename string, data ...int) error {
	// 检查文件是否存在
	fileExists := fileExists(filename)

	// 打开或创建文件
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 创建 CSV 写入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 如果文件不存在或为空，写入表头
	if !fileExists || isFileEmpty(filename) {
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("写入表头失败: %w", err)
		}
	}

	// 将 int 数组转换为字符串数组
	record := make([]string, len(data))
	for i, val := range data {
		record[i] = strconv.Itoa(val)
	}

	// 写入一行数据
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("写入 CSV 数据失败: %w", err)
	}

	return nil
}

// 检查文件是否存在
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// 检查文件是否为空
func isFileEmpty(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return true
	}
	return info.Size() == 0
}
