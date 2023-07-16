#!/bin/bash

# setup variables
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