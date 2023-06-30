package evaluation

import (
	"RecommenderServer/schematree"
	"RecommenderServer/strategy"
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

type modelResult struct {
	EvalCount uint64 // number of pairs evaluated
	EvalTime  int64  // time taken to evaluate all pairs
}

type evalResult struct {
	TransID      uint32   // transaction ID (unique for each transaction)
	SetSize      uint16   // number of properties used to generate recommendations (both type and non-type)
	LeftOut      []string // qualifiers that have been left out
	NumLeftOut   uint16   // number of properties that have been left out an needed to be recommended back
	NumTypes     uint16   // number of type properties
	NumObjTypes  uint16   // number of object type properties
	NumSubjTypes uint16   // number of object type properties
	Rank         uint16   // rank calculated for recommendation, equal to lec(recommendations)+1 if not fully recommendated back
	HitsAt1      uint8    // rank hits@1, value is 1 if leftOut is first in recommendations, 0 otherwise
	HitsAt5      uint8    // rank hits@5, value is 1 if leftOut is in the first 5 recommendations, 0 otherwise
	HitsAt10     uint8    // rank hits@10, value is 1 if leftOut is in the first 10 recommendations, 0 otherwise
	Duration     int64    // duration (in nanoseconds) of how long the recommendation took
}

// Setup variables for evaluation
var allRecs schematree.PropertyRecommendations
var typeRecs schematree.PropertyRecommendations
var VERBOSE bool

// Evaluate will run the evaluation process for the given model, workflow, testset, and handler.
// The results will be returned as a list of evaluation results, one for each transaction.
func EvaluateDataset(
	model *schematree.SchemaTree,
	workflow *strategy.Workflow,
	testset, handlerName string,
	verbose bool,
) []evalResult {

	var results []evalResult
	VERBOSE = verbose

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

	// read testset
	limitScan := 0 // 0 means no limit
	// summary := readTestSet(testset, limitScan)
	transIDs, qualifiersList, objTypesList, subjTypesList := readTestSet(testset, limitScan)

	// run for each transaction (line), take one out and evaluate
	for _, transID := range transIDs {
		qualifiers := qualifiersList[transID]
		objTypes := objTypesList[transID]
		subjTypes := subjTypesList[transID]

		// cache recommendations without any qualifier information
		t0 := time.Now()
		types := append(objTypes, subjTypes...)
		instanceTypes := schematree.NewInstanceFromInput([]string{}, types, model, true)
		recommendationsTypes := workflow.Recommend(instanceTypes)

		if verbose {
			log.Println("Type recommendations", recommendationsTypes)
			log.Println("Recommendations", len(recommendationsTypes), "for types found in", time.Since(t0))
			log.Println("___________________")
			log.Println("Transaction:", transID)
			log.Println("Types:", types)

			log.Println("###################")
			log.Println("Qualifiers:", qualifiers)
			log.Println(len(objTypes), "Object types:", objTypes)
			log.Println(len(subjTypes), "Subject types:", subjTypes)
		}

		// construct method that will be used to evaluate each transaction
		evaluator := func(reduced, leftout []string) evalResult {
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

		if verbose {
			log.Println("Number of transactions:", numTrans)
			log.Println("Evaluation took", time.Since(evalStart))
		}

	}

	return results
}

func finalRecommendations(fullRecs, typeRecs, allRecs schematree.PropertyRecommendations, limitRecs int) (int, []string) {
	existsRec := make(map[string]bool)
	outputRecs := make([]string, 0)
	for _, item := range fullRecs {
		if len(outputRecs) >= limitRecs && limitRecs != -1 {
			break
		}
		if item.Property.IsQualifier() {
			outputRecs = append(outputRecs, *item.Property.Str)
			existsRec[*item.Property.Str] = true // add to the map
		}
	}

	// concatenate typeRecs and outputRecs
	// only add the ones that are not already in the list
	// this is to avoid duplicates
	for _, item := range typeRecs {
		if len(outputRecs) >= limitRecs && limitRecs != -1 {
			break
		}
		if _, exists := existsRec[*item.Property.Str]; !exists && item.Property.IsQualifier() {
			outputRecs = append(outputRecs, *item.Property.Str)
			existsRec[*item.Property.Str] = true // add to the map
		}
	}

	// concatenate allRecs and outputRecs
	// only add the ones that are not already in the list
	for _, item := range allRecs {
		if len(outputRecs) >= limitRecs && limitRecs != -1 {
			break
		}
		if _, exists := existsRec[*item.Property.Str]; !exists && item.Property.IsQualifier() {
			outputRecs = append(outputRecs, *item.Property.Str)
			existsRec[*item.Property.Str] = true // add to the map
		}
	}

	return len(outputRecs), outputRecs
}

// evaluatePair will evaluate a pair made of a reduced set of qualifiers / properties and a left out qualifier
func evaluatePair(
	model *schematree.SchemaTree,
	workflow *strategy.Workflow,
	reducedSet, leftoutSet []string,
) evalResult {

	types := []string{} // the types are included in the reduced set (done by handler)
	instance := schematree.NewInstanceFromInput(reducedSet, types, model, true)
	fullRecs := workflow.Recommend(instance)

	if VERBOSE {
		log.Println("Full recommendations", fullRecs, "\n length", len(fullRecs))
	}

	// get the final recommendations
	_, outputRecs := finalRecommendations(fullRecs, typeRecs, allRecs, -1)

	// check if the recommendation contains the left out qualifiers
	rankLeftOut := 5843 // if not found, set to high value (for debugging)
	for _, lop := range leftoutSet {
		for idx, item := range outputRecs {
			if VERBOSE {
				log.Println("Checking", item, "against", lop)
			}
			if item == lop {
				// keep the smallest rank of the left out qualifiers
				// in order to have the best rank
				rank := idx + 1
				if rank < rankLeftOut {
					rankLeftOut = rank // set the rank
				}
				break
			}
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
		log.Println(len(outputRecs), "recommendations full info after adding types")
		log.Println(len(outputRecs), "recommendations full info after adding types and others")
		log.Println("Recommendations full info:", outputRecs)
		log.Println("Best rank of left out qualifiers", leftoutSet, "is", rankLeftOut)
		log.Println("Rank @1:", hitsAt1)
		log.Println("Rank @5:", hitsAt5)
		log.Println("Rank @10:", hitsAt10)
	}

	return evalResult{
		LeftOut:  leftoutSet,
		SetSize:  uint16(len(reducedSet) + len(types)),
		NumTypes: uint16(len(types)),
		Rank:     uint16(rankLeftOut),
		HitsAt1:  hitsAt1,
		HitsAt5:  hitsAt5,
		HitsAt10: hitsAt10,
	}
}

func categorizeItems(items []string) (qualifiers, objTypes, subjTypes []string) {

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
	return
}

// WriteResultsToFile will output the entire evalResult array to a json file
func WriteResultsToFile(filename string, results []evalResult) (outputfile string) {

	outputfile = filename + ".json"

	f, err := os.Create(outputfile)
	if err != nil {
		log.Fatalln("Could not create / open .json file")
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	file, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Panicln("Could not marshal evalResult to json")
	}

	_, err = w.Write(file)
	if err != nil {
		log.Panicln("Could not write evalResult to file")
	}

	return
}
