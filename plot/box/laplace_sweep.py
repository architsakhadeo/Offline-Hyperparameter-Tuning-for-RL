import os
import sys
import numpy
cwd = os.getcwd()
sys.path.insert(0, cwd+'/../..')
from plot.box.utils_data import *
from plot.box.paths_laplace import *

def env_param_eval(path, eval_key):
    res_dict = loading_info_all(path, eval_key=eval_key)
    print(dict(sorted(res_dict.items(), key=lambda item: item[1],reverse=True)))

# # print("\nAcrobot laplace test0")
# # env_param_eval(acrobot_laplace_test0, "Averaged improvement = ")
# print("\nAcrobot laplace test1")
# env_param_eval(acrobot_laplace_test1, "Dynamic awareness = ")
#
# # print("\nPuddlerand laplace test0")
# # env_param_eval(pdrand_laplace_test0, "Averaged improvement = ")
# print("\nPuddlerand laplace test1")
# env_param_eval(pdrand_laplace_test1, "Dynamic awareness = ")
#
# print("\nAcrobot network scaled")
# env_param_eval(acrobot_scale_network, "Averaged loss = ")

print("\nAcrobot network scaled laplace")
env_param_eval(acrobot_scale_laplace_network, "Averaged loss = ")

print("\nPuddlerand network scaled")
env_param_eval(pdrand_scale_network, "Averaged loss = ")

print("\nPuddlerand network scaled laplace")
env_param_eval(pdrand_scale_laplace_network, "Averaged loss = ")


# print("\nCartpole Laplace Optimal\n")
# env_param_eval(cartpole_laplace_test1_optimal, "Dynamic awareness = ")
#
# print("\nCartpole Laplace Suboptimal\n")
# env_param_eval(cartpole_laplace_test1_suboptimal, "Dynamic awareness = ")
#
# print("\nCartpole Laplace Random\n")
# env_param_eval(cartpole_laplace_test1_random, "Dynamic awareness = ")
