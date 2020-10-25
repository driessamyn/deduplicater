package maing

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"

	"github.com/driessamyn/deduplicater/internal/indexer"
)

func main() {
	parser := argparse.NewParser("duplicates", "Find and manage duplicate files")

	// general
	dirpath := parser.String("d", "dir", &argparse.Options{Required: true, Help: "Directory of files to use"})
	indexPath := parser.String("i", "index", &argparse.Options{Required: true, Help: "Path to the index file to create/use"})

	// index args
	indexCmd := parser.NewCommand("index", "Index allfiles")
	md5Flag := indexCmd.Flag("m", "md5", &argparse.Options{
		Required: false,
		Help:     "Use md5 hash",
		Default:  true,
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	if indexCmd.Happened() {
		fmt.Printf("Indexing %v to %v\n", *dirpath, *indexPath)

		index := indexer.NewIndexer(*dirpath, *indexPath, *md5Flag)
		err := index.Create()

		if nil != err {
			fmt.Printf("Failed creating index: %v", err)
		}
	}
}
