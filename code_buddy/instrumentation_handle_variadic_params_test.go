package code_buddy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Original function
func instrumentationMethodVariadicParam(i ...string) (string, error) {
	fmt.Printf("instrumentationMethod: %s \n", i[0])
	return i[0], nil
}

// After Instrumentation function
func instrumentationMethodVariadicParamAfter(i ...string) (ret0 string, ret1 error) {
	// AUTO-GENERATED
	beforeInstrumentationVariadicParamMethod([]interface{}{&i})
	defer func() { afterInstrumentationVariadicParamMethod([]interface{}{&ret0}) }()

	fmt.Printf("instrumentationMethod: %s \n", i[0])
	return i[0], nil
}

// before instrumentation method
func beforeInstrumentationVariadicParamMethod(args []interface{}) {
	var i = args[0].(*[]string)
	(*i)[0] = "Hello"
}

// after instrumentation method
func afterInstrumentationVariadicParamMethod(ret []interface{}) {
	var i = ret[0].(*string)
	*i = "World"
}

func TestInstrumentationWithVariadicParamMethod(t *testing.T) {
	r, _ := instrumentationMethodVariadicParam("asb", "cde", "fgs")
	assert.Equal(t, "asb", r)

	r, _ = instrumentationMethodVariadicParamAfter("asb", "cde", "fgs")
	assert.Equal(t, "World", r)
}
