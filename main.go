package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"log"
	"os"
	"strings"
	"archive/zip"
	"bufio"
	"io"
	"archive/tar"
)

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
		inputFileRC, err := zip.OpenReader(*inputFilePath)
		if err != nil {
			log.Fatalf("Cannot open input file: %v, error: %v", *inputFilePath, err)
		}
		defer inputFileRC.Close()

		flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		outputFile, err := os.OpenFile(*outputFilePath, flags, 0644)
		if err != nil {
			log.Fatalf("Cannot open output file: %v, error: %v", outputFilePath, err)
		}
		defer outputFile.Close()

		tw := tar.NewWriter(outputFile)
		defer tw.Close()

		for _, archivedFile := range inputFileRC.File {
			err := processArchivedFile(archivedFile, tw)
			if err != nil {
				break
			}
		}

		log.Println("Application finished")
	}
	app.Run(os.Args)
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
		//TODO implement transformInt
		processedContent, err = processLineByLine(file, transformString)
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
	var processedLines []string = make([]string, 0)
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

func transformString(line string) string {
	tokens := strings.Split(line, " ")
	return strings.Join(tokens, " ")
}

func noTransform(line string) string {
	return line
}