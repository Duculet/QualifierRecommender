# make a program that can evaluate the performance of the model
# load the evaluation results from a json file
# the json file should be in the following format:
# the json file should be a list of dictionaries, each dictionary is the evaluation result of a transaction
# INPUT:
#[
# {
#     "TransID": 0,
#     "LeftOut": "P582",
#     "SetSize": 8,
#     "NumTypes": 8,
#     "NumObjTypes": 1,
#     "NumSubjTypes": 7,
#     "Rank": 6,
#     "HitsAt1": 0,
#     "HitsAt5": 0,
#     "HitsAt10": 1,
#     "Duration": 722271
# },
# {
#     "TransID": 1,
#     "LeftOut": "P642",
#     "SetSize": 2,
#     "NumTypes": 2,
#     "NumObjTypes": 1,
#     "NumSubjTypes": 1,
#     "Rank": 1,
#     "HitsAt1": 1,
#     "HitsAt5": 1,
#     "HitsAt10": 1,
#     "Duration": 750203
#   },
#   ...
# ]


import json
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import os
import sys
import argparse

parser = argparse.ArgumentParser(description='Evaluate the performance of the model')
parser.add_argument('--input', type=str, default='testdata/splits/TT/eval/P9979.eval.takeOneButType.json', help='path to the evaluation results')
# parser.add_argument('--output', type=str, default='evaluation/eval_results.png', help='path to the output image')


if __name__ == '__main__':
    args = parser.parse_args()
    input_path = args.input
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






