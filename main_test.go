package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testDir = "godev_test"
const mainFile = "main.go"
const godevBinary = "godev_binary"

func TestGoDevTool(t *testing.T) {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	t.Logf("Current working directory: %s", wd)

	// Compile godev
	err = compileGodev(wd)
	if err != nil {
		t.Fatalf("Failed to compile godev: %v", err)
	}
	defer os.Remove(filepath.Join(wd, godevBinary))

	// Setup test environment
	err = setupTestEnvironment(wd)
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}
	defer cleanupTestEnvironment(wd)

	// Run godev binary
	cmd := exec.Command(filepath.Join(wd, godevBinary), filepath.Join(testDir, mainFile))
	cmd.Dir = wd // Set the working directory for the command

	// Create a pipe for the output
	outputReader, outputWriter := io.Pipe()
	cmd.Stdout = outputWriter
	cmd.Stderr = outputWriter

	// Start reading the output in a separate goroutine
	outputChannel := make(chan string)
	go func() {
		scanner := bufio.NewScanner(outputReader)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("godev output:", line) // Log all output
			outputChannel <- line
		}
	}()

	err = cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start godev: %v", err)
	}
	defer cmd.Process.Kill()

	// Check if the program is running and producing output
	t.Log("Waiting for initial 'Hello, World!' output...")
	if !checkOutput(t, outputChannel, "Hello, World!", 30*time.Second) {
		t.Fatalf("Did not receive expected 'Hello, World!' output in time")
	}

	// Modify the main.go file
	t.Log("Modifying main.go file...")
	err = modifyMainFile(wd)
	if err != nil {
		t.Fatalf("Failed to modify main file: %v", err)
	}

	// Check if the program has been updated
	t.Log("Waiting for modified 'Hello, Modified World!' output...")
	if !checkOutput(t, outputChannel, "Hello, Modified World!", 30*time.Second) {
		t.Fatalf("Did not receive expected 'Hello, Modified World!' output in time")
	}

	// Check if the executable is cleaned up
	t.Log("Checking for executable cleanup...")
	checkExecutableCleanup(t, wd)
}

func compileGodev(wd string) error {
	cmd := exec.Command("go", "build", "-o", godevBinary, "main.go")
	cmd.Dir = wd
	return cmd.Run()
}

func setupTestEnvironment(wd string) error {
	testDirPath := filepath.Join(wd, testDir)
	err := os.Mkdir(testDirPath, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	initialContent := []byte(`
package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("Hello, World!")
		time.Sleep(time.Second)
	}
}
`)

	return ioutil.WriteFile(filepath.Join(testDirPath, mainFile), initialContent, 0644)
}

func modifyMainFile(wd string) error {
	modifiedContent := []byte(`
package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("Hello, Modified World!")
		time.Sleep(time.Second)
	}
}
`)

	return ioutil.WriteFile(filepath.Join(wd, testDir, mainFile), modifiedContent, 0644)
}

func checkOutput(t *testing.T, outputChannel <-chan string, expected string, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case output := <-outputChannel:
			if strings.Contains(output, expected) {
				t.Logf("Received expected output: %s", output)
				return true
			}
		case <-timer.C:
			t.Logf("Timeout reached while waiting for: %s", expected)
			return false
		}
	}
}

func checkExecutableCleanup(t *testing.T, wd string) {
	files, err := ioutil.ReadDir(filepath.Join(wd, testDir))
	if err != nil {
		t.Fatalf("Failed to read test directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || file.Name() == mainFile {
			continue
		}
		t.Errorf("Unexpected file found: %s", file.Name())
	}
}

func cleanupTestEnvironment(wd string) {
	os.RemoveAll(filepath.Join(wd, testDir))
}
