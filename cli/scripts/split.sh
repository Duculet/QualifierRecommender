#!/bin/bash
#SBATCH -o ./job_outputs/SPLIT%j.output
#SBATCH -e ./job_outputs/SPLIT%j.error
#SBATCH --job-name=Experiment_Split
#SBATCH --time=24:00:00
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=32
#SBATCH --partition=thin

input=$1
output=$2
# FF if no arguments are given
experiment={3:-"FF"}

echo "Splitting train-test data for $experiment"
echo "Input: $input"
echo "Output: $output"

$HOME/preprocessQ split -i $input -o $output

echo "$experiment done!"