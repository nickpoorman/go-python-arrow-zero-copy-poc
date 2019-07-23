# go-python-arrow-zero-copy-poc

This is a simple POC to demonstrate the ability to allocate data in Python and materialize it Go all within the same process without copying any of the data between the language boundaries using Apache Arrow.

**For a more complete example that includes record batches and tables see: [nickpoorman/go-py-arrow-bridge](https://github.com/nickpoorman/go-py-arrow-bridge)**

## Benchmark Results

As you can see below, the amount of time to move data across the Python/Go language boundary stays constant as the amount of data is increased.

```
(bullseye) ➜  go-python-arrow-zero-copy-poc git:(master) ✗ make bench
PKG_CONFIG_PATH=/Users/nick/anaconda3/envs/bullseye/lib/pkgconfig LD_LIBRARY_PATH=/Users/nick/anaconda3/envs/bullseye/lib/python3.7:/Users/nick/anaconda3/envs/bullseye/lib PYTHONPATH=/Users/nick/anaconda3/envs/bullseye/lib/python3.7/site-packages:/Users/nick/projects/go-python-arrow-zero-copy-poc/__python__ go test  -bench=. -run=- ./...
goos: darwin
goarch: amd64
pkg: github.com/nickpoorman/go-python-arrow-zero-copy-poc
BenchmarkAll/BenchmarkZeroCopy_0-4                  5000            263060 ns/op
BenchmarkAll/BenchmarkZeroCopy_500-4                5000            258294 ns/op
BenchmarkAll/BenchmarkZeroCopy_1000-4               5000            261115 ns/op
BenchmarkAll/BenchmarkZeroCopy_1500-4               5000            259570 ns/op
BenchmarkAll/BenchmarkZeroCopy_2000-4               5000            258090 ns/op
BenchmarkAll/BenchmarkZeroCopy_2500-4               5000            258926 ns/op
BenchmarkAll/BenchmarkZeroCopy_3000-4               5000            259880 ns/op
BenchmarkAll/BenchmarkZeroCopy_3500-4               5000            263706 ns/op
BenchmarkAll/BenchmarkZeroCopy_4000-4               5000            256914 ns/op
BenchmarkAll/BenchmarkZeroCopy_4500-4               5000            258202 ns/op
BenchmarkAll/BenchmarkZeroCopy_5000-4               5000            260762 ns/op
BenchmarkAll/BenchmarkZeroCopy_5500-4               5000            260207 ns/op
BenchmarkAll/BenchmarkZeroCopy_6000-4               5000            257175 ns/op
BenchmarkAll/BenchmarkZeroCopy_6500-4               5000            259549 ns/op
BenchmarkAll/BenchmarkZeroCopy_7000-4               5000            260398 ns/op
BenchmarkAll/BenchmarkZeroCopy_7500-4               5000            257537 ns/op
BenchmarkAll/BenchmarkZeroCopy_8000-4               5000            265975 ns/op
BenchmarkAll/BenchmarkZeroCopy_8500-4               5000            271194 ns/op
BenchmarkAll/BenchmarkZeroCopy_9000-4               5000            271857 ns/op
BenchmarkAll/BenchmarkZeroCopy_9500-4               5000            280712 ns/op
BenchmarkAll/BenchmarkZeroCopy_10000-4              5000            262157 ns/op
BenchmarkAll/BenchmarkZeroCopy_500000-4             5000            258362 ns/op
BenchmarkAll/BenchmarkZeroCopy_1500000-4            5000            258089 ns/op
BenchmarkAll/BenchmarkZeroCopy_2500000-4            5000            260858 ns/op
BenchmarkAll/BenchmarkZeroCopy_3500000-4            5000            307309 ns/op
BenchmarkAll/BenchmarkZeroCopy_4500000-4            5000            257813 ns/op
BenchmarkAll/BenchmarkZeroCopy_5500000-4            5000            261621 ns/op
PASS
ok      github.com/nickpoorman/go-python-arrow-zero-copy-poc    76.070s
```

## Run Results

```
(bullseye) ➜  go-python-arrow-zero-copy-poc git:(master) ✗ make run
rm -rf ./__python__/*.pyc
echo "Building bin/poc"
Building bin/poc
PKG_CONFIG_PATH=/Users/nick/anaconda3/envs/bullseye/lib/pkgconfig LD_LIBRARY_PATH=/Users/nick/anaconda3/envs/bullseye/lib/python3.7:/Users/nick/anaconda3/envs/bullseye/lib PYTHONPATH=/Users/nick/anaconda3/envs/bullseye/lib/python3.7/site-packages:/Users/nick/projects/go-python-arrow-zero-copy-poc/__python__ go build -o bin/poc poc.go
PKG_CONFIG_PATH=/Users/nick/anaconda3/envs/bullseye/lib/pkgconfig LD_LIBRARY_PATH=/Users/nick/anaconda3/envs/bullseye/lib/python3.7:/Users/nick/anaconda3/envs/bullseye/lib PYTHONPATH=/Users/nick/anaconda3/envs/bullseye/lib/python3.7/site-packages:/Users/nick/projects/go-python-arrow-zero-copy-poc/__python__ ./bin/poc
zero_copy values from Python:

[1237.9646270918913 1544.229225295952 1369.9551665480792 1603.9200385961944 1625.720304108054 1065.528859239813 1013.1679915548741 1837.46908209646 1259.3540143280077 1234.3309610466963 1995.6448355104628 1470.263507522448 1836.4614512743888 1476.353208699335 1639.068140544162 1150.616424023524 1634.8606582851885 1868.0453071432967 1523.1812103833013 1741.2518562014902]
```

## Details

This POC uses [pytasks](https://github.com/nickpoorman/pytasks) under the hood to handle embedding Python in the Go process and some rudimentary Python GIL locking.

It also uses some [additions to python3](https://github.com/nickpoorman/go-python3/commit/bfc9d2df89d46ad6b6e8ea80d0b15f2721031dba) module to handle raw pointers and slices between Python and Go.
