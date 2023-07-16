#!/bin/bash

# setup variables
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