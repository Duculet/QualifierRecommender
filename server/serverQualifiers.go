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

// QRecommenderRequest is the data representation of the request json.
type QRecommenderRequest struct {
	Property   string   `json:"property"`
	Qualifiers []string `json:"qualifiers"`
	SubjTypes  []string `json:"subjTypes"`
	ObjTypes   []string `json:"objTypes"`
}

// QRecommenderResponse is the data representation of the response json.
type QRecommenderResponse struct {
	Recommendations []QRecommendationOutputEntry `json:"recommendations"`
}

// QRecommendationOutputEntry is each entry that is returned from the server.
type QRecommendationOutputEntry struct {
	QualifierStr *string `json:"qualifier"`
	Probability  float64 `json:"probability"`
}

// formatForLoggingQ formats the input for logging by removing newlines and carriage returns.
func formatForLoggingQ(input QRecommenderRequest) string {
	var jsonstring = fmt.Sprintln(input)
	escapedjsonstring := strings.Replace(jsonstring, "\n", "", -1)
	escapedjsonstring = strings.Replace(escapedjsonstring, "\r", "", -1)
	return escapedjsonstring
}

// Load all models into a map with the model id as key.
var models = make(map[string]*schematree.SchemaTree, 0)
var workflows = make(map[string]*strategy.Workflow, 0)

func LoadAllModels(models_dir, workflowFile string) {
	items, err := os.ReadDir(models_dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range items {
		if !item.IsDir() {
			if strings.HasSuffix(item.Name(), ".tsv.schemaTree.typed.pb") {
				id := strings.TrimSuffix(item.Name(), ".tsv.schemaTree.typed.pb")
				model_path := filepath.Clean(filepath.Join(models_dir, item.Name()))
				models[id] = GetModel(model_path)
				workflows[id] = GetWorkflow("", models[id])
			}
		}
	}

	log.Println("Models loaded:", len(models))
}

// GetWorkflow returns the workflow to be used for the recommendation.
func GetWorkflow(workflowFile string, model *schematree.SchemaTree) *strategy.Workflow {
	var workflow *strategy.Workflow
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

// GetModel returns the model to be used for the recommendation.
func GetModel(model_path string) *schematree.SchemaTree {
	cleanedmodelBinary := filepath.Clean(model_path)

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

// getTypes combines the subject and object types into a single list.
// The types are prefixed with "s/" or "o/" to indicate whether they are subject or object types.
// This is the format expected by the schematree.
func getTypes(subjTypes, objTypes []string) []string {
	types := make([]string, 0)
	for _, subjType := range subjTypes {
		types = append(types, fmt.Sprintf("s/%s", subjType))
	}
	for _, objType := range objTypes {
		types = append(types, fmt.Sprintf("o/%s", objType))
	}
	return types
}

// setupQualifierRecommender sets up the handler for the qualifier recommender.
// It loads all models from the given directory and returns a handler function.
// The handler function expects a json input of type QRecommenderRequest and returns a json output of type QRecommenderResponse.
func setupQualifierRecommender(models_dir, workflowFile string, hardLimit int) func(http.ResponseWriter, *http.Request) {
	if models_dir == "" {
		log.Panicln("No path for the models specified")
	}
	if hardLimit < 1 && hardLimit != -1 {
		log.Panic("hardLimit must be positive, or -1")
	}
	LoadAllModels(models_dir, workflowFile)
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

		// Select the model and workflow based on the input.
		model := models[input.Property]
		workflow := workflows[input.Property]

		// Combine the subject and object types into a single list.
		types := getTypes(input.SubjTypes, input.ObjTypes)

		// Create an instance from the input.
		instance := schematree.NewInstanceFromInput(input.Qualifiers, types, model, true)

		// Start the timer for the recommendation.
		t1 := time.Now()
		// Make a recommendation based on the chosen strategy and the assessed input.
		recommendation := workflow.Recommend(instance)
		// Print the request and the time it took to answer it.
		log.Println("request ", escapedjsonstring, " answered in ", time.Since(t1))

		// Put a hard limit on the recommendations returned
		outputRecs := limitRecommendationsQ(recommendation, hardLimit)

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
		if recommendation.Property.IsQualifier() {
			outputRecs = append(outputRecs, QRecommendationOutputEntry{
				QualifierStr: recommendation.Property.Str,
				Probability:  recommendation.Probability,
			})
		}
	}
	return outputRecs
}

// SetupEndpoints configures a router with all necessary endpoints and their corresponding handlers.
func SetupNewEndpoints(models_dir, workflowFile string, hardLimit int) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/Qrecommender", setupQualifierRecommender(models_dir, workflowFile, hardLimit))
	return router
}
