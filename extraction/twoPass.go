package extraction

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/grailbio/base/tsv"
	"gitlab.com/tozd/go/errors"
	"gitlab.com/tozd/go/mediawiki"
)

// typeMap is a map from an entity id to a list of types
var typeMap = make(map[string][]string)

func FirstPass(dumpPath string) errors.E {
	startTime := time.Now()
	counter := 0
	errE := mediawiki.ProcessWikidataDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{
			Path:                   dumpPath,
			ItemsProcessingThreads: 1,
		},
		func(_ context.Context, a mediawiki.Entity) errors.E {
			counter++
			if counter%100000 == 0 {
				log.Println(counter)
				log.Println("Runtime: ", time.Since(startTime).String())
			}
			claim_types := a.Claims["P31"]
			for _, subject := range claim_types {
				if subject.MainSnak.SnakType == mediawiki.Value {
					if subject.MainSnak.DataValue == nil {
						log.Fatal("Found a main snak with type Value, while it does not have a value. This is an error in the dump.")
					}
					value := subject.MainSnak.DataValue.Value
					switch v := value.(type) {
					default:
						log.Printf("Unexpected type %T for subject %s", value, a.ID)
					case mediawiki.WikiBaseEntityIDValue:
						typeMap[a.ID] = append(typeMap[a.ID], v.ID)
					}
				} else {
					log.Printf("Found a type statement without a value: %v", subject)
				}
			}
			return nil
		},
	)
	return errE
}

func SecondPass(dumpPath, outputDir string, obj, subj bool) errors.E {
	startTime := time.Now()
	counter := 0
	errE := mediawiki.ProcessWikidataDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{
			Path:                   dumpPath,
			ItemsProcessingThreads: 1,
		},
		func(_ context.Context, a mediawiki.Entity) errors.E {
			counter++
			if counter%100000 == 0 {
				log.Println(counter)
				log.Println("Runtime: ", time.Since(startTime).String())
			}
			for s_ID, statement := range a.Claims {
				has_qualifiers := false
				for _, value := range statement {
					// skip if value has no qualifiers
					if value.QualifiersOrder == nil {
						continue
					} else {
						has_qualifiers = true
					}
				}
				// skip if statement has no qualifiers
				if !has_qualifiers {
					continue
				}

				outputDir := filepath.Clean(outputDir)
				path := filepath.Clean(filepath.Join(outputDir, fmt.Sprintf("/%s.tsv", s_ID)))

				err := os.MkdirAll(filepath.Dir(path), 0750)
				if err != nil {
					log.Fatalf("Failed creating directory: %v", err)
				}

				// tsvFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0660)
				// if err != nil {
				// 	log.Fatalf("Failed opening file: %v", err)
				// }

				// if file exists, append to it
				// otherwise create it
				tsvFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
				if err != nil {
					log.Fatalf("Failed opening file: %v", err)
				}
				defer tsvFile.Close()

				tsvWriter := tsv.NewWriter(tsvFile)
				defer tsvWriter.Flush()

				for _, value := range statement {
					// skip if value has no qualifiers
					if value.QualifiersOrder == nil {
						continue
					}

					// save / export all data for this value to a line
					line := make([]string, 0)

					// collecting qualifiers
					line = append(line, value.QualifiersOrder...)

					// Include object types only if obj parameter is true
					if obj {
						// collect object types
						objTypes := make([]string, 0)
						switch value.MainSnak.SnakType {
						case mediawiki.Value:
							val := value.MainSnak.DataValue.Value
							switch v := val.(type) {
							default:
								objTypes = append(objTypes, fmt.Sprintf("%T", v))
							case mediawiki.WikiBaseEntityIDValue:
								v_id := v.ID
								newObjtypes, exists := typeMap[v_id]
								if exists {
									objTypes = append(objTypes, newObjtypes...)
								} else {
									log.Printf("Found an object of which the type could not be found, %v", v.ID)
									objTypes = append(objTypes, "NOTFOUND")
								}
							case mediawiki.MonolingualTextValue, mediawiki.StringValue:
								objTypes = append(objTypes, "string")
							case mediawiki.QuantityValue:
								objTypes = append(objTypes, "quantity")
							case mediawiki.TimeValue:
								objTypes = append(objTypes, "time")
							case mediawiki.GlobeCoordinateValue:
								objTypes = append(objTypes, "coordinate")
							}
						case mediawiki.NoValue:
							objTypes = append(objTypes, "NoValue")
						case mediawiki.SomeValue:
							objTypes = append(objTypes, "SomeValue")
						default:
							log.Panicf("This type does not exist. %v\n", value.MainSnak.SnakType)
						}

						for _, objType := range objTypes {
							// extend name by adding "o/" for object
							typeName := fmt.Sprintf("o/%s", objType)
							line = append(line, typeName)
						}
					}

					// Include subject types only if subj parameter is true
					if subj {
						subjTypes := typeMap[a.ID]

						for _, subjType := range subjTypes {
							// extend name by adding "s/" for subject
							typeName := fmt.Sprintf("s/%s", subjType)
							line = append(line, typeName)
						}
					}

					// // adding line to output for this property
					for _, val := range line {
						tsvWriter.WriteString(val)
					}
					tsvWriter.EndLine()
				}

			}
			return nil
		},
	)
	return errE
}
