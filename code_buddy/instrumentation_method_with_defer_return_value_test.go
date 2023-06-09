package code_buddy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Original function
func instrumentationMethodDeferWithReturnValue(i int, b bool) (j int, k error) {
	fmt.Printf("instrumentationMethod: %d, %v \n", i, b)
	defer func() {
		fmt.Printf("defer function\n")
		j = 5
		k = nil
	}()
	return i + 1, nil
}

// After Instrumentation function
func instrumentationMethodDeferWithReturnValueAfter(i int, b bool) (j int, k error) {
	beforeInstrumentationDeferWithReturnValueMethod([]interface{}{&i, &b})
	defer func() { afterInstrumentationDeferWithReturnValueMethod([]interface{}{&j, &k}) }()

	fmt.Printf("instrumentationMethod: %d, %v \n", i, b)
	defer func() {
		fmt.Printf("defer function\n")
		j = 5
		k = nil
	}()
	return i + 1, nil
}

// before instrumentation method
func beforeInstrumentationDeferWithReturnValueMethod(args []interface{}) {
	var i = args[0].(*int)
	*i = 100
}

// after instrumentation method
func afterInstrumentationDeferWithReturnValueMethod(ret []interface{}) {
	var i = ret[0].(*int)
	*i = 100000
}

func TestInstrumentationWithDeferWithReturnValueMethod(t *testing.T) {
	r, _ := instrumentationMethodDeferWithReturnValue(10, true)
	assert.Equal(t, 5, r)

	r, _ = instrumentationMethodDeferWithReturnValueAfter(10, true)
	assert.Equal(t, 100000, r)
}
