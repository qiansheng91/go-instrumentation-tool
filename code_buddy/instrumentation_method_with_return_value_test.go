package code_buddy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Original function
func instrumentationMethod(i int, b bool) (int, error) {
	fmt.Printf("instrumentationMethod: %d, %v \n", i, b)
	return i + 1, nil
}

// After Instrumentation function
func instrumentationMethodAfter(i int, b bool) (ret0 int, ret2 error) {
	// AUTO-GENERATED
	beforeInstrumentationMethod([]interface{}{&i, &b})
	defer func() { afterInstrumentationMethod([]interface{}{&ret0, &ret2}) }()

	fmt.Printf("instrumentationMethod: %d, %v \n", i, b)
	return i + 1, nil
}

// before instrumentation method
func beforeInstrumentationMethod(args []interface{}) {
	var i = args[0].(*int)
	*i = 100
}

// after instrumentation method
func afterInstrumentationMethod(ret []interface{}) {
	var i = ret[0].(*int)
	*i = 100000
}

func TestInstrumentationWithReturnValue(t *testing.T) {
	r, _ := instrumentationMethod(10, true)
	assert.Equal(t, 11, r)

	r, _ = instrumentationMethodAfter(10, true)
	assert.Equal(t, 100000, r)
}
