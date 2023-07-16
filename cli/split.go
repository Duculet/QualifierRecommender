package cli

import (
	"RecommenderServer/split"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func CommandSplit() *cobra.Command {

	// Setup flags
	var inputPath string
	var outputDir string
	var percentageTest int
	var randomSeed int64

	// Setup variables
	var err error

	cmdSplit := &cobra.Command{
		Use:   "split -i <file / directory path> -o <output directory> [-p <percentage for test>] [-r <random seed>]",
		Short: "Splits existing tsv file(s) into train and test files",
		Long: "Splits existing tsv file(s) into train and test files \n" +
			"The input <file> is expected to be a tsv file (or directory of tsvs). The output files will be tsv files. " +
			"The <output directory> is expected to be a directory where the output files will be written to. \n" +
			"The optional <percentage for test> is expected to be a number between 0 and 100, and indicates the percentage of the input file that will be used for the test file. " +
			"If not specified, the default value is 20.",
		Run: func(cmd *cobra.Command, args []string) {

			log.Println("Processing file / directory: ", inputPath)
			log.Println("Saving output to: ", outputDir)
			log.Println("Percentage for test: ", percentageTest)

			// Check if file is a directory
			// if so, split all files in the directory
			// if not, split the file
			if fileInfo, err := os.Stat(inputPath); err == nil && fileInfo.IsDir() {
				log.Println("Splitting all files in directory")
				split.SplitAll(inputPath, outputDir, percentageTest, randomSeed)
			} else {
				log.Println("Splitting file")
				split.SplitTrainTest(inputPath, outputDir, percentageTest, randomSeed)
			}
		},
	}

	cmdSplit.Flags().StringVarP(&inputPath, "input", "i", "", "Path to the input file or directory")
	cmdSplit.MarkFlagRequired("file")
	cmdSplit.Flags().StringVarP(&outputDir, "output", "o", "", "Path to the output directory")
	cmdSplit.MarkFlagRequired("output")
	cmdSplit.Flags().IntVarP(&percentageTest, "test", "p", 20, "Percentage of the input file that will be used for the test file")
	cmdSplit.Flags().Int64VarP(&randomSeed, "seed", "r", 0, "Random seed for shuffling the input file")

	err = cmdSplit.MarkFlagDirname("output")
	if err != nil {
		log.Panicln(err)
	}

	return cmdSplit
}
