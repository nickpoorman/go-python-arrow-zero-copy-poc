# foo.py
import random
import pandas as pd
import pyarrow as pa

random.seed(3)


def create_random_series(num=50):
    return pd.Series([random.uniform(1000, 2000) for x in range(num)])


def zero_copy(data):
    arrdata = pa.array(data, type=pa.float64())
    return arrdata
