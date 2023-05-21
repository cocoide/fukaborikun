package utils_test

import (
	"reflect"
	"testing"

	"github.com/cocoide/fukaborikun/utils"
)

func Test_ExtractTextLines(t *testing.T) {
	text := "ewaea\newae\newa"
	expected := []string{"ewaea", "ewae", "ewa"}

	extracted := utils.ExtractTextLines(text)

	if !reflect.DeepEqual(extracted, expected) {
		t.Errorf("Unexpected result. Got: %v, Want: %v", extracted, expected)
	}
}
