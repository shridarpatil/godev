package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	watchedExtensions = []string{".go", ".html", ".css", ".js"}
	lastReload        time.Time
	debounceInterval  = 100 * time.Millisecond
	cmd               *exec.Cmd
	cmdMutex          sync.Mutex
	executableName    string
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("godev output: Usage: godev <path-to-your-go-file>")
		os.Exit(1)
	}

	fmt.Println(`
	 ██████╗  ██████╗ ██████╗ ███████╗██╗   ██╗
	██╔════╝ ██╔═══██╗██╔══██╗██╔════╝██║   ██║
	██║  ███╗██║   ██║██║  ██║█████╗  ██║   ██║
	██║   ██║██║   ██║██║  ██║██╔══╝  ╚██╗ ██╔╝
	╚██████╔╝╚██████╔╝██████╔╝███████╗ ╚████╔╝ 
	 ╚═════╝  ╚═════╝ ╚═════╝ ╚══════╝  ╚═══╝  
	                                           
	Running in godev mode - Hot Reload Activated
	`)

	targetFile := os.Args[1]
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	executableName = "godev"

	// Set up clean up on interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()

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
						fmt.Println("godev output: Modified file:", event.Name)
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

	fmt.Printf("godev output: Watching directory: %s\n", dir)
	fmt.Printf("godev output: Building and running file: %s\n", targetFile)
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
	cleanup() // Remove previous build before rebuilding

	fmt.Println("godev output: Building...")
	buildCmd := exec.Command("go", "build", "-o", executableName, file)
	buildCmd.Stderr = os.Stderr
	err := buildCmd.Run()
	if err != nil {
		fmt.Printf("Build failed: %v\n", err)
		return
	}

	fmt.Println("godev output: Running...")
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	// Terminate the previous process if it's still running
	if cmd != nil && cmd.Process != nil {
		fmt.Println("godev output: Terminating previous process...")
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			fmt.Printf("godev output: Failed to interrupt process: %v\n", err)
			if err := cmd.Process.Kill(); err != nil {
				fmt.Printf("godev output: Failed to kill process: %v\n", err)
			}
		}
		_, _ = cmd.Process.Wait()
	}

	cmd = exec.Command("./" + executableName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		fmt.Printf("godev output: Failed to start the program: %v\n", err)
		return
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			if err.Error() != "signal: interrupt" {
				fmt.Printf("godev output: Program exited with error: %v\n", err)
			} else {
				fmt.Println("godev output: Program terminated for rebuild")
			}
		} else {
			fmt.Println("godev output: Program exited successfully")
		}
	}()
}

func cleanup() {
	if executableName == "" {
		return
	}

	err := os.Remove(executableName)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("godev output: Failed to remove executable: %v\n", err)
	} else if err == nil {
		fmt.Printf("godev output: Removed executable: %s\n", executableName)
	}
}
