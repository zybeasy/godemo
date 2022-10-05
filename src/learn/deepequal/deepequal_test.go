package deepequal

import (
	// "bytes"
	"fmt"
	"testing"
)

func TestEqual(t *testing.T) {
	one, oneAgain, two := 1, 1, 2
	type CyclePtr *CyclePtr
	var cyclePtr1, cyclePtr2 CyclePtr
	cyclePtr1 = &cyclePtr1
	cyclePtr2 = &cyclePtr2

	fmt.Println(one, oneAgain, two, cyclePtr1, cyclePtr2)
}
