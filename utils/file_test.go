package utils

import (
	"fmt"
	"testing"
)

func TestAddIndexToFilename(t *testing.T) {
	fmt.Println(AddIndexToFilename("Yes.Prime.Minister.S01E01.The.Grand.Design.mkv", 1))
}
