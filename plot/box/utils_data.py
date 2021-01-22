import os
import math
import numpy as np
import pandas as pd

"""
input: 
    models_data: {
        modelname: {
            run number 0: {parameter0: worst auc over all ensemble paths, ...}
            run number 1: {parameter0: worst auc over all ensemble paths, ...}
            ...
        }
    }
return:
    ranked: {
        modelname: {
            run number 0: ((param_1st: auc), (param_2nd: auc), ...)
            run number 1: ((param_1st: auc), (param_2nd: auc), ...)
            ...
        }
    }
"""
def ranking_allruns(models_data):
    ranked = {}
    for model in models_data.keys():
        ranked[model] = {}
        for rk in models_data[model].keys():
            ranked[model][rk] = sorting(models_data[model][rk])
    return ranked
def sorting(pvs):
    return (sorted(pvs.items(), key=lambda item:item[1]))[::-1]
    # st = [(k, v) for k, v in sorted(pvs.items(), key=lambda item: item[1])]
    # return st

"""
input:
    data: {
            run number 0: ((param_1st: auc), (param_2nd: auc), ...)
            run number 1: ((param_1st: auc), (param_2nd: auc), ...)
            ...
    }
"""
def percentile(ranked, low, high, mode="pessimistic"):
    filtered = []
    for rk in ranked.keys():
        l = int(len(ranked[rk]) * low)
        h = int(len(ranked[rk]) * high)
        filtered += [[rk, kv[0], kv[1]] for kv in ranked[rk][l: h]] # run number, parameter, performance

    if mode=="pessimistic":
        worst_per_run = []
        for rk in ranked.keys():
            temp = {}
            for item in filtered:
                if item[0] == rk:
                    temp[item[1]] = item[2]
            sorted = sorting(temp)
            worst_per_run.append([rk, sorted[-1][0], sorted[-1][1]])
        return worst_per_run
    elif mode=="optimistic":
        best_per_run = []
        for rk in ranked.keys():
            temp = {}
            for item in filtered:
                if item[0] == rk:
                    temp[item[1]] = item[2]
            sorted = sorting(temp)
            best_per_run.append([rk, sorted[0][0], sorted[0][1]])
        return best_per_run
    else:
        return filtered

"""
input:
    path: ensemble_paths of one ensemble model [path_ens1, path_ens2, ...]
return 
    data: {
            run number 0: {parameter0: [auc_ens1, auc_ens2, ...]}
            run number 1: {parameter0: [auc_ens1, auc_ens2, ...]}
          }
"""
def load_rewards(paths):
    data = {}
    for path in paths: # each ensemble seed
        params = os.listdir(path)
        for param in params: # each param
            pp = os.path.join(path, param)
            p_key = param#int(param.split("_")[1])

            temp = os.listdir(pp)
            runs = []
            for t in temp:
                if "rewards" in t:
                    runs.append(t)

            all_runs = {}
            for run in runs:
                r_per_step = pd.read_csv(os.path.join(pp, run))['rewards']
                # all_runs[int(run.split("-")[1].split(".")[0])] = np.mean(np.array(r_per_step)) # {run number: auc / total step}
                all_runs["run"+run.split("-")[1].split(".")[0]] = np.mean(np.array(r_per_step)) # {run number: auc / total step}

            for rk in all_runs.keys():
                if rk not in data.keys():
                    data[rk] = {p_key: []}
                if p_key not in data[rk].keys():
                    data[rk][p_key] = []
                data[rk][p_key].append(all_runs[rk])
    return data

"""
input:
    path: ensemble_paths of models
        {
            modelname: [path_ens1, path_ens2, ...]
        }
return:
    data: 
    { modelname: 
        {
            run number 0: {parameter0: worst auc over all ensemble paths, ...}
            run number 1: {parameter0: worst auc over all ensemble paths, ...}
        }
    }
"""
def loading_pessimistic(models_paths):
    models_data = {}
    for model in models_paths.keys():
        paths = models_paths[model]
        data = load_rewards(paths)
        for rk in data.keys():
            for pk in data[rk].keys():
                # print(rk, pk, data[rk][pk], np.array(data[rk][pk]).min())
                data[rk][pk] = np.array(data[rk][pk]).min()
        models_data[model] = data
    return models_data