# make a program that can evaluate the performance of the model
# load the evaluation results from a json file
# the json file should be in the following format:
# the json file should be a list of dictionaries, each dictionary is the evaluation result of a transaction
# INPUT:
#


import json
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import os
import sys
import argparse


if __name__ == '__main__':
    input_path = "evaluation/eval_results/eval_results_1.json"
    # output_path = args.output

    # load the evaluation results from the json file
    # use the json file name as the title of the plot
    # use numpy to calculate the average performance of the model

    # load the evaluation results from the json file
    with open(input_path, 'r') as f:
        eval_results = pd.read_json(f)

        # display the evaluation results
        print(eval_results)

    # use the json file name as the title of the plot
    title = os.path.basename(input_path)

    # use numpy to calculate the average performance of the model
    # the average performance is calculated by averaging the performance of each transaction
    # the performance of each transaction is calculated by the number of hits at 1, 5, 10
    # the performance of each transaction is calculated by the rank of the left out triple
    # transactions = np.load(eval_results)

    # calculate the average performance of the model
    # avg_rank = np.mean(transactions[])

    # plot the average performance of the model
    # the x axis is the number of types in the transaction
    # the y axis is the average performance of the model






