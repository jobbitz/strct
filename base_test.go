// Copyright 2019 Job Stoit. All rights reserved.

package strct

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

type testObj struct {
	Empty              int
	Str                string    `default:"test"`
	shouldNotRead      string    `default:"SHOULDNOTREAD"`
	ShouldNotOverWrite string    `default:"override"`
	Bool               bool      `default:"true"`
	Flt                float64   `default:"4.5"`
	Int                int       `default:"4"`
	Sli                []int     `default:"1;2;3"`
	File               io.Reader `default:"./base.go"`
}

func TestScanAndParse(t *testing.T) {
	obj := new(testObj)
	obj.ShouldNotOverWrite = `another test`

	err := Scan(obj, func(field reflect.StructField, value *reflect.Value) error {
		tagVal := field.Tag.Get(`default`)
		if tagVal == `` {
			return nil
		}

		if tagVal == `SHOULDNOTREAD` {
			t.Errorf("Shouldn't be able to read shouldNotRead")
			return nil
		}

		return Parse(tagVal, value)
	})
	if err != nil {
		t.Errorf("parse error: %v\n", err)
	}

	eq(`test`, obj.Str, t)
	eq(`another test`, obj.ShouldNotOverWrite, t)
	eq(true, obj.Bool, t)
	eq(4.5, obj.Flt, t)
	eq(4, obj.Int, t)
	eq([]int{1, 2, 3}, obj.Sli, t)

	if obj.File == nil {
		t.Error(`file not parsed`)
	}
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(obj.File); err != nil {
		t.Error(err)
	}

	if buf.String() == `` {
		t.Error(`file should not be empty`)
	}

}

func eq(expected, actual interface{}, t *testing.T) {
	if fmt.Sprint(expected) != fmt.Sprint(actual) {
		t.Errorf("Unexpected value:\nexpected: %v\nactual: %v\n", expected, actual)
	}
}
