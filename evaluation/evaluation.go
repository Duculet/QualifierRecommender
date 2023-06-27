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

type evalResult struct {
	transID      uint32 // transaction ID (unique for each transaction)
	leftOut      string // property that has been left out
	setSize      uint16 // number of properties used to generate recommendations (both type and non-type)
	numTypes     uint16 // number of type properties
	numObjTypes  uint16 // number of object type properties
	numSubjTypes uint16 // number of object type properties
	rank         uint16 // rank calculated for recommendation, equal to lec(recommendations)+1 if not fully recommendated back
	rankAt1      uint8  // rank hits@1, value is 1 if leftOut is first in recommendations, 0 otherwise
	rankAt5      uint8  // rank hits@5, value is 1 if leftOut is in the first 5 recommendations, 0 otherwise
	rankAt10     uint8  // rank hits@10, value is 1 if leftOut is in the first 10 recommendations, 0 otherwise
	duration     int64  // duration (in nanoseconds) of how long the recommendation took
	// numLeftOut uint16 // number of properties that have been left out an needed to be recommended back
}

// Evaluate will run the evaluation process for the given workflow and testset.
// The results will be written to the given output file.
func EvaluateDataset(model *schematree.SchemaTree, workflow *strategy.Workflow, testset string, limitScan int, verbose bool) []evalResult {
	var results []evalResult

	// cache recommendations without any information
	// this is the baseline
	t0 := time.Now()
	instanceAll := schematree.NewInstanceFromInput([]string{}, []string{}, model, true)
	recommendationsAll := workflow.Recommend(instanceAll)

	if verbose {
		log.Println("Recommendations", len(recommendationsAll), "for all found in", time.Since(t0))
	}

	// read testset
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
			log.Println("___________________")
			log.Println("Transaction:", transID)
			log.Println("Types:", types)
			log.Println("Recommendations", len(recommendationsTypes), "for types found in", time.Since(t0))

			log.Println("###################")
			log.Println("Qualifiers:", qualifiers)
			log.Println("Object types:", objTypes)
			log.Println("Subject types:", subjTypes)
			log.Println(len(objTypes), "Object types:", objTypes)
			log.Println(len(subjTypes), "Subject types:", subjTypes)
		}

		// run for each transaction (line), take one out and evaluate
		for idx, leftOut := range qualifiers {
			// create a reduced set of qualifiers
			reducedSet := append([]string{}, qualifiers...)
			reducedSet = append(reducedSet[:idx], reducedSet[idx+1:]...)

			if verbose {
				log.Println("")
				log.Println("Initial set of qualifiers:", qualifiers)
				log.Println("Length of initial set of qualifiers:", len(qualifiers))
				log.Println("Reduced set of qualifiers:", reducedSet)
				log.Println("Evaluating with left out qualifier", leftOut)
				log.Println("-------------------")
			}

			// evaluate the reduced set
			evalTrans := evaluatePair(
				model, workflow,
				transID, leftOut,
				reducedSet, objTypes, subjTypes,
				recommendationsAll, recommendationsTypes,
				-1, verbose,
			)

			// add the evaluation result to the list
			results = append(results, evalTrans)
		}
	}
	return results
}

// evaluatePair will evaluate a pair made of a reduced set of qualifiers / properties and a left out qualifier
func evaluatePair(
	model *schematree.SchemaTree,
	workflow *strategy.Workflow,
	transID uint32, leftOut string,
	reducedSet, objTypes, subjTypes []string,
	allRecs, typeRecs schematree.PropertyRecommendations,
	limitRecs int, verbose bool,
) evalResult {

	types := append(objTypes, subjTypes...)
	instance := schematree.NewInstanceFromInput(reducedSet, types, model, true)

	evalStart := time.Now()
	recommendation := workflow.Recommend(instance)
	if verbose {
		log.Println("Recommendations", len(recommendation))
		log.Println("Recommendation took", time.Since(evalStart))
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

	// check if the recommendation contains the leftOut qualifier
	rankLeftOut := len(outputRecs) + 1 // if not found, set to the last rank
	for rank, item := range outputRecs {
		if verbose {
			log.Println("Checking", item, "against", leftOut)
		}
		if item == leftOut {
			rankLeftOut = rank
			break
		}
	}

	var rankAt1, rankAt5, rankAt10 uint8
	// set the rank hits @1, @5, @10
	if rankLeftOut == 1 {
		rankAt1 = 1
	}
	if rankLeftOut <= 5 {
		rankAt5 = 1
	}
	if rankLeftOut <= 10 {
		rankAt10 = 1
	}

	evalDuration := time.Since(evalStart)

	if verbose {
		log.Println(len(outputRecs), "recommendations full info")
		log.Println(len(outputRecs), "recommendations full info after adding types")
		log.Println(len(outputRecs), "recommendations full info after adding types and others")
		log.Println("Recommendations full info:", outputRecs)
		log.Println("Rank of left out qualifier", leftOut, "is", rankLeftOut)
		log.Println("Rank @1:", rankAt1)
		log.Println("Rank @5:", rankAt5)
		log.Println("Rank @10:", rankAt10)
		log.Println("Evaluation took", evalDuration)
	}

	return evalResult{
		transID:      transID,
		leftOut:      leftOut,
		setSize:      uint16(len(reducedSet) + len(types)),
		numTypes:     uint16(len(types)),
		numObjTypes:  uint16(len(objTypes)),
		numSubjTypes: uint16(len(subjTypes)),
		rank:         uint16(rankLeftOut),
		rankAt1:      rankAt1,
		rankAt5:      rankAt5,
		rankAt10:     rankAt10,
		duration:     evalDuration.Nanoseconds(),
	}
}

func categorizeItems(items []string) ([]string, []string, []string) {
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
	return qualifiers, objTypes, subjTypes
}

// readTestSet will read the test set file
// each line is a transaction
// each transaction is a list of qualifiers separated by a space
// returns lists of transaction IDs, qualifiers, object types and subject types
// (IDEA) func readTestSet(filename string, limitScan int) map[uint8][][]string {
// it will return a map of transaction IDs to three lists of qualifiers, object types and subject types
func readTestSet(filename string, limitScan int) ([]uint32, [][]string, [][]string, [][]string) {
	// load the testset
	testsetFile, err := os.Open(filename)
	if err != nil {
		log.Panicln(err)
	}
	defer testsetFile.Close()

	// var transactions map[uint8][][]string
	var tranIDs []uint32
	var qualifiersList, objTypesList, subjTypesList [][]string

	tsvScanner := bufio.NewScanner(testsetFile)
	for transID := 0; tsvScanner.Scan() && (transID < limitScan || limitScan == 0); transID++ {
		line := tsvScanner.Text()
		items := strings.Split(line, "\t")

		// categorize the items
		qualifiers, objTypes, subjTypes := categorizeItems(items)

		// (IDEA) append the items to the transactions map
		// transactions[uint8(transID)] = [][]string{qualifiers, objTypes, subjTypes}

		// append the items to the lists
		tranIDs = append(tranIDs, uint32(transID))
		qualifiersList = append(qualifiersList, qualifiers)
		objTypesList = append(objTypesList, objTypes)
		subjTypesList = append(subjTypesList, subjTypes)
	}

	return tranIDs, qualifiersList, objTypesList, subjTypesList
}

// WriteResultsToFile will output the entire evalResult array to a json file
func WriteResultsToFile(filename string, results []evalResult) string {
	f, err := os.Create(filename + ".json")
	if err != nil {
		log.Fatalln("Could not create / open .json file")
	}
	defer f.Close()
	// write the array of evalResults to the file
	enc := json.NewEncoder(f)
	err = enc.Encode(results)
	if err != nil {
		log.Fatalln("Could not encode results to json")
	}

	return filename + ".json"
}
