// Copyright 2019 Job Stoit. All rights reserved.

/*
Package strct a simplified golang reflect package

Usage

Get the values you need using the scanner
	func TagEnvParser(obj interface{}, parsevalue string) error {
		return strct.Scan(obj, func(field reflect.StructField, value *reflect.Value) error {
			tagVal := field.Tag.Get(`env`)
			if tagVal == `` {
				return nil
			}

			setValue(value, os.GetEnv(tagval))
		})
	}

*/
package strct

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ErrNoPtr gets thrown if the inserted object is not a pointer or a struct type
var ErrNoPtr = fmt.Errorf(`insert is not a pointer or a struct`)

// Scan scans each structs attribute
func Scan(obj interface{}, action func(reflect.StructField, *reflect.Value) error) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrNoPtr
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return ErrNoPtr
	}

	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		switch f.Kind() {
		case reflect.Ptr:
			if f.Elem().Kind() != reflect.Struct {
				break
			}

			f = f.Elem()
			fallthrough

		case reflect.Struct:
			if !f.Addr().CanInterface() {
				continue
			}

			if err := Scan(f.Addr().Interface(), action); err != nil {
				return err
			}
		}

		if !f.CanSet() {
			continue
		}

		if err := action(t.Field(i), &f); err != nil {
			return err
		}

	}
	return nil
}

// Parse sets a string value to the given value
func Parse(val string, fv *reflect.Value) error {
	if val == `` {
		return nil
	}

	switch fv.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		fv.SetBool(v)

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(val, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetFloat(v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if t := fv.Type(); t.PkgPath() == `time` && t.Name() == `Duration` {
			v, err := time.ParseDuration(val)
			if err != nil {
				return err
			}
			fv.SetInt(int64(v))
		} else {
			v, err := strconv.ParseInt(val, 0, fv.Type().Bits())
			if err != nil {
				return err
			}
			fv.SetInt(v)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(val, 0, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetUint(v)

	case reflect.String:
		fv.SetString(val)

	case reflect.Slice:
		parts := strings.Split(val, `;`)
		slice := reflect.MakeSlice(fv.Type(), len(parts), len(parts))
		for i, part := range parts {
			part = strings.TrimSpace(part)
			in := slice.Index(i)
			if err := Parse(part, &in); err != nil {
				return err
			}
		}
		fv.Set(slice)
	}
	return nil
}
