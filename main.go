package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"log"
	"os"
	"strings"
	"time"
)

type conf struct {
	inputFilePath  string
	outputFilePath string
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
		defer func(start time.Time) {
			elapsed := time.Since(start)
			log.Printf("Application finished. Took %v seconds", elapsed.Seconds())
		}(time.Now())

		if err := validateInputFile(*inputFilePath); err != nil {
			log.Fatalf("%v", err)
			return
		}
		if err := validateOutputFile(*outputFilePath); err != nil {
			log.Fatalf("%v", err)
			return
		}

		processingApp := &archiveProcessor{config: conf{
			inputFilePath:  *inputFilePath,
			outputFilePath: *outputFilePath,
		}}

		err := processingApp.init()
		defer processingApp.shutdown()
		if err != nil {
			log.Fatalf("%v", err)
			return
		}

		if err := processingApp.process(); err != nil {
			log.Fatalf("%v", err)
		}

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
