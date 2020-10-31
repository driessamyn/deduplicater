package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"

	"github.com/driessamyn/deduplicater/pkg/deduper"
)

func main() {
	parser := argparse.NewParser("duplicates", "Find and manage duplicate files")

	// general
	indexPath := parser.String("i", "index", &argparse.Options{Required: true, Help: "Path to the index file to create/use"})
	md5Flag := parser.Flag("m", "md5", &argparse.Options{
		Required: false,
		Help:     "Use md5 hash",
		Default:  true,
	})

	// index
	indexCmd := parser.NewCommand("index", "Index allfiles")
	dirpath := indexCmd.String("d", "dir", &argparse.Options{Required: true, Help: "Directory of files to use"})

	// find
	findCmd := parser.NewCommand("find", "Find duplicates")

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	deduper := deduper.NewDeduper(*indexPath, *md5Flag)

	if indexCmd.Happened() {
		fmt.Printf("Indexing %v to %v\n", *dirpath, *indexPath)

		err := deduper.Create(*dirpath)

		if nil != err {
			fmt.Printf("Failed creating index: %v\n", err)
		}
	}

	if findCmd.Happened() {
		fmt.Printf("Finding duplicates in %v using 'md5 checksum'\n", *indexPath)

		err := deduper.Load()
		if nil != err {
			fmt.Printf("Failed loading index: %v\n", err)
		}

		dupes := deduper.Find()
		if len(dupes) == 0 {
			fmt.Println("No duplicates found")
			return
		}

		fmt.Printf("%v duplicates found:\n", len(dupes))
		for _, v := range deduper.Find() {
			fmt.Printf("%v\n", v)
		}
	}
}
