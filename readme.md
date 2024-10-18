# godev: Go Development Hot-Reload Tool

`godev` is a command-line tool designed to streamline Go development by providing hot-reload functionality. It watches your project directory for changes in Go, HTML, CSS, and JavaScript files, automatically rebuilding and rerunning your Go application when changes are detected.

## Features

- Watches the current directory and subdirectories for file changes
- Automatically rebuilds and reruns your Go application on file changes
- Supports Go, HTML, CSS, and JavaScript file monitoring
- Gracefully terminates the previous process before starting a new one
- Easy to use with a simple command-line interface

## Installation

To install `godev`, follow these steps:

1. Ensure you have Go installed on your system (version 1.16 or later recommended).

2. Open a terminal and run the following command:

   ```
   go install github.com/shridarpatil/godev@latest
   ```

3. Make sure your Go bin directory (usually `~/go/bin/`) is in your system's PATH.

## Usage

To use `godev`, navigate to your Go project's root directory in the terminal and run:

```
godev path/to/your/main.go
```

Replace `path/to/your/main.go` with the relative or absolute path to your main Go file.

Example:

```
godev cmd/myapp/main.go
```

The tool will build and run your Go application, then watch for any changes in the current directory and its subdirectories. When a change is detected in a Go, HTML, CSS, or JavaScript file, `godev` will automatically rebuild and rerun your application.

## Configuration

Currently, `godev` doesn't require any configuration files. It uses sensible defaults for file watching and debounce intervals.

## Troubleshooting

If you encounter any issues:

1. Ensure you're running `godev` from your project's root directory.
2. Check that the specified main Go file path is correct.
3. Verify that you have write permissions in the current directory for building the Go binary.

## Contributing

Contributions to `godev` are welcome! Please feel free to submit pull requests, report bugs, or suggest features through the project's GitHub page.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Thanks to the creators and maintainers of the `fsnotify` package, which this tool relies on for file system notifications.
- Inspired by various hot-reload tools in other ecosystems, adapted for Go development workflows.

## Support

If you encounter any problems or have any questions, please open an issue on the GitHub repository.

Happy coding with `godev`!
