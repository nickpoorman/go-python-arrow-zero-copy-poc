package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/DataDog/go-python3"
	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/nickpoorman/pytasks"
)

// Create some data in Python and materialize it using Go without copying
// the underlying pyarrow buffer.
func main() {
	var numElements = flag.Int("n", 20, "number of elements")
	flag.Parse()

	py := pytasks.GetPythonSingleton()
	defer func() {
		// At this point we know we won't need Python anymore in this
		// program, we can restore the state and lock the GIL to perform
		// the final operations before exiting.
		err := pytasks.GetPythonSingleton().Finalize()
		if err != nil {
			panic(err)
		}
	}()

	fooModule, err := py.ImportModule("foo")
	if err != nil {
		panic(err)
	}
	defer func() {
		err := pytasks.GetPythonSingleton().NewTaskSync(func() {
			fooModule.DecRef()
		})
		if err != nil {
			panic(err)
		}
	}()

	randomSeries, err := CreateRandomSeries(py, fooModule, *numElements)
	if err != nil {
		panic(err)
	}

	zeroCopyMethodRes := ZeroCopyTask(py, fooModule, randomSeries)

	data := ArrowDataFromBuffer(zeroCopyMethodRes.Buffer, zeroCopyMethodRes.Offset, zeroCopyMethodRes.NullCount)
	defer data.Release()

	iface := ArrowInterfaceFromData(data)
	defer iface.Release()

	floatArray := iface.(*array.Float64)
	values := floatArray.Float64Values()
	fmt.Printf("zero_copy values from Python:\n\n%v\n", values)
}

func ArrowInterfaceFromData(data *array.Data) array.Interface {
	// array.NewFloat64Data(data)
	// MakeFromData constructs a strongly-typed array instance from generic Data.
	return array.MakeFromData(data)
}

func ArrowDataFromBuffer(buffer []byte, offset, nullCount int) *array.Data {
	buff := memory.NewBufferBytes(buffer)
	buffLen := buff.Len() / arrow.Float64SizeBytes
	data := array.NewData(
		arrow.PrimitiveTypes.Float64, buffLen,
		[]*memory.Buffer{nil, buff},
		nil, nullCount, offset,
	)
	return data
}

// CreateRandomSeries generates a random pandas series using the method implemented in python.
func CreateRandomSeries(py pytasks.PythonSingleton, module *python3.PyObject, numElements int) (*python3.PyObject, error) {
	var randomSeries *python3.PyObject
	var randomSeriesErr error

	wg1, err := py.NewTask(func() {
		createRandomSeriesFn := module.GetAttrString("create_random_series")
		defer createRandomSeriesFn.DecRef()

		pyNumElements := python3.PyLong_FromGoInt(numElements)
		randomSeries = createRandomSeriesFn.CallFunctionObjArgs(pyNumElements)
		if randomSeries == nil {
			randomSeriesErr = errors.New("randomSeries is nil")
		}
	})
	if err != nil {
		return nil, err
	}
	wg1.Wait()

	return randomSeries, randomSeriesErr
}

type ZeroCopyResult struct {
	Offset    int
	NullCount int
	Buffer    []byte
}

// CallZeroCopyPyFunc calls the zero_copy method implemented in python and
// extracts the underlying pyarrow buffer, offset, and nullcount from the result.
func CallZeroCopyPyFunc(module *python3.PyObject, args ...*python3.PyObject) *ZeroCopyResult {
	zeroCopyFunc := module.GetAttrString("zero_copy")
	defer zeroCopyFunc.DecRef()

	arrdata := zeroCopyFunc.CallFunctionObjArgs(args...)
	if arrdata == nil {
		panic("arrdata is nil")
	}

	pyNullCount := arrdata.GetAttrString("null_count")
	if pyNullCount == nil {
		panic("pyNullCount is nil")
	}
	defer pyNullCount.DecRef()
	nullCount := python3.PyLong_AsLong(pyNullCount)

	pyOffset := arrdata.GetAttrString("offset")
	if pyOffset == nil {
		panic("pyOffset is nil")
	}
	defer pyOffset.DecRef()
	offset := python3.PyLong_AsLong(pyOffset)

	buffersFunc := arrdata.GetAttrString("buffers")
	if buffersFunc == nil {
		panic("buffersFunc is nil")
	}
	defer buffersFunc.DecRef()

	buffers := buffersFunc.CallFunctionObjArgs()
	if buffers == nil {
		panic("buffers is nil")
	}

	// First buffer in the list is the null mask buffer, second is the values.
	pyByteArray := python3.PyList_GetItem(buffers, 1)
	if pyByteArray == nil {
		panic("pyByteArray is nil")
	}
	defer pyByteArray.DecRef()

	pybuf, err := python3.PyObject_GetBuffer(pyByteArray, python3.PyBUF_SIMPLE)
	if err {
		panic("pybuf PyObject_GetBuffer err")
	}

	goBytes := python3.PyObject_GetBufferBytes(pybuf)
	if goBytes == nil {
		panic("goBytes is nill")
	}

	return &ZeroCopyResult{
		Offset:    offset,
		NullCount: nullCount,
		Buffer:    goBytes,
	}
}

func ZeroCopyTask(py pytasks.PythonSingleton, module, randomSeries *python3.PyObject) *ZeroCopyResult {
	var result *ZeroCopyResult

	err := py.NewTaskSync(func() {
		result = CallZeroCopyPyFunc(module, randomSeries)
	})
	if err != nil {
		panic(err)
	}

	return result
}
