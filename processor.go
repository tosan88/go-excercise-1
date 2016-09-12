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
	"sync"
)

type archiveProcessor struct {
	config      conf
	inputFileRC *zip.ReadCloser
	outputFile  *os.File
	tw          *tar.Writer
	lk          sync.Mutex
}

type archivable struct {
	content      string
	archivedFile *zip.File
	writeHandler func(*archiveProcessor, archivable, chan<- error)
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
	errCh := make(chan error)
	archivableCh := make(chan archivable)

	var wg sync.WaitGroup
	wg.Add(len(app.inputFileRC.File))

	for _, archivedFile := range app.inputFileRC.File {
		go func(wg *sync.WaitGroup, archivedFile *zip.File) {
			defer wg.Done()
			app.processArchivedFile(archivedFile, errCh, archivableCh)
		}(&wg, archivedFile)
	}

	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(archivableCh)
	}(&wg)

	var writerWg sync.WaitGroup

	for {
		select {
		case err := <-errCh:
			return err
		case arch, ok := <-archivableCh:
			if !ok {
				writerWg.Wait()
				return app.tw.Flush()
			}
			writerWg.Add(1)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				app.lk.Lock()
				defer app.lk.Unlock()
				arch.writeHandler(app, arch, errCh)
			}(&writerWg)
		}
	}
}

func (app *archiveProcessor) processArchivedFile(archivedFile *zip.File, errCh chan error, archivableCh chan archivable) {
	log.Printf("Processing file: %v", archivedFile.Name)

	fileName := archivedFile.Name
	if strings.Contains(fileName, "_integers_") {
		log.Printf("Archived file with integers: %v", fileName)
		processedContent, err := processLineByLine(archivedFile, transformInt)
		if err != nil {
			errCh <- fmt.Errorf("Error by processing line by line: %v", err)
		}
		archivableCh <- archivable{
			content:      processedContent,
			archivedFile: archivedFile,
			writeHandler: handleProcessedContent,
		}

	} else if strings.Contains(fileName, "_strings_") {
		log.Printf("Archived file with strings: %v", fileName)
		processedContent, err := processLineByLine(archivedFile, transformString)
		if err != nil {
			errCh <- fmt.Errorf("Error by processing line by line: %v", err)
		}
		archivableCh <- archivable{
			content:      processedContent,
			archivedFile: archivedFile,
			writeHandler: handleProcessedContent,
		}
	} else {
		log.Printf("File is not eligible for transformation: %v", fileName)
		archivableCh <- archivable{
			archivedFile: archivedFile,
			writeHandler: handleUnprocessedContent,
		}
	}

}

func processLineByLine(archivedFile *zip.File, handler func(string) string) (string, error) {
	inputFile, err := archivedFile.Open()
	if err != nil {
		return "", err
	}
	defer inputFile.Close()

	var processedLines []string
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
			//how to handle it seriously?
			panic(fmt.Sprintf("Line too big for buffer, only first %d bytes returned", len(lineContent)))
		}
		processedLines = append(processedLines, handler(string(lineContent)))
	}
}

func (app *archiveProcessor) copyFile(archivedFile *zip.File, errCh chan<- error) {
	fileInfoHeader, err := tar.FileInfoHeader(archivedFile.FileInfo(), "")
	if err != nil {
		errCh <- fmt.Errorf("Cannot create tar header for %v, error: %v", archivedFile.Name, err)
	}
	err = app.tw.WriteHeader(fileInfoHeader)
	if err != nil {
		errCh <- fmt.Errorf("Cannot write tar header for %v, error: %v", archivedFile.Name, err)
	}
	file, err := archivedFile.Open()
	if err != nil {
		errCh <- err
	}
	defer file.Close()

	_, err = io.Copy(app.tw, file)
	if err != nil {
		errCh <- fmt.Errorf("Cannot copy tar content for %v, error: %v", archivedFile.Name, err)
	}
}

func (app *archiveProcessor) writeContent(archivedFile *zip.File, content string, errCh chan<- error) {
	fileInfoHeader, err := tar.FileInfoHeader(archivedFile.FileInfo(), "")
	if err != nil {
		errCh <- fmt.Errorf("Cannot create tar header for %v, error: %v", archivedFile.Name, err)
	}

	fileInfoHeader.Size = int64(len(content))
	err = app.tw.WriteHeader(fileInfoHeader)
	if err != nil {
		errCh <- fmt.Errorf("Cannot write tar header for %v, error: %v", archivedFile.Name, err)
	}

	_, err = app.tw.Write([]byte(content))
	if err != nil {
		errCh <- fmt.Errorf("Cannot write tar content for %v, error: %v", archivedFile.Name, err)
	}
}

func handleProcessedContent(app *archiveProcessor, arch archivable, errCh chan<- error) {
	app.writeContent(arch.archivedFile, arch.content, errCh)
}

func handleUnprocessedContent(app *archiveProcessor, arch archivable, errCh chan<- error) {
	app.copyFile(arch.archivedFile, errCh)
}
