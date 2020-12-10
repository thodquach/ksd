package main

import (
	"fmt"
	"os"

	"ksd/filehandler"
)

func main() {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (info.Mode()&os.ModeCharDevice) != 0 || info.Size() < 0 {
		fmt.Fprintln(os.Stderr, "Command is intended to work with pipes.")
		fmt.Fprintln(os.Stderr, "Usage: kubectl get secret <secret-name> -o <yaml|json> |", os.Args[0])
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "< secret.<yaml|json>")
		os.Exit(1)
	}

	FileHandler := filehandler.NewFileHandler()

	stdin := FileHandler.Read(os.Stdin)
	output, err := FileHandler.Parse(stdin)

	if err != nil {
		fmt.Fprintf(os.Stderr, "could not decode secret: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stdout, string(output))
}
