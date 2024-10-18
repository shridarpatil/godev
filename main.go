package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	watchedExtensions = []string{".go", ".html", ".css", ".js"}
	lastReload        time.Time
	debounceInterval  = 100 * time.Millisecond
	cmd               *exec.Cmd
	cmdMutex          sync.Mutex
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: godev <path-to-your-go-file>")
		os.Exit(1)
	}

	targetFile := os.Args[1]
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if shouldReload(event.Name) {
						fmt.Println("Modified file:", event.Name)
						debounceReload(targetFile)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Watching directory: %s\n", dir)
	fmt.Printf("Building and running file: %s\n", targetFile)
	buildAndRun(targetFile)

	<-done
}

func shouldReload(filename string) bool {
	ext := filepath.Ext(filename)
	for _, watchedExt := range watchedExtensions {
		if ext == watchedExt {
			return true
		}
	}
	return false
}

func debounceReload(file string) {
	if time.Since(lastReload) > debounceInterval {
		lastReload = time.Now()
		buildAndRun(file)
	}
}

func buildAndRun(file string) {
	fmt.Println("Building...")
	buildCmd := exec.Command("go", "build", file)
	buildCmd.Stderr = os.Stderr
	err := buildCmd.Run()
	if err != nil {
		fmt.Printf("Build failed: %v\n", err)
		return
	}

	fmt.Println("Running...")
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	// Terminate the previous process if it's still running
	if cmd != nil && cmd.Process != nil {
		fmt.Println("Terminating previous process...")
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			fmt.Printf("Failed to interrupt process: %v\n", err)
			if err := cmd.Process.Kill(); err != nil {
				fmt.Printf("Failed to kill process: %v\n", err)
			}
		}
		_, _ = cmd.Process.Wait()
	}

	execName := strings.TrimSuffix(filepath.Base(file), ".go")
	cmd = exec.Command("./" + execName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Failed to start the program: %v\n", err)
		return
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			if err.Error() != "signal: interrupt" {
				fmt.Printf("Program exited with error: %v\n", err)
			} else {
				fmt.Println("Program terminated for rebuild")
			}
		} else {
			fmt.Println("Program exited successfully")
		}
	}()
}
