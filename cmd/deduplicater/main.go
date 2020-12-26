package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/spf13/afero"

	"github.com/driessamyn/deduplicater/pkg/deduper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

type FindAction int

const (
	Unknown FindAction = iota
	Move               = iota
	Delete             = iota
)

func main() {
	run(os.Args)
}

func run(args []string) {
	parser := argparse.NewParser("duplicates", "Find and manage duplicate files")

	// general
	versionFlag := parser.Flag(
		"v", "version", &argparse.Options{
			Help: "Current version",
		},
	)

	indexPath := parser.String("f", "file", &argparse.Options{Required: false, Help: "Path to the index file to create/use"})
	md5Flag := parser.Flag("", "md5", &argparse.Options{
		Required: false,
		Help:     "Use md5 hash",
		Default:  false,
	})

	imageHashFlag := parser.Flag("", "imagehash", &argparse.Options{
		Required: false,
		Help:     "Use image hash",
		Default:  false,
	})

	// index
	indexCmd := parser.NewCommand("index", "Index allfiles")
	dirpath := indexCmd.String("d", "dir", &argparse.Options{Required: true, Help: "Directory of files to use"})

	// find
	findCmd := parser.NewCommand("find", "Find duplicates")
	deleteFlag := findCmd.Flag("", "remove", &argparse.Options{Required: false, Help: "Force remove duplicate files"})
	moveDir := findCmd.String("", "move-dir", &argparse.Options{Required: false, Help: "Directory to move the files to"})

	err := parser.Parse(args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	deduper := deduper.NewDeduper(afero.NewOsFs(), *indexPath, *md5Flag, *imageHashFlag)

	switch {
	case indexCmd.Happened():
		fmt.Printf("Indexing %v to %v\n", *dirpath, *indexPath)

		err := deduper.Create(*dirpath)

		if nil != err {
			fmt.Printf("Failed creating index: %v\n", err)
		}

	case findCmd.Happened():
		fmt.Printf("Finding duplicates in %v using 'md5 checksum'\n", *indexPath)

		err := deduper.Load()
		if nil != err {
			fmt.Printf("Failed loading index: %v\n", err)
		}

		dupes, err := deduper.Find()
		if nil != err {
			fmt.Printf("Failed finding duplicates: %v\n", err)
		}

		if len(dupes) == 0 {
			fmt.Println("No duplicates found")
			return
		}

		fmt.Printf("%v duplicates found:\n", len(dupes))
		for _, v := range dupes {
			fmt.Printf("%v\n", v)
		}

		var findAction FindAction
		if *deleteFlag {
			findAction = Delete
		} else if "" != *moveDir {
			findAction = Move
		} else {
			findAction, moveDir = PromptAction(deduper.IsDirExist)
		}

		if Move == findAction {
			err := deduper.MoveDuplicates(dupes, *moveDir)
			if nil != err {
				fmt.Printf("Failed to move files: %v", err)
			}
		} else if Delete == findAction {
			fmt.Println("TODO: delete")
		} else {
			fmt.Println("Do Nothing")
		}

	case *versionFlag:
		fmt.Printf("deduplicater %v (%v - %v)", version, commit, date)
	}
}
