package code_buddy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Original function
func instrumentationMethodDefer(i int, b bool) (int, error) {
	fmt.Printf("instrumentationMethod: %d, %v \n", i, b)
	defer func() {
		fmt.Printf("defer function\n")
	}()
	return i + 1, nil
}

// After Instrumentation function
func instrumentationMethodDeferAfter(i int, b bool) (ret0 int, ret2 error) {
	// AUTO-GENERATED
	beforeInstrumentationDeferMethod([]interface{}{&i, &b})
	defer func() { afterInstrumentationDeferMethod([]interface{}{&ret0, &ret2}) }()

	fmt.Printf("instrumentationMethod: %d, %v \n", i, b)
	defer func() {
		fmt.Printf("defer function\n")
	}()
	return i + 1, nil
}

// before instrumentation method
func beforeInstrumentationDeferMethod(args []interface{}) {
	var i = args[0].(*int)
	*i = 100
}

// after instrumentation method
func afterInstrumentationDeferMethod(ret []interface{}) {
	var i = ret[0].(*int)
	*i = 100000
}

func TestInstrumentationWithDeferMethod(t *testing.T) {
	r, _ := instrumentationMethodDefer(10, true)
	assert.Equal(t, 11, r)

	r, _ = instrumentationMethodDeferAfter(10, true)
	assert.Equal(t, 100000, r)
}
