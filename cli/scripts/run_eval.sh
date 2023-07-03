#!/bin/bash

experiments=("FF" "FT" "TF" "TT")
handlers=("baseline" "takeOneButType" "takeAllButType")

echo "Running evaluation for all experiments"

for experiment in "${experiments[@]}"
do
    echo "Running evaluation for $experiment"
    for handler in "${handlers[@]}"
    do
        echo "Running evaluation with $handler"
        sbatch eval.sh $experiment $handler
    done
done

echo "All experiments started!"