package helloworld_plugin

import "fmt"

func beforeNewMethod([]interface{}) {
	fmt.Printf("Before main Method")
}

func afterNewMethod([]interface{}) {
	fmt.Printf("After main Method")
}
