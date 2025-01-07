package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func OpenOrCreateCSVFile(filePath string) (*os.File, *csv.Writer, error) {
	var file *os.File
	var err error
	if fileExists(filePath) {
		file, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			return nil, nil, fmt.Errorf("无法打开文件: %v", err)
		}
	} else {
		file, err = os.Create(filePath)
		if err != nil {
			return nil, nil, fmt.Errorf("无法创建文件: %v", err)
		}
	}
	writer := csv.NewWriter(file)
	return file, writer, nil
}

func CheckCSVFileExists(inputFilePath, validFilePath, invalidFilePath string) ([][]string, *csv.Writer, *csv.Writer, error) {
	// 打开源 CSV 文件
	sourceFile, err := os.Open(inputFilePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error opening file: %v", err)
	}
	defer sourceFile.Close()

	// 读取 CSV 文件内容
	reader := csv.NewReader(sourceFile)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error reading CSV: %v", err)
	}

	// 打开或创建有效 CSV 文件
	_, validWriter, err := OpenOrCreateCSVFile(validFilePath)
	if err != nil {
		return nil, nil, nil, err
	}

	// 打开或创建无效 CSV 文件
	_, inValidWriter, err := OpenOrCreateCSVFile(invalidFilePath)
	if err != nil {
		return nil, nil, nil, err
	}

	return records, validWriter, inValidWriter, nil
}
