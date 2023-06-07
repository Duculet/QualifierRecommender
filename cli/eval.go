package cli

import (
	"RecommenderServer/schematree"
	"RecommenderServer/server"
	"RecommenderServer/strategy"
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func CommandWikiEvaluate() *cobra.Command {

	var modelFile, dataset, outputDir, handler string
	var verbose bool
	var model *schematree.SchemaTree
	var workflow *strategy.Workflow

	cmdEvalTree := &cobra.Command{
		Use:   "evaluate -m <modelfile> -d <testset> [-o <outputdir>] [-h <handler>] [-v <verbose>]",
		Short: "Evaluate the model against the dataset",
		Long: "Evaluate the model against the dataset. \n" +
			"The model should be a schematree binary file. \n" +
			"The dataset should be a tsv file with one transaction per line. \n" +
			"The output file will be generated in the output directory and with suffixed names, namely:" +
			" '<dataset>.eval.<handler>.tsv'\n" +
			"The handler should be the way of handling the transactions during evaluation.",
		Run: func(cmd *cobra.Command, args []string) {

			log.Println("Evaluating model", model, "against dataset", dataset, "with handler", handler)

			// remove the .tsv extension from the dataset
			// this is the name of the output file
			if outputDir == "" {
				// if no output directory is provided, use the directory of the dataset
				outputDir = filepath.Dir(dataset)
			} else {
				// if an output directory is provided, use it
				outputDir = filepath.Clean(outputDir)
			}
			// create directory if it does not exist
			if _, err := os.Stat(outputDir); os.IsNotExist(err) {
				os.MkdirAll(outputDir, os.ModePerm)
			}
			// the output file will be generated in the output directory
			datasetID := strings.TrimSuffix(filepath.Base(dataset), ".tsv")
			outputFileName := filepath.Join(outputDir, datasetID+".eval."+handler+".tsv")
			outputFileName = filepath.Clean(outputFileName)

			// load the model
			model = server.GetModel(modelFile)
			workflow = server.GetWorkflow("", model)

			// cache recommendations without any information
			// this is the baseline
			t0 := time.Now()
			instanceAll := schematree.NewInstanceFromInput([]string{}, []string{}, model, true)
			recommendationsAll := workflow.Recommend(instanceAll)

			if verbose {
				log.Println("Recommendations", len(recommendationsAll), "for all found in", time.Since(t0))
			}

			// load the dataset
			datasetFile, err := os.Open(dataset)
			if err != nil {
				log.Panicln(err)
			}
			defer datasetFile.Close()

			// setup variables
			limitScan := 10000000
			limitRecs := -1
			avgRanks := make([]float32, 0)

			tsvScanner := bufio.NewScanner(datasetFile)
			for cntScan := 0; tsvScanner.Scan() && cntScan < limitScan; cntScan++ {
				line := tsvScanner.Text()

				items := strings.Split(line, "\t")
				var qualifiers, objTypes, subjTypes []string
				for _, item := range items {
					if strings.HasPrefix(item, "P") {
						qualifiers = append(qualifiers, item)
					} else if strings.HasPrefix(item, "o/") {
						// remove the o/ prefix
						item = strings.TrimPrefix(item, "o/")
						objTypes = append(objTypes, item)
					} else if strings.HasPrefix(item, "s/") {
						// remove the s/ prefix
						item = strings.TrimPrefix(item, "s/")
						subjTypes = append(subjTypes, item)
					} else {
						log.Panicln("Unknown item", item)
					}
				}

				// cache recommendations without any qualifier information
				t0 := time.Now()
				types := append(objTypes, subjTypes...)
				instanceTypes := schematree.NewInstanceFromInput([]string{}, types, model, true)
				recommendationsTypes := workflow.Recommend(instanceTypes)
				if verbose {
					log.Println("Types:", types)
					log.Println("Recommendations", len(recommendationsTypes), "for types found in", time.Since(t0))

					log.Println("###################")
					log.Println("Evaluating line", line)
					log.Println("Qualifiers:", qualifiers)
					log.Println(len(objTypes), "Object types:", objTypes)
					log.Println(len(subjTypes), "Subject types:", subjTypes)
				}

				// remove the first item from the qualifiers and add it to the leftOut
				// skip if there is only one qualifier
				// MUST DO THIS DIFFERENT FOR TYPED VERSION
				// if len(qualifiers)-1 == 0 {
				// 	continue
				// }

				sumTransRanks := 0

				// run for each transaction (line), take one out and evaluate
				// CAN MAKE IT RANGE
				for idx, leftOut := range qualifiers {
					// reducedSet := make([]string, len(qualifiers))
					// c := copy(reducedSet, qualifiers)
					reducedSet := append([]string{}, qualifiers...)
					// log.Println("Copied", c, "qualifiers")
					reducedSet = append(reducedSet[:idx], reducedSet[idx+1:]...)
					// }
					// for idx := 0; idx < len(qualifiers); idx++ {
					// 	leftOut := qualifiers[idx]
					// 	// make a reduced set of qualifiers without the left out qualifier
					// 	// reducedSet := append(qualifiers[:idx], qualifiers[idx+1:]...)
					// 	reducedSet := make([]string, 0, len(qualifiers)-1)
					// 	for _, item := range qualifiers {
					// 		// use index
					// 		if item != leftOut {
					// 			reducedSet = append(reducedSet, item)
					// 		}
					// 	}
					if verbose {
						log.Println("")
						log.Println("Initial set of qualifiers:", qualifiers)
						log.Println("Length of initial set of qualifiers:", len(qualifiers))
						log.Println("Reduced set of qualifiers:", reducedSet)
						log.Println("Evaluating with left out qualifier", leftOut)
						log.Println("-------------------")
					}

					instance := schematree.NewInstanceFromInput(reducedSet, types, model, true)

					t1 := time.Now()
					recommendation := workflow.Recommend(instance)
					if verbose {
						log.Println("Recommendations", len(recommendation))
						log.Println("Recommendation took", time.Since(t1))
					}

					existsRec := make(map[string]bool)
					outputRecs := make([]string, 0)
					for _, item := range recommendation {
						if len(outputRecs) >= limitRecs && limitRecs != -1 {
							break
						}
						if item.Property.IsQualifier() {
							outputRecs = append(outputRecs, *item.Property.Str)
							existsRec[*item.Property.Str] = true // add to the map
						}
					}

					// concatenate recommendationsTypes and outputRecs
					// only add the ones that are not already in the list
					// this is to avoid duplicates
					for _, item := range recommendationsTypes {
						if len(outputRecs) >= limitRecs && limitRecs != -1 {
							break
						}
						if _, exists := existsRec[*item.Property.Str]; !exists && item.Property.IsQualifier() {
							outputRecs = append(outputRecs, *item.Property.Str)
							existsRec[*item.Property.Str] = true // add to the map
						}
					}

					// concatenate recommendationsAll and outputRecs
					// only add the ones that are not already in the list
					for _, item := range recommendationsAll {
						if len(outputRecs) >= limitRecs && limitRecs != -1 {
							break
						}
						if _, exists := existsRec[*item.Property.Str]; !exists && item.Property.IsQualifier() {
							outputRecs = append(outputRecs, *item.Property.Str)
							existsRec[*item.Property.Str] = true // add to the map
						}
					}

					// check if the recommendation contains the leftOut qualifier
					containsLeftOut := false
					rankLeftOut := 500
					for r, item := range outputRecs {
						if verbose {
							log.Println("Checking", item, "against", leftOut)
						}
						if item == leftOut {
							containsLeftOut = true
							rankLeftOut = r
							break
						}
					}

					if verbose {
						log.Println(len(outputRecs), "recommendations full info")
						log.Println("Recommendations full info:", outputRecs)
						log.Println(len(outputRecs), "recommendations full info after adding types")
						log.Println(len(outputRecs), "recommendations full info after adding types and others")
						log.Println("Left out qualifier", leftOut, "is contained in the recommendation:", containsLeftOut)
						log.Println("Rank of left out qualifier", leftOut, "is", rankLeftOut)
					}
					// if !containsLeftOut {
					// 	// wait for 5 seconds
					// 	time.Sleep(5 * time.Second)
					// }

					// add the rank to the sum
					sumTransRanks += rankLeftOut
				}

				// calculate the average rank
				avgRank := float32(sumTransRanks) / float32(len(qualifiers))

				if verbose {
					log.Println("AVG RANK:", avgRank)
				}

				// add the average rank to the list
				avgRanks = append(avgRanks, avgRank)
			}

			// calculate the average of the average ranks
			sumAvgRanks := float32(0)
			for _, avgRank := range avgRanks {
				sumAvgRanks += avgRank
			}
			modelAvgRank := sumAvgRanks / float32(len(avgRanks))

			log.Println("Model evaluated:", strings.TrimSuffix(filepath.Base(modelFile), ".tsv.schemaTree.typed.pb"))
			log.Println("Model avg rank:", modelAvgRank)

			log.Println("Writing to file", outputFileName)

			// write the output to a file
			outputFile, err := os.Create(outputFileName)
			if err != nil {
				log.Panicln(err)
			}
			defer outputFile.Close()

			// write the header
			outputFile.WriteString("model\tmodelAvgRank\n")
			// write the model name and the average rank
			outputFile.WriteString(strings.TrimSuffix(filepath.Base(modelFile), ".tsv.schemaTree.typed.pb") + "\t" + fmt.Sprintf("%f", modelAvgRank) + "\n")

			log.Println("EVALUATION FINISHED!")
		},
	}

	cmdEvalTree.Flags().StringVarP(&modelFile, "model", "m", "", "The model to evaluate")
	cmdEvalTree.Flags().StringVarP(&dataset, "dataset", "d", "", "The dataset to evaluate against")
	cmdEvalTree.Flags().StringVarP(&outputDir, "output", "o", "", "The output directory to write the results to")
	cmdEvalTree.Flags().StringVarP(&handler, "handler", "k", "takeOneButType", "The handler to use for evaluation: takeOneButType")
	cmdEvalTree.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmdEvalTree
}
