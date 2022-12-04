package aorm

import (
	"fmt"
	"gopkg.in/guregu/null.v4"
	"testing"
)

type Person struct {
	PersonName null.String
}

func TestAll(t *testing.T) {
	fmt.Println("TestAll")
}
