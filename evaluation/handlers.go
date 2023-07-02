package evaluation

type handlerFunc func(transactionSummary, func(reducedSet []string, leftOut string) evalResult) []evalResult

// This is the baseline handler
// It will try to recommend the left out qualifier without any information
func baseline(
	s transactionSummary,
	evaluator func(reducedSet []string, leftOut string) evalResult,
) (results []evalResult) {

	// get recommendations without any information
	// also take one out version
	for _, leftOut := range s.qualifiers {
		newResult := evaluator([]string{}, leftOut)
		results = append(results, newResult)

		// fill in remaining information
		newResult.TransID = s.transID
	}

	return
}

// This handler will take one qualifier out
// It will evaluate with the rest, including the types
func takeOneButType(
	s transactionSummary,
	evaluator func(reducedSet []string, leftOut string) evalResult,
) (results []evalResult) {

	types := append(s.objTypes, s.subjTypes...)

	// take one qualifier out and evaluate with the rest, including the types
	for idx, leftOut := range s.qualifiers {
		reducedSet := append([]string{}, s.qualifiers...)
		reducedSet = append(reducedSet[:idx], reducedSet[idx+1:]...) // remove the left out
		reducedSet = append(reducedSet, types...)                    // add the types

		newResult := evaluator(reducedSet, leftOut)
		// fill in remaining information
		newResult.TransID = s.transID
		newResult.NumObjTypes = uint16(len(s.objTypes))
		newResult.NumSubjTypes = uint16(len(s.subjTypes))
		newResult.NumTypes = uint16(len(types))

		results = append(results, newResult)
	}

	return
}

// This handler will take out one qualifier
// It will evaluate only with the types
func takeAllButType(
	s transactionSummary,
	evaluator func(reducedSet []string, leftOut string) evalResult,
) (results []evalResult) {

	types := append(s.objTypes, s.subjTypes...)

	for _, leftOut := range s.qualifiers {
		reducedSet := append([]string{}, types...) // add the types

		newResult := evaluator(reducedSet, leftOut)
		// fill in remaining information
		newResult.TransID = s.transID
		newResult.NumObjTypes = uint16(len(s.objTypes))
		newResult.NumSubjTypes = uint16(len(s.subjTypes))
		newResult.NumTypes = uint16(len(types))

		results = append(results, newResult)
	}

	return results
}
