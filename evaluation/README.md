# Recommender Evaluation suite

Note that all commands are run from the root of the project. These were tested on Snellius, the SURFsara supercomputer. That is the reason for the sbatch commands. You can run the commands without sbatch if you want to run them locally. Also, note that the commands contain subcommands that run other shell scripts. These can be found in hte cli package / directory of the project.

## Run Single Test: 

**Runs the standard recommender**
```bash
./RecommenderServer evaluate -m <modelfile> -d <testset> [-o <outputdir>] [-k <handler>] [-c <compress>] [-v <verbose>]
```

Note that you need to replace the names for the schematree model and the test set. Other parameters are optional. The handler is the name of the handler that you want to use as evaluation method. The default is the take one but type handler. The default output directory is the current directory. The default compression is gzip (value=true). The default verbosity is 0.

## Example of a data pipeline script

This is an example of how a complete data preparation pipeline could run. It also includes a 80:20 split of the dataset which is usually omitted for production usage.

### Data extraction

Data extraction is done in parallel for all experiments. The script will create a directory for each experiment and run the extraction in that directory. The script will also create a directory for the 80:20 split and run the split in that directory.

```bash
#!/bin/bash

# setup variables (change these to your liking)
experiments=("FF" "FT" "TF" "TT")
dump_path={1:-$HOME/wikidata-20230327-all.json.bz2}

echo "Running extractions for all experiments"

for experiment in "${experiments[@]}"
do
    output_path=$HOME/experiments/new/$experiment
    if [ ! -d "$output_path" ]
    then
        mkdir -p $output_path
    fi

    if [ $experiment == "FF" ]
    then
        obj=false
        subj=false
    elif [ $experiment == "FT" ]
    then
        obj=false
        subj=true
    elif [ $experiment == "TF" ]
    then
        obj=true
        subj=false
    elif [ $experiment == "TT" ]
    then
        obj=true
        subj=true
    fi
    
    echo "Running evaluation for $experiment"
    sbatch extract.sh -f $dump_path -d $output_path -o=$obj -s=$subj

done

echo "All extractions started!"
```

### Data preprocessing

This is usually omitted in production usage. It is only useful for testing purposes. It will split the dataset into a 80:20 split. The script will create a directory for each experiment and run the split in that directory.

```bash
#!/bin/bash

# setup variables (change these to your liking)
experiments=("FF" "FT" "TF" "TT")
input_dir={1:-$HOME/experiments/}
output_dir={2:-$HOME/experiments/}

echo "Running train-test splitter for all experiments"

for experiment in "${experiments[@]}"
do
    input_path=$input_dir/$experiment
    output_path=$output_dir/$experiment
    if [ ! -d "$output_path" ]
    then
        mkdir -p $output_path
    fi

    echo "Running train-test splitter for $experiment"
    sbatch split.sh -i $input_path -o $output_path
done

echo "All train-test splits started!"
```

### Data evaluation

What follows is an example of an evaluation script that is useful to generate statistics in multiple views in one go. It will run evaluations with multiple combinations of models and handlers for various experiments. After the evaluations are performed it will package the results for easier downloading.

```bash
#!/bin/bash

experiments=("FF" "FT" "TF" "TT")
handlers=("baseline" "takeOneButType" "takeAllButType")

# if argument is not provided, set default value to 0
# else set limit to argument value
limit_t=${1:-0}

echo "Running evaluation for all experiments with limit $limit_t for transactions"

for experiment in "${experiments[@]}"
do
    echo "Running evaluation for $experiment"
    for handler in "${handlers[@]}"
    do
        echo "Running evaluation with $handler"
        sbatch eval.sh $experiment $handler $limit_t
    done
done

echo "All experiments started!"
```