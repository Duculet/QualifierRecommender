#!/bin/bash
#SBATCH -o ./job_outputs/EV%j.output
#SBATCH -e ./job_outputs/EV%j.error
#SBATCH --job-name=Experiment_Eval
#SBATCH --time=12:00:00
#SBATCH -N 1
#SBATCH --ntasks-per-node=1

echo "Evaluation has started!"

# if argument is not provided, set default value to 100
# else set limit to argument value
limit_t=${1:-100}

echo "FF Begins"

models_dir=$HOME/experiments/FF/pbfiles/train
datasets_dir=$HOME/experiments/FF/tsvfiles/test
output_dir=$HOME/experiments/FF/evaluation

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
    $HOME/QualRecommender evaluate -m $model -d $dataset -o $output_dir -t=true
    echo "Saved in $output_dir"
done

echo "FF Done"

echo "#######"

echo "FT Begins"

models_dir=$HOME/experiments/FT/pbfiles/train
datasets_dir=$HOME/experiments/FT/tsvfiles/test
output_dir=$HOME/experiments/FT/evaluation

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
    $HOME/QualRecommender evaluate -m $model -d $dataset -o $output_dir -t=true
    echo "Saved in $output_dir"
done

echo "FT Done"

echo "#######"

echo "TF Begins"

models_dir=$HOME/experiments/TF/pbfiles/train
datasets_dir=$HOME/experiments/TF/tsvfiles/test
output_dir=$HOME/experiments/TF/evaluation

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
    $HOME/QualRecommender evaluate -m $model -d $dataset -o $output_dir -t=true
    echo "Saved in $output_dir"
done

echo "TF Done"

echo "#######"

echo "TT Begins"

models_dir=$HOME/experiments/TT/pbfiles/train
datasets_dir=$HOME/experiments/TT/tsvfiles/test
output_dir=$HOME/experiments/TT/evaluation

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
    $HOME/QualRecommender evaluate -m $model -d $dataset -o $output_dir -t=true
    echo "Saved in $output_dir"
done

echo "TT Done"

echo "Evaluation has finished!"