import os
import sys
cwd = os.getcwd()
sys.path.insert(0, cwd+'/../..')
from plot.box.utils_plot import *
from plot.box.paths_acrobot_shift import *

def learning_curve():
    paths = {
        "calibration hyperparameter transfer": acshift_online_shift_15k,
        # "shift-online (50k)": acshift_online_shift_50k,
        "calibration policy transfer": acshift_pitrans_calibration_fixed,
        # "calibration policy trans (learning 15k)": acshift_pitrans_calibration_learning15k,
        # "calibration policy trans (learning 50k)": acshift_pitrans_calibration_learning50k,
        # "shift-trans": acshift_pitrans_shift,
        # "default-online": acshift_online_default,
        # "default-trans(timeout1000)": acshift_pitrans_timeout1000_default,
        "true policy transfer": acshift_pitrans_default,
    }
    # plot_learning_perform(paths, "totals", "../img/acshift_total_zoomin", res_scale=-1, ylim=[80, 1000])
    # plot_learning_perform(paths, "totals", "../img/acshift_total", res_scale=-1)
    plot_learning_perform(paths, "totals", "../img/nonstationary_acrobot", res_scale=-1)

def sweep():
    paths = {
        "knn(laplace) orginal env": ["../../data/hyperparam_v5/acrobot/offline_learning/knn/learning/k3_laplace/timeout500/esarsa/step5k_env/data_optimal/drop0/sweep_rep1/"],
    }
    plot_param_sweep(paths, "totals", "../img/acshift_sweep", res_scale=-1)

learning_curve()
# sweep()