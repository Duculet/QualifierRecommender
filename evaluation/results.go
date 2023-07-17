package evaluation

import (
	"RecommenderServer/schematree"
	"RecommenderServer/strategy"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type modelResult struct {
	ModelID      string            // model ID
	EvalCount    int64             // number of pairs evaluated
	EvalTime     int64             // time taken to evaluate all pairs
	QualifierPop map[string]uint64 // qualifier popularity (number of times it appears in the test set)
	EvalResults  []evalResult      // evaluation results
}

type evalResult struct {
	TransID      uint32 // transaction ID (unique for each transaction)
	SetSize      uint16 // number of properties used to generate recommendations (both type and non-type)
	LeftOut      string // qualifier that has been left out
	NumTypes     uint16 // number of type properties
	NumObjTypes  uint16 // number of object type properties
	NumSubjTypes uint16 // number of object type properties
	Rank         uint16 // rank calculated for recommendation, equal to 5843 if not fully recommendated back
	HitsAt1      uint8  // rank hits@1, value is 1 if leftOut is first in recommendations, 0 otherwise
	HitsAt5      uint8  // rank hits@5, value is 1 if leftOut is in the first 5 recommendations, 0 otherwise
	HitsAt10     uint8  // rank hits@10, value is 1 if leftOut is in the first 10 recommendations, 0 otherwise
}

// Setup variables for evaluation
var MODEL *schematree.SchemaTree
var WORKFLOW *strategy.Workflow
var allRecs, typeRecs, fullRecs schematree.PropertyRecommendations
var qualifierPop map[string]uint64
var evalCount, evalTime int64
var VERBOSE bool

// Evaluate will run the evaluation process for the given model, workflow, testset, and handler.
// The results will be returned as a list of evaluation results, one for each transaction.
func EvaluateDataset(
	model *schematree.SchemaTree,
	workflow *strategy.Workflow,
	testset, handlerName string,
	verbose bool,
) (results []evalResult) {

	// save to global variables
	MODEL, WORKFLOW, VERBOSE = model, workflow, verbose
	qualifierPop = make(map[string]uint64)

	evalStart := time.Now()

	// cache recommendations without any information
	// this is the baseline
	instanceAll := schematree.NewInstanceFromInput([]string{}, []string{}, model, true)
	allRecs = workflow.Recommend(instanceAll)

	if verbose {
		log.Println("All recommendations", allRecs, "\n length", len(allRecs))
	}

	// depending on evaluation method, use different handlers
	var handler handlerFunc
	switch handlerName {
	case "takeOneButType":
		handler = takeOneButType
	case "takeAllButType":
		handler = takeAllButType
	case "baseline":
		handler = baseline
	default:
		log.Panicln("Unknown handler", handlerName)
	}

	// construct method that will be used to evaluate each transaction
	evaluator := func(reduced []string, leftout string) evalResult {
		evalCount++
		qualifierPop[leftout]++
		return evaluatePair(model, workflow, reduced, leftout)
	}

	// build the callback function for the property summary reader
	// given a transaction summary, it will use the handler to split
	// the transaction into a reduced and a leftout set
	// and then evaluate the reduced set
	transactionCallback := func(summary transactionSummary) {
		newResults := handler(summary, evaluator)
		results = append(results, newResults...)
	}

	// start the property summary reader
	numTrans := propertySummaryReader(testset, transactionCallback)

	evalTime = time.Since(evalStart).Nanoseconds()

	if verbose {
		log.Println("Number of transactions:", numTrans)
		log.Println("It took", evalTime, "nanoseconds to evaluate", evalCount, "pairs.")
		log.Println("Average time per pair:", evalTime/evalCount, "nanoseconds.")
		log.Println("Evaluation done.")
	}
	return
}

// evaluatePair will evaluate a pair made of a reduced set of qualifiers / properties and a left out qualifier
func evaluatePair(
	model *schematree.SchemaTree,
	workflow *strategy.Workflow,
	reducedSet []string,
	leftOut string,
) evalResult {

	// get recommendation only if the reduced set is not empty
	if len(reducedSet) != 0 {
		types := []string{} // the types are included in the reduced set (done by handler)
		instance := schematree.NewInstanceFromInput(reducedSet, types, model, true)
		fullRecs = workflow.Recommend(instance)
	}

	if VERBOSE {
		log.Println("Full recommendations", fullRecs, "\n length", len(fullRecs))
	}

	// get the final recommendations
	// concatenate all recommendations, remove duplicates, and keep only qualifiers
	_, outputRecs := finalRecommendations(fullRecs, typeRecs, allRecs, -1)

	// get the rank of the left out qualifier
	rankLeftOut := 5843 // if not found, set to predefined value
	for idx, item := range outputRecs {
		if VERBOSE {
			log.Println("Checking", item, "against", leftOut)
		}
		if item == leftOut {
			rankLeftOut = idx + 1
			break
		}
	}

	var hitsAt1, hitsAt5, hitsAt10 uint8
	// set the rank hits @1, @5, @10
	if rankLeftOut == 1 {
		hitsAt1 = 1
	}
	if rankLeftOut <= 5 {
		hitsAt5 = 1
	}
	if rankLeftOut <= 10 {
		hitsAt10 = 1
	}

	if VERBOSE {
		log.Println(len(outputRecs), "recommendations full info")
		log.Println("Recommendations full info:", outputRecs)
		log.Println("Rank of left out qualifier", leftOut, "is", rankLeftOut)
		log.Println("Rank @1:", hitsAt1)
		log.Println("Rank @5:", hitsAt5)
		log.Println("Rank @10:", hitsAt10)
	}

	return evalResult{
		SetSize:  uint16(len(reducedSet)),
		LeftOut:  leftOut,
		Rank:     uint16(rankLeftOut),
		HitsAt1:  hitsAt1,
		HitsAt5:  hitsAt5,
		HitsAt10: hitsAt10,
	}
}

// WriteResultsToFile will output the entire evalResult array to a json file
// also add the model results to the file (if any)
// and return the name of the file
func WriteResultsToFile(filename string, evalResults []evalResult, compress bool) (outputfile string) {

	outputfile = filename + ".json"

	// create the model result
	// modelID should be the basename of the model file (without the extension)
	modelID := strings.Split(filepath.Base(filename), ".")[0]
	modelResult := modelResult{
		ModelID:      modelID,
		QualifierPop: qualifierPop,
		EvalResults:  evalResults,
		EvalCount:    evalCount,
		EvalTime:     evalTime,
	}

	f, err := os.Create(outputfile)
	if err != nil {
		log.Fatalln("Could not create / open .json file")
	}
	defer f.Close()

	if compress {
		g := gzip.NewWriter(f)
		defer g.Close()

		e := json.NewEncoder(g)

		// write the results
		err = e.Encode(modelResult)
		if err != nil {
			log.Panicln("Could not encode results to json")
		}
	} else {
		w := bufio.NewWriter(f)
		defer w.Flush()

		// write the results
		out, err := json.MarshalIndent(modelResult, "", "  ")
		if err != nil {
			log.Panicln("Could not marshal results to json")
		}

		_, err = w.Write(out)
		if err != nil {
			log.Panicln("Could not write results to file")
		}
	}

	return
}
