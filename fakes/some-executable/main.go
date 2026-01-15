package main

import (
	"fmt"
	"io"
	"os"
)

var fail string

func main() {
	_, _ = fmt.Fprintf(os.Stdout, "Output on stdout\n")
	_, _ = fmt.Fprintf(os.Stderr, "Output on stderr\n")
	_, _ = fmt.Printf("Arguments: %v\n", os.Args)

	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, _ = fmt.Printf("Input on stdin\n%s\n", stdin)

	pwd, _ := os.Getwd()
	_, _ = fmt.Printf("PWD=%s\n", pwd)

	for _, env := range os.Environ() {
		_, _ = fmt.Printf("%s\n", env)
	}

	if fail == "true" {
		_, _ = fmt.Fprintf(os.Stdout, "Error on stdout\n")
		_, _ = fmt.Fprintf(os.Stderr, "Error on stderr\n")
		os.Exit(1)
	}
}
