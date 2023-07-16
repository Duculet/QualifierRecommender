package cli

import (
	"RecommenderServer/extraction"
	"log"

	"github.com/spf13/cobra"
)

func CommandExtract() *cobra.Command {

	// Setup flags
	var dumpPath string
	var outputDir string
	var objectType bool
	var subjectType bool

	cmdExtract := &cobra.Command{
		Use:   "extract -f <dump file> -d <output directory> [-o=<object type>] [-s=<subject type>]",
		Short: "Extracts data about wikidata qualifiers",
		Long: "Extracts data about wikidata qualifiers from a bzip wikidata dump. \n" +
			"The input <dump file> is expected to be a json bzip2 compressed wikidata dump. The output files will be tsv files. " +
			"The <output directory> is expected to be a directory where the output files will be written to. " +
			"All output files will be named after the property they contain data about, namely '<property_id>.tsv'. \n" +
			"Optionally, the user can specify whether to also save object types and/or subject types." +
			"By default, both object types and subject types are saved (both =true). ",
		Run: func(cmd *cobra.Command, args []string) {

			log.Println("Processing dump: ", dumpPath)
			log.Println("Saving output to: ", outputDir)
			log.Println("Saving object types: ", objectType)
			log.Println("Saving subject types: ", subjectType)

			log.Println("First pass started")
			errE := extraction.FirstPass(dumpPath)
			if errE != nil {
				log.Panicln("Something went wrong during the first pass: ", errE)
			}

			log.Println("Second pass started")
			errE = extraction.SecondPass(dumpPath, outputDir, objectType, subjectType)
			if errE != nil {
				log.Panicln("Something went wrong during the second pass: ", errE)
			}
		},
	}

	cmdExtract.Flags().StringVarP(&dumpPath, "dump", "f", "", "Path to the wikidata dump")
	cmdExtract.MarkFlagRequired("dump")
	cmdExtract.Flags().StringVarP(&outputDir, "output", "d", "", "Path to the output directory")
	cmdExtract.MarkFlagRequired("output")
	cmdExtract.Flags().BoolVarP(&objectType, "objType", "o", true, "Save object types")
	cmdExtract.Flags().BoolVarP(&subjectType, "subjType", "s", true, "Save subject types")

	err := cmdExtract.MarkFlagFilename("dump", "bz2")
	if err != nil {
		log.Panicln(err)
	}
	err = cmdExtract.MarkFlagDirname("output")
	if err != nil {
		log.Panicln(err)
	}

	return cmdExtract
}
