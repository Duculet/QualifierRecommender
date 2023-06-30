package evaluation

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type transactionSummary struct {
	transID    uint32
	qualifiers []string
	subjTypes  []string
	objTypes   []string
}

// propertySummaryReader will read the test set file (each file represents a property)
// it will create a transactionSummary for each transaction
// and pass it to the handler
func propertySummaryReader(
	fileName string,
	handler func(p transactionSummary),
) (transactionCount uint64) {
	// load the testset
	transIDs, qualifiersList, objTypesList, subjTypesList := readTestSet(fileName, 0)

	// for each transaction, create a transactionSummary and pass it to the handler
	for idx, transID := range transIDs {
		// create a transactionSummary
		p := transactionSummary{
			transID:    transID,
			qualifiers: qualifiersList[idx],
			subjTypes:  subjTypesList[idx],
			objTypes:   objTypesList[idx],
		}

		// pass it to the handler
		handler(p)
	}
	return uint64(len(transIDs))
}

// readTestSet will read the test set file
// each line is a transaction
// each transaction is a list of qualifiers separated by a space
// returns lists of transaction IDs, qualifiers, object types and subject types
// (IDEA) func readTestSet(fileName string, limitScan int) map[uint8][][]string {
// it will return a map of transaction IDs to three lists of qualifiers, object types and subject types
func readTestSet(fileName string, limitScan int) ([]uint32, [][]string, [][]string, [][]string) {
	// load the testset
	testsetFile, err := os.Open(fileName)
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
