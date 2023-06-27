#!/bin/bash
#SBATCH -o ./job_outputs/E%j.output
#SBATCH -e ./job_outputs/E%j.error
#SBATCH --job-name=Experiment_Eval
#SBATCH --time=12:00:00
#SBATCH -N 1
#SBATCH --ntasks-per-node=1

echo "Evaluation has started!"

echo "TT Begins"
models_dir=$HOME/experiments/TT/pbfiles/train
datasets_dir=$HOME/experiments/TT/tsvfiles/test
output_dir=$HOME/experiments/TT/evaluation
# if argument is not provided, set default value to 100
# else set limit to argument value
limit_t=${1:-100}
for dataset in "$datasets_dir"/*
do
    # if dataset is smaller than limit, skip it
    # else run evaluation
    if [ $(wc -l < $dataset) -lt $limit_t ]
    then
        echo "Skipping $dataset"
        continue
    fi
    model=$models_dir/$(basename $dataset).schemaTree.typed.pb
    echo "Running $model"
    echo "Against $dataset"
    $HOME/QrecEvalLisa evaluate -m $model -d $dataset -o $output_dir -t=true
    echo "Saved in $output_dir"
done
echo "TT Done"