package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

const paritySuiteName = "TestOpenAPIParitySuite"

type goTestEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
}

func main() {
	command := exec.Command(
		"go",
		"test",
		"./internal/app",
		"-run", "^"+paritySuiteName+"$",
		"-count=1",
		"-json",
	)
	output, runErr := command.CombinedOutput()
	_, _ = os.Stdout.Write(output)
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "OpenAPI parity suite failed: %v\n", runErr)
		os.Exit(1)
	}
	if err := verifyParityTestOutput(output); err != nil {
		fmt.Fprintf(os.Stderr, "OpenAPI parity gate rejected the test result: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("verified OpenAPI parity suite: %s\n", paritySuiteName)
}

func verifyParityTestOutput(output []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	namedSuitePassed := false
	packagePassed := false
	for scanner.Scan() {
		var event goTestEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		if event.Action == "pass" && event.Test == paritySuiteName {
			namedSuitePassed = true
		}
		if event.Action == "pass" && event.Test == "" && event.Package != "" {
			packagePassed = true
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read go test JSON: %w", err)
	}
	if !namedSuitePassed {
		return errors.New("go test completed without a passing TestOpenAPIParitySuite event")
	}
	if !packagePassed {
		return errors.New("OpenAPI parity package did not report a passing result")
	}
	return nil
}
