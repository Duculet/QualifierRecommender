package evaluation

import (
	"RecommenderServer/schematree"
	"log"
	"strings"
)

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

func finalRecommendations(
	fullRecs, typeRecs, allRecs schematree.PropertyRecommendations,
	limitRecs int,
) (int, []string) {

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
