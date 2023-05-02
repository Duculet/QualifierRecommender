package server

import (
	"RecommenderServer/configuration"
	"RecommenderServer/schematree"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"RecommenderServer/strategy"
)

type RecommenderRequest struct {
	Types      []string `json:"types"`
	Properties []string `json:"properties"`
}

// RecommenderResponse is the data representation of the json.
type RecommenderResponse struct {
	Recommendations []RecommendationOutputEntry `json:"recommendations"`
}

// RecommendationOutputEntry is each entry that is return from the server.
type RecommendationOutputEntry struct {
	PropertyStr *string `json:"property"`
	Probability float64 `json:"probability"`
}

// setupRecommender will setup a handler to recommend properties based on the list of properties and types.
// That handler will return an array of recommendations with their respective probabilities.
// This array will contain at most `hardlimit` resutls. To not have a limit, set to -1.
func setupLeanRecommender(
	model *schematree.SchemaTree,
	workflow *strategy.Workflow,
	hardLimit int,
) func(http.ResponseWriter, *http.Request) {
	if model == nil {
		log.Panicln("Nil model specified")
	}
	if workflow == nil {
		log.Panicln("Nil workflow specified")
	}
	if hardLimit < 1 && hardLimit != -1 {
		log.Panic("hardLimit must be positive, or -1")
	}

	return func(res http.ResponseWriter, req *http.Request) {

		// Decode the JSON input and build a list of input strings
		var input = RecommenderRequest{}

		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			res.WriteHeader(400)
			log.Println("Malformed Request.") // TODO: Json-Schema helps
			return
		}
		escapedlogstring := formatForLogging(input)
		log.Println("request received ", escapedlogstring)

		instance := schematree.NewInstanceFromInput(input.Properties, input.Types, model, true)

		// Make a recommendation based on the assessed input and chosen strategy.
		t1 := time.Now()
		recomendations := workflow.Recommend(instance)
		log.Println("request ", escapedlogstring, " answered in ", time.Since(t1))

		// Put a hard limit on the recommendations returned
		outputRecs := limitRecommendations(recomendations, hardLimit)

		// Pack everything into the response
		recResp := RecommenderResponse{Recommendations: outputRecs}

		// Write the recommendations as a JSON array.
		res.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(res).Encode(recResp)
		if err != nil {
			log.Println("Malformed Response.", &recResp)
			return
		}
	}
}

func formatForLogging(input RecommenderRequest) string {
	var jsonstring = fmt.Sprintln(input)
	escapedjsonstring := strings.Replace(jsonstring, "\n", "", -1)
	escapedjsonstring = strings.Replace(escapedjsonstring, "\r", "", -1)
	return escapedjsonstring
}

// Limit the recommendations to contain at most `hardLimit` items and convert to output entries.
// If hardLimit is -1, then no limit is imposed.
func limitRecommendations(recommendations schematree.PropertyRecommendations, hardLimit int) []RecommendationOutputEntry {

	capacity := len(recommendations)
	if hardLimit != -1 {
		if capacity > hardLimit {
			capacity = hardLimit
		}
	}
	outputRecs := make([]RecommendationOutputEntry, 0, capacity)

	for _, recommendation := range recommendations {
		if hardLimit != -1 && len(outputRecs) >= hardLimit {
			break
		}
		if recommendation.Property.IsProp() {
			outputRecs = append(outputRecs, RecommendationOutputEntry{
				PropertyStr: recommendation.Property.Str,
				Probability: recommendation.Probability,
			})
		}
	}
	return outputRecs
}

type QRecommenderRequest struct {
	Property   string   `json:"property"`
	Qualifiers []string `json:"qualifiers"`
	SubjTypes  []string `json:"subjTypes"`
	ObjTypes   []string `json:"objTypes"`
}

// RecommenderResponse is the data representation of the json.
type QRecommenderResponse struct {
	Recommendations []QRecommendationOutputEntry `json:"recommendations"`
}

// RecommendationOutputEntry is each entry that is return from the server.
type QRecommendationOutputEntry struct {
	QualifierStr *string `json:"qualifier"`
	Probability  float64 `json:"probability"`
}

func formatForLoggingQ(input QRecommenderRequest) string {
	var jsonstring = fmt.Sprintln(input)
	escapedjsonstring := strings.Replace(jsonstring, "\n", "", -1)
	escapedjsonstring = strings.Replace(escapedjsonstring, "\r", "", -1)
	return escapedjsonstring
}

// typePrefix is list of prefixes for each type.
var typePrefix = []string{"o/", "s/"}

func IsQualifier(name string) bool {
	for _, prefix := range typePrefix {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}
	return true
}

// make map of all models
// model directory and iterate over files
var models = make(map[string]*schematree.SchemaTree, 0)
var workflow *strategy.Workflow

func LoadAllModels() {
	// load all models
	items, err := os.ReadDir("testdata")
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range items {
		if !item.IsDir() {
			if strings.HasSuffix(item.Name(), ".tsv.schemaTree.typed.pb") {
				id := strings.TrimSuffix(item.Name(), ".tsv.schemaTree.typed.pb")
				models[id] = GetModel(id)
			}
		}
	}

	fmt.Println("Models loaded: ", len(models))
}

func GetWorkflow(workflowFile string, model *schematree.SchemaTree) *strategy.Workflow {
	if workflowFile != "" {
		config, err := configuration.ReadConfigFile(&workflowFile)
		if err != nil {
			log.Panicln(err)
		}
		err = config.Test()
		if err != nil {
			log.Panicln(err)
		}
		workflow, err = configuration.ConfigToWorkflow(config, model)
		if err != nil {
			log.Panicln(err)
		}
		log.Printf("Run Config Workflow %v", workflowFile)
	} else {
		workflow = strategy.MakePresetWorkflow("best", model)
		log.Printf("Run best Recommender ")
	}
	return workflow
}

func GetModel(path string) *schematree.SchemaTree {
	modelBinary := fmt.Sprintf("/home/aducu/testdata/%s.tsv.schemaTree.typed.pb", path)
	// modelBinary := fmt.Sprintf("testdata/%s.tsv.schemaTree.typed.pb", path)

	cleanedmodelBinary := filepath.Clean(modelBinary)

	// Load the schematree from the binary file.

	log.Printf("Loading schema (from file %v): ", cleanedmodelBinary)

	/// file handling
	f, err := os.Open(cleanedmodelBinary)
	if err != nil {
		log.Printf("Encountered error while trying to open the file: %v\n", err)
		log.Panic(err)
	}

	model, err := schematree.LoadProtocolBufferFromReader(f)
	if err != nil {
		log.Panicln(err)
	}
	schematree.PrintMemUsage()

	return model
}

func setupQualifierRecommender(hardLimit int) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {

		// Decode the JSON input and build a list of input strings
		var input = QRecommenderRequest{}

		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			res.WriteHeader(400)
			log.Println(err)
			log.Println("Malformed Request.") // TODO: Json-Schema helps
			return
		}
		escapedjsonstring := formatForLoggingQ(input)
		log.Println("request received ", escapedjsonstring)

		// Select the model based on the input.
		model := models[input.Property]

		// Prepend subject and object types to the qualifiers
		transaction := make([]string, len(input.Qualifiers))
		copy(transaction, input.Qualifiers)

		for _, subjType := range input.SubjTypes {
			transaction = append(transaction, fmt.Sprintf("s/%s", subjType))
		}
		for _, objType := range input.ObjTypes {
			transaction = append(transaction, fmt.Sprintf("o/%s", objType))
		}

		instance := schematree.NewInstanceFromInput(transaction, make([]string, 0), model, true)

		// Make a recommendation based on the assessed input and chosen strategy.
		t1 := time.Now()

		// Map including workflows and models
		workflow := GetWorkflow("", model)
		rec := workflow.Recommend(instance)
		log.Println("request ", escapedjsonstring, " answered in ", time.Since(t1))

		// Put a hard limit on the recommendations returned
		outputRecs := limitRecommendationsQ(rec, hardLimit)

		// FILTER into 2 groups:
		// 1. Constrained
		// 2. Unconstrained
		for _, rec := range rec {
			if IsQualifier(*rec.Property.Str) {
				outputRecs = append(outputRecs, QRecommendationOutputEntry{
					QualifierStr: rec.Property.Str,
					Probability:  rec.Probability,
				})
			}
		}

		// Pack everything into the response
		recResp := QRecommenderResponse{Recommendations: outputRecs}

		// Write the recommendations as a JSON array.
		res.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(res).Encode(recResp)
		if err != nil {
			log.Println("Malformed Response.", &recResp)
			return
		}
	}
}

// Limit the recommendations to contain at most `hardLimit` items and convert to output entries.
// If hardLimit is -1, then no limit is imposed.
func limitRecommendationsQ(recommendations schematree.PropertyRecommendations, hardLimit int) []QRecommendationOutputEntry {

	capacity := len(recommendations)
	if hardLimit != -1 {
		if capacity > hardLimit {
			capacity = hardLimit
		}
	}
	outputRecs := make([]QRecommendationOutputEntry, 0, capacity)

	for _, recommendation := range recommendations {
		if hardLimit != -1 && len(outputRecs) >= hardLimit {
			break
		}
		if recommendation.Property.IsProp() {
			outputRecs = append(outputRecs, QRecommendationOutputEntry{
				QualifierStr: recommendation.Property.Str,
				Probability:  recommendation.Probability,
			})
		}
	}
	return outputRecs
}

// SetupEndpoints configures a router with all necessary endpoints and their corresponding handlers.
// func SetupEndpoints(model *schematree.SchemaTree, glossary *glossary.Glossary, workflow *strategy.Workflow, hardLimit int) http.Handler {
func SetupEndpoints(model *schematree.SchemaTree, workflow *strategy.Workflow, hardLimit int) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/recommender", setupLeanRecommender(model, workflow, hardLimit))
	return router
}

func SetupNewEndpoints(hardLimit int) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/Qrecommender", setupQualifierRecommender(hardLimit))
	return router
}
