package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type archiveProcessor struct {
	config      conf
	inputFileRC *zip.ReadCloser
	outputFile  *os.File
	tw          *tar.Writer
}

func (app *archiveProcessor) init() error {
	inputFileRC, err := zip.OpenReader(app.config.inputFilePath)
	if err != nil {
		return fmt.Errorf("Cannot open input file: %v, error: %v", app.config.inputFilePath, err)
	}
	app.inputFileRC = inputFileRC

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	outputFile, err := os.OpenFile(app.config.outputFilePath, flags, 0644)
	if err != nil {
		return fmt.Errorf("Cannot open output file: %v, error: %v", app.config.outputFilePath, err)
	}

	app.outputFile = outputFile
	app.tw = tar.NewWriter(outputFile)
	return nil
}

func (app *archiveProcessor) shutdown() {
	if app.inputFileRC != nil {
		app.inputFileRC.Close()
	}
	if app.tw != nil {
		app.tw.Close()
	}
	if app.outputFile != nil {
		app.outputFile.Close()
	}
}

func (app *archiveProcessor) process() error {
	for _, archivedFile := range app.inputFileRC.File {
		err := app.processArchivedFile(archivedFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *archiveProcessor) processArchivedFile(archivedFile *zip.File) error {
	log.Printf("Processing file: %v", archivedFile.Name)
	file, err := archivedFile.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	var processedContent string
	if strings.Contains(archivedFile.Name, "_integers_") {
		log.Printf("Archived file with integers: %v", archivedFile.Name)
		processedContent, err = processLineByLine(file, transformInt)
		if err != nil {
			return fmt.Errorf("Error by processing line by line: %v", err)
		}
		log.Printf("Processed content: %+v", processedContent)
	} else if strings.Contains(archivedFile.Name, "_strings_") {
		log.Printf("Archived file with strings: %v", archivedFile.Name)
		processedContent, err = processLineByLine(file, transformString)
		if err != nil {
			return fmt.Errorf("Error by processing line by line: %v", err)
		}
		log.Printf("Processed content: %+v", processedContent)
	} else {
		log.Printf("File is not eligible for transformation: %v", archivedFile.Name)
		processedContent, err = processLineByLine(file, noTransform)
		if err != nil {
			return fmt.Errorf("Error by processing line by line: %v", err)
		}
	}

	fileInfoHeader, err := tar.FileInfoHeader(archivedFile.FileInfo(), "")
	if err != nil {
		return fmt.Errorf("Cannot create tar header, error: %v", err)
	}

	fileInfoHeader.Size = int64(len(processedContent))
	err = app.tw.WriteHeader(fileInfoHeader)
	if err != nil {
		return fmt.Errorf("Cannot write tar header, error: %v", err)
	}

	_, err = app.tw.Write([]byte(processedContent))
	if err != nil {
		log.Printf("Cannot write tar content, error: %v", err)
	}

	return app.tw.Flush()
}

func processLineByLine(inputFile io.ReadCloser, handler func(string) string) (string, error) {
	processedLines := make([]string, 0)
	bufferedReader := bufio.NewReader(inputFile)
	for {
		lineContent, prefix, err := bufferedReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return strings.Join(processedLines, "\n"), nil
			}
			return "", err
		}
		if prefix {
			log.Printf("Line too big for buffer, only first %d bytes returned", len(lineContent))
		}
		processedLines = append(processedLines, handler(string(lineContent)))
	}
}
