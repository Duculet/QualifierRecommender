# Snellius (bash) scripts for running the evaluation

## Data extraction

### run_extractions.sh
This script will run the extraction for all experiments. It will create a directory for each experiment and run the extraction in that directory.

### extract.sh
This script will run the extraction for a single experiment. It will create a directory for each experiment and run the extraction in that directory. It was designed to be run by the run_extract.sh script. This allowed parallel execution of the extractions on the Snellius cluster.

## Data preprocessing

### run_splits.sh
This script will run the split for all experiments. It will create a directory for each experiment and run the split in that directory.

### split.sh
This script will run the split for a single experiment. It will create a directory for each experiment and run the split in that directory. It was designed to be run by the run_split.sh script. This allowed parallel execution of the splits on the Snellius cluster.

### run_build_trees.sh (not finished)
This script will run the tree building for all experiments. It will create a directory for each experiment and run the tree building in that directory.

### build_tree.sh (not finished)
This script will run the tree building for all experiments. It will create a directory for each experiment and run the tree building in that directory. It was designed to be run by the run_build_trees.sh script. This allowed parallel execution of the tree building on the Snellius cluster.

## Evaluation

# run_eval.sh
This script will run the evaluation for all experiments. It will create a directory for each experiment and run the evaluation in that directory.

### eval.sh
This script will run the evaluation for a single experiment. It will create a directory for each experiment and run the evaluation in that directory. It was designed to be run by the run_eval.sh script. This allowed parallel execution of the evaluation on the Snellius cluster.