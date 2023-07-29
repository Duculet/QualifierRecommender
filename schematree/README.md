# Schematree Module

The Schematree Module contains the main datastructure and the standard recommendation algorithm.

SchemaTree.go, SchemaNode.go and DataTypes.go form the tree data structure. Recommendation.go performs the recommendations.

## Interface

## Properties

```go

Create(filename string, firstNsubjects uint64, typed bool, minSup uint32) // creates a new Schematree from a rdf file
Load(filePath string) // loads a schematree from a encoded file

Recommend(properties []string, types []string) // recommends a list of property candidates

```

## Qualifiers

```go

LoadProtocolBufferFromReader(input io.Reader) // loads a schematree from a protocol buffer file

NewInstanceFromInput(argProps []string, argTypes []string, argTree *SchemaTree, argUseCache bool) // creates a new instance of recommendations

Recommend(asm *Instance) // recommends a list of property candidates for the given workflow

```
