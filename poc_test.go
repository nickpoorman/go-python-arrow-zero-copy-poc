package main

import (
	"fmt"
	"testing"

	"github.com/nickpoorman/pytasks"
)

// This benchmark generates random data in Python and then materializes it
// in Go without copying the underlying pyarrow buffer. To verify this,
// the time it takes to run each benchmark should not increase as the amount
// of random data to materialize increases.
func BenchmarkAll(b *testing.B) {
	for i := 0; i <= 10000; i += 500 {
		b.Run(fmt.Sprintf("BenchmarkZeroCopy_%d", i), zeroPyCopyBenchmarkN(i))
	}
	for i := 500000; i <= 5500000; i += 1000000 {
		b.Run(fmt.Sprintf("BenchmarkZeroCopy_%d", i), zeroPyCopyBenchmarkN(i))
	}

	// At this point we know we won't need Python anymore in this
	// program, we can restore the state and lock the GIL to perform
	// the final operations before exiting.
	err := pytasks.GetPythonSingleton().Finalize()
	if err != nil {
		b.Fatal(err)
	}
}

// So the benchmarks don't get compiled out during optimization.
var zeroCopyMethodRes []byte

func zeroPyCopyBenchmarkN(numRows int) func(b *testing.B) {
	return func(b *testing.B) {
		py := pytasks.GetPythonSingleton()

		fooModule, err := py.ImportModule("foo")
		if err != nil {
			b.Fatal(err)
		}
		defer fooModule.DecRef()

		randomSeries, err := CreateRandomSeries(py, fooModule, numRows)
		if err != nil {
			b.Fatal(err)
		}

		// CreateRandomSeries takes some setup time that should not
		// be included in the benchmark so reset the timer.
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := ZeroCopyTask(py, fooModule, randomSeries)
			zeroCopyMethodRes = result.Buffer
		}
	}
}
