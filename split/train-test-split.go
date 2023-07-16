package split

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/grailbio/base/tsv"
)

// SplitAll takes a directory of tsv files and splits them into train and test files
func SplitAll(inputDir, outputDir string, percentageTest int, randomSeed int64) {
	items, err := os.ReadDir(inputDir)
	if err != nil {
		log.Panicln("Failed to read input directory", err)
	}
	for _, item := range items {
		if !item.IsDir() && filepath.Ext(item.Name()) == ".tsv" {
			filePath := filepath.Join(inputDir, item.Name())
			SplitTrainTest(filePath, outputDir, percentageTest, randomSeed)
		}
	}
}

// Split takes a tsv file and splits it into a train and test file
func SplitTrainTest(filePath, outputDir string, percentageTest int, randomSeed int64) {

	// Set random seed and create random generator
	randGen := rand.New(rand.NewSource(int64(randomSeed)))

	// PropId is the property id of the file being split
	propId := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	inFile, err := os.Open(filePath)
	if err != nil {
		log.Panicln("Failed to open input file", err)
	}
	defer inFile.Close()

	trainingDir := filepath.Join(outputDir, "train")
	err = os.MkdirAll(trainingDir, os.ModePerm)
	if err != nil {
		log.Panicln("Failed to create training directory", err)
	}
	trainFile, err := os.Create(filepath.Join(trainingDir, fmt.Sprintf("%s.tsv", propId)))
	if err != nil {
		log.Panicln("Failed to create train file", err)
	}
	defer trainFile.Close()

	testDir := filepath.Join(outputDir, "test")
	err = os.MkdirAll(testDir, os.ModePerm)
	if err != nil {
		log.Panicln("Failed to create test directory", err)
	}
	testFile, err := os.Create(filepath.Join(testDir, fmt.Sprintf("%s.tsv", propId)))
	if err != nil {
		log.Panicln("Failed to create test file", err)
	}
	defer testFile.Close()

	tsvTrain := tsv.NewWriter(trainFile)
	tsvTest := tsv.NewWriter(testFile)
	defer tsvTrain.Flush()
	defer tsvTest.Flush()

	tsvScanner := bufio.NewScanner(inFile)
	for tsvScanner.Scan() {
		if randGen.Intn(100) < percentageTest {
			tsvTest.WriteBytes(tsvScanner.Bytes())
			tsvTest.EndLine()
		} else {
			tsvTrain.WriteBytes(tsvScanner.Bytes())
			tsvTrain.EndLine()
		}
	}
}
