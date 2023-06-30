package evaluation

type handlerFunc func(transactionSummary, func(reduced, leftout []string) evalResult) []evalResult

func baseline(
	s transactionSummary,
	evaluator func(reduced, leftout []string) evalResult,
) []evalResult {

	results := make([]evalResult, 0, 1)

	// get recommendations without any information
	// also take one out version
	for _, leftOut := range s.qualifiers {
		newResult := evaluator([]string{}, []string{leftOut})
		results = append(results, newResult)
	}

	return results
}

func takeOneButType(
	s transactionSummary,
	evaluator func([]string, []string) evalResult,
) []evalResult {

	results := make([]evalResult, 0, 1)
	types := append(s.objTypes, s.subjTypes...)

	// get recommendations only with type information

	// take one qualifier out and evaluate with the rest, including the types
	for idx, left := range s.qualifiers {
		reducedSet := append([]string{}, s.qualifiers...)
		reducedSet = append(reducedSet[:idx], reducedSet[idx+1:]...) // remove the left out
		reducedSet = append(reducedSet, types...)                    // add the types

		leftOut := []string{left} // only one left out

		newResult := evaluator(reducedSet, leftOut)
		// fill in remaining information
		newResult.TransID = s.transID
		newResult.NumObjTypes = uint16(len(s.objTypes))
		newResult.NumSubjTypes = uint16(len(s.subjTypes))

		results = append(results, newResult)
	}

	return results
}

func takeAllButType(
	s transactionSummary,
	evaluator func([]string, []string) evalResult,
) []evalResult {

	results := make([]evalResult, 0, 1)
	types := append(s.objTypes, s.subjTypes...)

	// get recommendations only with type information

	// take one qualifier out and evaluate with the rest, including the types
	for idx, left := range s.qualifiers {
		reducedSet := append([]string{}, s.qualifiers...)
		reducedSet = append(reducedSet[:idx], reducedSet[idx+1:]...) // remove the left out
		reducedSet = append(reducedSet, types...)                    // add the types

		leftOut := []string{left} // only one left out

		newResult := evaluator(reducedSet, leftOut)
		newResult.TransID = uint32(s.transID)

		results = append(results, newResult)
	}

	return results
}
