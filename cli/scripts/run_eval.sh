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