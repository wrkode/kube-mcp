package config

import (
	"reflect"
	"time"
)

// mergeConfig performs a deep merge of src into dst.
func mergeConfig(dst, src *Config) {
	if src == nil {
		return
	}

	mergeStruct(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem())
}

// mergeStruct merges two struct values, with src taking precedence.
func mergeStruct(dst, src reflect.Value) {
	if !dst.IsValid() || !src.IsValid() {
		return
	}

	if dst.Kind() != reflect.Struct || src.Kind() != reflect.Struct {
		return
	}

	srcType := src.Type()

	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		srcFieldType := srcType.Field(i)

		// Skip unexported fields
		if !srcField.CanInterface() {
			continue
		}

		// Find corresponding field in dst
		dstField := dst.FieldByName(srcFieldType.Name)
		if !dstField.IsValid() || !dstField.CanSet() {
			continue
		}

		// Merge based on type
		switch srcField.Kind() {
		case reflect.Struct:
		// Special handling for Duration and time.Duration
		if srcField.Type() == reflect.TypeOf(Duration(0)) || srcField.Type() == reflect.TypeOf(time.Duration(0)) {
			if !srcField.IsZero() {
				dstField.Set(srcField)
			}
		} else {
				mergeStruct(dstField, srcField)
			}

		case reflect.Slice:
			if !srcField.IsNil() && srcField.Len() > 0 {
				// For slices, replace if src has values
				dstField.Set(srcField)
			}

		case reflect.Map:
			if !srcField.IsNil() && srcField.Len() > 0 {
				// For maps, merge entries
				if dstField.IsNil() {
					dstField.Set(reflect.MakeMap(dstField.Type()))
				}
				for _, key := range srcField.MapKeys() {
					dstField.SetMapIndex(key, srcField.MapIndex(key))
				}
			}

		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String:
			// For primitive types, use src value if not zero
			if !srcField.IsZero() {
				dstField.Set(srcField)
			}

		case reflect.Ptr:
			if !srcField.IsNil() {
				if dstField.IsNil() {
					dstField.Set(reflect.New(dstField.Type().Elem()))
				}
				mergeStruct(dstField.Elem(), srcField.Elem())
			}

		default:
			// For other types, replace if src is not zero
			if !srcField.IsZero() {
				dstField.Set(srcField)
			}
		}
	}
}
