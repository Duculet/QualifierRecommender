package cli

import (
	"RecommenderServer/evaluation"
	"RecommenderServer/schematree"
	"RecommenderServer/server"
	"RecommenderServer/strategy"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func CommandWikiEvaluate() *cobra.Command {

	var modelFile, testset, outputDir, handler string
	var verbose bool
	var model *schematree.SchemaTree
	var workflow *strategy.Workflow

	cmdEvalTree := &cobra.Command{
		Use:   "evaluate -m <modelfile> -d <testset> [-o <outputdir>] [-k <handler>] [-v <verbose>]",
		Short: "Evaluate the model against the testset",
		Long: "Evaluate the model against the testset. \n" +
			"The model should be a schematree binary file. \n" +
			"The testset should be a tsv file with one transaction per line. \n" +
			"The output file will be generated in the output directory and with suffixed names, namely:" +
			" '<testset>.eval.<handler>.tsv'\n" +
			"The handler should be the way of handling the transactions during evaluation.",
		Run: func(cmd *cobra.Command, args []string) {

			log.Println("Evaluating model", modelFile, "against testset", testset, "with handler", handler)

			// load the model and the workflow
			model = server.GetModel(modelFile)
			workflow = server.GetWorkflow("", model)

			// evaluate the model
			results := evaluation.EvaluateDataset(model, workflow, testset, 0, verbose)

			// write the results to the output file
			// remove the .tsv extension from the testset
			// this is the name of the output file
			if outputDir == "" {
				// if no output directory is provided, use the directory of the testset
				outputDir = filepath.Dir(testset)
			} else {
				// if an output directory is provided, use it
				outputDir = filepath.Clean(outputDir)
			}
			// create directory if it does not exist
			if _, err := os.Stat(outputDir); os.IsNotExist(err) {
				err = os.MkdirAll(outputDir, os.ModePerm)
				if err != nil {
					log.Panicln(err, "Could not create output directory", outputDir)
				}
			}

			// the output file will be generated in the output directory
			testsetID := strings.TrimSuffix(filepath.Base(testset), ".tsv")
			outputFileName := filepath.Clean(filepath.Join(outputDir, testsetID+".eval."+handler))

			// write the results to the output file
			outputFileName = evaluation.WriteResultsToFile(outputFileName, results)

			log.Println("Results:", results)
			log.Println("Results written to", outputFileName)

			log.Println("EVALUATION FINISHED!")
		},
	}

	cmdEvalTree.Flags().StringVarP(&modelFile, "model", "m", "", "The model to evaluate")
	cmdEvalTree.Flags().StringVarP(&testset, "testset", "d", "", "The testset to evaluate against")
	cmdEvalTree.Flags().StringVarP(&outputDir, "output", "o", "", "The output directory to write the results to")
	cmdEvalTree.Flags().StringVarP(&handler, "handler", "k", "takeOneButType", "The handler to use for evaluation: takeOneButType")
	cmdEvalTree.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmdEvalTree
}
