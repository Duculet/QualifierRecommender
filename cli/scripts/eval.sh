#!/bin/bash
#SBATCH -o ./job_outputs/E%j.output
#SBATCH -e ./job_outputs/E%j.error
#SBATCH --job-name=Experiment_Eval
#SBATCH --time=08:00:00
#SBATCH -N 1
#SBATCH --ntasks-per-node=1

echo "Evaluation has started!"

echo "TT Begins"
models_dir=$HOME/experiments/TT/pbfiles/train
datasets_dir=$HOME/experiments/TT/tsvfiles/test
output_dir=$HOME/experiments/TT/evaluation
for dataset in "$datasets_dir"/*
do
    model=$models_dir/$(basename $dataset).schemaTree.typed.pb
    echo "Running $model"
    echo "Against $dataset"
    $HOME/QrecEvalLisa evaluate -m $model -d $dataset -o $output_dir -t=true
    echo "Saved in $output_dir"
done
echo "TT Done"