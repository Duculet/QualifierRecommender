#!/bin/bash
#SBATCH -o ./job_outputs/EXTR%j.output
#SBATCH -e ./job_outputs/EXTR%j.error
#SBATCH --job-name=Experiment_Exract
#SBATCH --time=48:00:00
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=32
#SBATCH --partition=thin

# setup variables
$dump_path=$1
$output_path=$2
# FF if no arguments are given
$obj={3:-"false"}
$subj={4:-"false"}

$HOME/preprocessQ extract -f $1 -d $2 -o=$3 -s=$4