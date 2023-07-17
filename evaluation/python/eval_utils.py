from itertools import islice
from tqdm import tqdm
import numpy as np
import pandas as pd
import json
import gzip
import glob
import csv

class ModelResult:
    """
    Stores the results of a model.
    Class variables:
        model_id: the id of the model
        eval_count: the number of evaluations
        trans_count: the number of transactions
        eval_time: the total time spent on evaluations
        qualifier_pop: the popularity of each qualifier in the model (counter)
        eval_results: the results of each evaluation
    Also provides methods to compute metrics from the results.
    """

    # constructor
    def __init__(
            self, 
            model_id: int, 
            trans_count: int,
            eval_count: int, 
            eval_time: int, 
            qualifier_pop: pd.Series, 
            eval_results: pd.DataFrame
            ):
        self.model_id = model_id
        self.trans_count = trans_count
        self.eval_count = eval_count
        self.eval_time = eval_time
        self.qualifier_pop = qualifier_pop
        self.eval_results = eval_results
        self.extra_setup()

    # string representation
    def __str__(self):
        return f"Model {self.model_id} with {self.eval_count} evaluations in {self.eval_time} ns"
    
    # representation
    def __repr__(self):
        return f"Model {self.model_id} with {self.eval_count} evaluations in {self.eval_time} ns"
    
    # get the number of evaluations
    def __len__(self):
        return len(self.eval_results)
    
    def extra_setup(self):
        # store the results with rank 5843 (missing recommendations)
        self.missing_results = self.eval_results[self.eval_results['Rank'] == 5843].copy()
        # only keep valid results for further processing
        self.eval_results = self.eval_results[self.eval_results['Rank'] != 5843].copy()
        # add another column for the number of non type items (qualifiers)
        # these represent other data that was used in for the recommendation
        self.eval_results['NumNonTypes'] = self.eval_results['SetSize'] - self.eval_results['NumTypes']
        # store some metrics for faster access
        self.missing_count = len(self.missing_results)
        self.missing_percent = self.missing_count / self.eval_count
        # to be computed and stored when needed
        self.unique_qualifiers = None  
        self.missing_qualifiers = None
        self.grouped_stats = dict()
        self.avg_rank = None
        self.avg_rank_qualifier = dict()
        self.hits_at_r = dict()
        self.hits_at_r_qualifier = dict(dict())
    
    # get unique qualifiers (either valid or missing = /w rank 5843)
    def get_qualifiers(self, missing = False):
        if missing:
            if self.missing_qualifiers is not None:
                return self.missing_qualifiers
            self.missing_qualifiers = self.missing_results['LeftOut'].unique()
            return self.missing_qualifiers
        else:
            if self.unique_qualifiers is not None:
                return self.unique_qualifiers
            self.unique_qualifiers = self.eval_results['LeftOut'].unique()
            return self.unique_qualifiers
        
    
    # get qualifiers ranked by popularity, ascending = True for least popular first
    def get_qualifiers_popranked(self, ascending = False, include_missing = False):
        if include_missing:
            return self.qualifier_pop.sort_values(ascending=ascending).index
        else:
            return self.qualifier_pop.drop(self.get_qualifiers(missing=True)).sort_values(ascending=ascending).index
    
    # get the popularity of a qualifier
    def get_qualifier_pop(self, qualifier: str):
        try:
            return self.qualifier_pop[qualifier]
        except:
            raise KeyError(f"Qualifier {qualifier} not found, must be one of {self.get_qualifiers()}")
    
    # get the results for a qualifier
    def get_qualifier_results(self, qualifier: str):
        try:
            return self.eval_results[self.eval_results['LeftOut'] == qualifier]
        except:
            raise KeyError(f"Qualifier {qualifier} not found, must be one of {self.get_qualifiers()}")
    
    # get the evaluation time per evaluation
    # unit: 'ns', 'ms', 's' (default: 'ns' - nanoseconds)
    # avg: True to get the average time per evaluation, False to get the total time
    def get_eval_time(self, unit = 'ns', avg = False):
        time = self.eval_time / self.eval_count if avg else 1
        match unit:
            case 'ns':
                return time
            case 'ms':
                return time / 1000000
            case 's':
                return time / 1000000000
            case _:
                raise ValueError(f"Invalid unit {unit}, must be 'ns', 'ms' or 's'")
            
    def get_statistics(self) -> pd.Series:
        """
        Get the statistics for the model.
        """

        # compute the statistics
        rank = self.get_avg_rank()
        hits_at_1 = self.get_hits_at_r(1)
        hits_at_5 = self.get_hits_at_r(5)
        hits_at_10 = self.get_hits_at_r(10)

        # create a series with the statistics
        stats = pd.Series({
            'Mean': rank,
            'Top1': hits_at_1,
            'Top5': hits_at_5,
            'Top10': hits_at_10
        })

        # round to 4 decimal places
        stats = stats.round(4)

        # return the statistics
        return stats
    
    # get a dataframe of statistics grouped by a column
    def get_grouped_statistics(self, groupby: str):
        # check if the groupby column is valid
        if groupby not in self.eval_results.columns:
            raise ValueError(f"Invalid groupby column {groupby}, must be one of {self.eval_results.columns}")
        # check if the statistics have already been computed
        if groupby in self.grouped_stats:
            return self.grouped_stats[groupby]
        # define columns to compute statistics on
        columns = ['Rank', 'HitsAt1', 'HitsAt5', 'HitsAt10']
        # define the statistics to compute for each column
        statistics = {'Rank': ['count', 'mean', 'median', 'std'], 'HitsAt1': 'mean', 'HitsAt5': 'mean', 'HitsAt10': 'mean'}
        # micro-average the statistics for each transaction
        grouped_stats = self.eval_results.groupby([groupby, "TransID"])[columns].mean().reset_index()
        # macro-average the statistics for each group value
        grouped_stats = grouped_stats.groupby(groupby)[columns].agg(statistics).fillna(0)
        # save stats to a dataframe
        stats = pd.DataFrame(grouped_stats.to_records())
        # rename columns
        stats.columns = [groupby, 'Count', 'Mean', 'Median', 'Stddev', 'Top1', 'Top5', 'Top10']
        # multiply the mean by 100 to get a percentage
        stats[['Top1', 'Top5', 'Top10']] *= 100
        # round values to 4 decimals
        stats = stats.round(4)
        # store the statistics
        self.grouped_stats[groupby] = stats
        # free up memory
        del grouped_stats
        # return the statistics
        return stats
    
    # compute the average rank of the model (for a given qualifier)
    # if parameter groupby is set, compute the average ranks for each group
    def get_avg_rank(self, qualifier: str = None):
        if qualifier:
            if qualifier in self.avg_rank_qualifier:
                return self.avg_rank_qualifier[qualifier]
            # get the average rank of the qualifier
            avg_rank = round(self.get_qualifier_results(qualifier)['Rank'].mean(), 4)
            # store the average rank of the qualifier
            self.avg_rank_qualifier[qualifier] = avg_rank
        else:
            # get the average ranks of each transaction (micro-average)
            avg_ranks = self.eval_results.groupby('TransID')['Rank'].mean()
            # get the global average rank (macro-average)
            avg_rank = round(avg_ranks.mean(), 4)
            # free up memory
            avg_ranks = None
            # store the average rank of the model
            self.avg_rank = avg_rank
        # return the average rank of the model
        return avg_rank
    
    # # explanatory only (for lambda function below)
    # def get_perc_hit(results):
    #     # get the number of hits at rank r
    #     hits = results[f'HitsAt{r}'].sum()
    #     # get the number of evaluations
    #     evals = results[f'HitsAt{r}'].count()
    #     # get the percentage of hits at rank r for each transaction (micro-average)
    #     perc_hits = hits / evals * 100
    #     # get the global percentage of hits at rank r for the model / qualifier (macro-average)
    #     perc_hit = round(perc_hits.mean())
    #     # # free up memory
    #     # hits = None
    #     # evals = None
    #     # perc_hits = None
    #     return perc_hit
    
    # compute the percentage of hits at rank r (for a given qualifier)
    # r can be 1, 5 or 10 (default: 1)
    # if parameter qualifier is set, compute the percentage of hits at rank r for the qualifier
    def get_hits_at_r(self, r: int = 1, qualifier: str = None):
        get_perc_hit = lambda res: round((res[f'HitsAt{r}'].sum() / res[f'HitsAt{r}'].count() * 100).mean(), 4)
        try:
            if qualifier:
                if r in self.hits_at_r_qualifier:
                    if qualifier in self.hits_at_r_qualifier[r]:
                        return self.hits_at_r_qualifier[r][qualifier]
                    # get the qualifier results
                    results = self.get_qualifier_results(qualifier)
                    # get the percentage of hits for these results
                    perc_hit = get_perc_hit(results)
                    # save the result
                    self.hits_at_r_qualifier[r][qualifier] = perc_hit
                else:
                    # get the qualifier results
                    results = self.get_qualifier_results(qualifier)
                    # get the percentage of hits for these results
                    perc_hit = get_perc_hit(results)
                    # save the result
                    self.hits_at_r_qualifier[r] = {qualifier: perc_hit}
            else:
                if r in self.hits_at_r:
                    return self.hits_at_r[r]
                # get the results grouped by transaction
                results = self.eval_results.groupby('TransID')
                # get the percentage of hits for these results
                perc_hit = get_perc_hit(results)
                # save the result
                self.hits_at_r[r] = perc_hit
            # free up memory
            results = None
            # return the percentage of hits at rank r
            return perc_hit
        except:
            raise ValueError(f"Invalid r {r}, must be 1, 5 or 10")

def generate_paths(
        base_path: str, 
        eval_methods: list[str], 
        experiments: list[str]
        ) -> dict[str, dict[str, str]]:
    """
    Generate a dictionary of paths for each eval_method and experiment.
    """
    paths = dict()

    for eval_method in eval_methods:
        paths[eval_method] = dict()
        for experiment in experiments:
            paths[eval_method][experiment] = f"{base_path}{eval_method}/{experiment}/"
    
    return paths

def generate_models(files: str, filelimit=0, mintrans=0, verbose=False) -> list[ModelResult]:
    """
    Generate a list of ModelResult objects from the given files.
    Only uses the first filelimit number of files, if filelimit is set.
    Only uses files with at least mintrans number of transactions, if mintrans is set.
    If verbose is set, a progress bar is displayed.
    """
    files = glob.glob(files + "*.json")
    files = sorted(files)
    if filelimit:
        files = islice(files, filelimit)

    if verbose:
        num_files = filelimit if filelimit else len(files)
        pbar = tqdm(total=num_files, desc="Processing files")

    models = []
    for file in files:
        with gzip.open(file) as f:
            data = json.load(f)
            eval_results = data['EvalResults']
            t_count = (eval_results[-1]['TransID'] + 1) if eval_results else 0

            if verbose:
                pbar.update(1)

            if t_count == 0 or t_count < mintrans:
                continue

            model = ModelResult(
                model_id = data['ModelID'],
                trans_count = t_count,
                eval_count = data['EvalCount'],
                eval_time = data['EvalTime'],
                qualifier_pop = pd.Series(data['QualifierPop']),
                eval_results = pd.DataFrame(data['EvalResults'])
            )

            models.append(model)

    # free up memory
    model = None
    eval_results = None
    data = None
    f = None

    if verbose:
        print(len(models), "models generated")
        pbar.close()
        del pbar
        
    return models

# compute various statistics for the given models
def get_stats(models: list[ModelResult], groupby: str = None) -> pd.DataFrame:
    """
    Compute various statistics for the given models.
    """
    combined_stats = []
    for model in models:
        stats = model.get_grouped_statistics(groupby)
        combined_stats.append(stats)
    combined_stats = pd.concat(combined_stats)

    # in this case, we want the experiment statistics
    # therefore, we only keep the stats that can be further described
    # we only keep the groupby column
    # and only the columns [Count, Mean, Top1, Top5, Top10]
    columns = [groupby, 'Count', 'Mean', 'Top1', 'Top5', 'Top10']
    combined_stats = combined_stats[columns].reset_index(drop=True)
    # we group by the groupby column
    grouped_stats = combined_stats.groupby(groupby)
    # we recompute statistics, store them in a dataframe, and change the column names
    statistics = {'Count': 'sum', 'Mean': 'mean', 'Top1': 'mean', 'Top5': 'mean', 'Top10': 'mean'}
    experiment_stats = pd.DataFrame(grouped_stats.agg(statistics).to_records())
    experiment_stats.columns = [groupby, 'Count', 'Mean', 'Top1', 'Top5', 'Top10']
    # round to 4 decimal places
    experiment_stats = experiment_stats.round(4)
    # return the experiment statistics
    return experiment_stats

def save_stats(out_path: str, name: str, stats: pd.DataFrame) -> None:
    out_file = out_path + name + '_stats.csv'
    stats.to_csv(out_file, index=False)

    print("Saved results to", out_file)
    
def load_stats(in_path: str, name: str) -> pd.DataFrame:
    in_file = in_path + name + '_stats.csv'
    stats = pd.read_csv(in_file)
    return stats

def get_model_stats(model: ModelResult) -> pd.Series:
    """
    Get the statistics for the given model.
    """
    # get the statistics for the model
    rank = model.get_avg_rank()
    hits_at_1 = model.get_hits_at_r(1)
    hits_at_5 = model.get_hits_at_r(5)
    hits_at_10 = model.get_hits_at_r(10)

    # create a series with the statistics
    stats = pd.Series({
        'Mean': rank,
        'Top1': hits_at_1,
        'Top5': hits_at_5,
        'Top10': hits_at_10
    })

    # round to 4 decimal places
    stats = stats.round(4)

    # return the statistics
    return stats

# compute various statistics for the given models
def get_models_simple_stats(models: list[ModelResult]) -> pd.Series:
    ranks = np.array([])
    hits_at_1 = np.array([])
    hits_at_5 = np.array([])
    hits_at_10 = np.array([])
    missing_percent = np.array([])
    duration = np.array([])

    for model in models:
        ranks = np.append(ranks, model.get_avg_rank())
        hits_at_1 = np.append(hits_at_1, model.get_hits_at_r(1))
        hits_at_5 = np.append(hits_at_5, model.get_hits_at_r(5))
        hits_at_10 = np.append(hits_at_10, model.get_hits_at_r(10))
        missing_percent = np.append(missing_percent, model.missing_percent)
        duration = np.append(duration, model.get_eval_time('ms', True))

    # calculate the average rank, median rank, and standard deviation of the ranks
    mean = np.mean(ranks).round(4)
    median = np.median(ranks).round(4)
    stddev = np.std(ranks).round(4)
    mean_hits_at_1 = np.mean(hits_at_1).round(4)
    mean_hits_at_5 = np.mean(hits_at_5).round(4)
    mean_hits_at_10 = np.mean(hits_at_10).round(4)
    mean_missing_percent = np.mean(missing_percent).round(4)
    mean_duration = np.mean(duration).round(4)

    # combine statistics into a pd.Series
    stats = pd.Series({
        'Mean': mean,
        'Median': median,
        'StdDev': stddev,
        'Top1': mean_hits_at_1,
        'Top5': mean_hits_at_5,
        'Top10': mean_hits_at_10,
        'Missing': mean_missing_percent,
        'Duration': mean_duration
    })

    # return the statistics
    return stats

# display the stats
def display_stats(eval_method, mean, median, stddev, mean_hits_at_1, mean_hits_at_5, mean_hits_at_10):
    print("Results for eval_method", eval_method)
    print("Rank:")
    print("\tMean: " + str(mean))
    print("\tMedian: " + str(median))
    print("\tStandard Deviation: " + str(stddev))
    print("Hits:")
    print("\t@1: " + str(mean_hits_at_1) + "%")
    print("\t@5: " + str(mean_hits_at_5) + "%")
    print("\t@10: " + str(mean_hits_at_10) + "%")

# save the stats to a csv file
def save_model_stats(out_path: str, name: str, *stats) -> None:
    out_file = out_path + name + '_stats.csv'

    with open(out_file, 'w') as csvfile:
        w = csv.writer(csvfile)
        w.writerow(['Mean', 'Median', 'Standard Deviation', 'Hits@1', 'Hits@5', 'Hits@10', 'Duration'])
        w.writerow(stats)

    print("Saved results to", out_file)

