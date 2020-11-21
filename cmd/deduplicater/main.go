package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/manifoldco/promptui"
	"github.com/spf13/afero"

	"github.com/driessamyn/deduplicater/pkg/deduper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	parser := argparse.NewParser("duplicates", "Find and manage duplicate files")

	// general
	versionFlag := parser.Flag(
		"v", "version", &argparse.Options{
			Help: "Current version",
		},
	)

	indexPath := parser.String("i", "index", &argparse.Options{Required: false, Help: "Path to the index file to create/use"})
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

	deduper := deduper.NewDeduper(afero.NewOsFs(), *indexPath, *md5Flag)

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

		dupes := deduper.Find()
		if len(dupes) == 0 {
			fmt.Println("No duplicates found")
			return
		}

		fmt.Printf("%v duplicates found:\n", len(dupes))
		for _, v := range deduper.Find() {
			fmt.Printf("%v\n", v)
		}

		const CANCEL = "Do nothing"
		const DELETE = "Delete duplicates"
		const MOVE = "Move files to another folder"
		items := []string{CANCEL, DELETE, MOVE}

		prompt := promptui.Select{
			Label: "What do you want to do with the duplucates?",
			Items: items,
		}

		_, result, _ := prompt.Run()

		switch result {
		case DELETE:
			confirmPrompt := promptui.Prompt{
				Label:     "Are you sure you want to permanantly delete duplicates? (THIS CANNOT BE UNDONE)",
				IsConfirm: true,
			}

			if confirm, _ := confirmPrompt.Run(); "Y" == confirm {
				fmt.Printf("TODO: delete %q\n", result)
			}
		case MOVE:
			movePrompt := promptui.Prompt{
				Label:    "Location to copy duplicate files to",
				Validate: deduper.IsDirExist,
			}

			directory, _ := movePrompt.Run()
			err := deduper.MoveDuplicates(dupes, directory)
			if nil != err {
				fmt.Printf("Failed to move files: %v", err)
			}
		case CANCEL:
		default:
			return
		}

	case *versionFlag:
		fmt.Printf("deduplicater %v (%v - %v)", version, commit, date)
	}
}
