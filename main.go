package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"fmt"
	"github.com/jawher/mow.cli"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type conf struct {
	inputFilePath  string
	outputFilePath string
}

type archiveProcessor struct {
	config      conf
	inputFileRC *zip.ReadCloser
	outputFile  *os.File
	tw          *tar.Writer
}

type processor interface {
	init() error
	shutdown()
	process() error
}

func main() {
	log.Printf("Application starting with args %s", os.Args)
	app := cli.App("go-exercise-1",
		"Enumerate files in a zip archive, transform the files content with respect to certain rules and write the transformed files to a tar archive")

	inputFilePath := app.String(cli.StringOpt{
		Name:  "input-file",
		Value: "",
		Desc:  "The input archive file path",
	})

	outputFilePath := app.String(cli.StringOpt{
		Name:  "output-file",
		Value: "",
		Desc:  "The output archive file path",
	})

	app.Action = func() {
		if err := validateInputFile(*inputFilePath); err != nil {
			log.Fatalf("%v", err)
			return
		}
		if err := validateOutputFile(*outputFilePath); err != nil {
			log.Fatalf("%v", err)
			return
		}
		processingApp := &archiveProcessor{config: conf{inputFilePath: *inputFilePath, outputFilePath: *outputFilePath}}

		err := processingApp.init()
		defer processingApp.shutdown()
		if err != nil {
			log.Fatalf("%v", err)
			return
		}
		if err := processingApp.process(); err != nil {
			log.Fatalf("%v", err)
		}

		log.Println("Application finished")
	}
	app.Run(os.Args)
}

func (app *archiveProcessor) process() error {
	for _, archivedFile := range app.inputFileRC.File {
		err := processArchivedFile(archivedFile, app.tw)
		if err != nil {
			return err
		}
	}

	return nil
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

func validateInputFile(inputFilePath string) error {
	if inputFilePath == "" {
		return fmt.Errorf("Input file path is empty")
	}
	if !strings.HasSuffix(inputFilePath, ".zip") {
		return fmt.Errorf("Input file is not a zip file")
	}

	return nil
}

func validateOutputFile(outputFilePath string) error {
	if outputFilePath == "" {
		return fmt.Errorf("Output file path is empty")
	}
	if !strings.HasSuffix(outputFilePath, ".tar") {
		return fmt.Errorf("Output file is not a tar file")
	}

	return nil
}

func processArchivedFile(archivedFile *zip.File, tw *tar.Writer) error {
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
			log.Printf("Error by processing line by line: %v", err)
			return err
		}
		log.Printf("Processed content: %+v", processedContent)
	} else if strings.Contains(archivedFile.Name, "_strings_") {
		log.Printf("Archived file with strings: %v", archivedFile.Name)
		processedContent, err = processLineByLine(file, transformString)
		if err != nil {
			log.Printf("Error by processing line by line: %v", err)
			return err
		}
		log.Printf("Processed content: %+v", processedContent)
	} else {
		log.Printf("File is not eligible for transformation: %v", archivedFile.Name)
		processedContent, err = processLineByLine(file, noTransform)
		if err != nil {
			log.Printf("Error by processing line by line: %v", err)
			return err
		}
	}

	fileInfoHeader, err := tar.FileInfoHeader(archivedFile.FileInfo(), "")
	if err != nil {
		log.Printf("Cannot create tar header, error: %v", err)
		return err
	}

	fileInfoHeader.Size = int64(len(processedContent))
	err = tw.WriteHeader(fileInfoHeader)
	if err != nil {
		log.Printf("Cannot write tar header, error: %v", err)
		return err
	}

	_, err = tw.Write([]byte(processedContent))
	if err != nil {
		log.Printf("Cannot write tar content, error: %v", err)
	}
	tw.Flush()

	return nil
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

func transformInt(line string) string {
	tokens := strings.Split(line, " ")
	var transformedTokens []string
	for _, token := range tokens {
		num, err := strconv.Atoi(token)
		if err != nil {
			transformedTokens = append(transformedTokens, token)
		} else {
			transformedTokens = append(transformedTokens, strconv.Itoa(num+123))
		}
	}

	return strings.Join(transformedTokens, " ")
}

func transformString(line string) string {
	tokens := strings.Split(line, " ")
	var transformedTokens []string
	for _, token := range tokens {
		size := len(token)
		var reversed []rune = make([]rune, size)
		for i, ch := range token {
			if unicode.IsLower(ch) {
				reversed[size-i-1] = unicode.ToUpper(ch)
			} else if unicode.IsUpper(ch) {
				reversed[size-i-1] = unicode.ToLower(ch)
			} else {
				reversed[size-i-1] = ch
			}
		}
		transformedTokens = append(transformedTokens, string(reversed))
	}
	return strings.Join(transformedTokens, " ")
}

func noTransform(line string) string {
	return line
}
