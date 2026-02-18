package config

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var interpolationPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}`)

func InterpolateString(input string, vars map[string]string) (string, error) {
	var unresolved []string

	output := interpolationPattern.ReplaceAllStringFunc(input, func(match string) string {
		groups := interpolationPattern.FindStringSubmatch(match)
		if len(groups) != 2 {
			return match
		}
		key := strings.TrimSpace(groups[1])
		value, ok := vars[key]
		if !ok {
			unresolved = append(unresolved, key)
			return match
		}
		return value
	})

	if len(unresolved) > 0 {
		return "", fmt.Errorf("unresolved variable(s): %s", strings.Join(unique(unresolved), ", "))
	}

	return output, nil
}

func InterpolateAny(target any, vars map[string]string) error {
	return interpolateValue(reflect.ValueOf(target), vars)
}

func interpolateValue(v reflect.Value, vars map[string]string) error {
	if !v.IsValid() {
		return nil
	}

	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return nil
		}
		return interpolateValue(v.Elem(), vars)
	case reflect.Interface:
		if v.IsNil() {
			return nil
		}
		val := v.Elem()
		if val.Kind() == reflect.String {
			out, err := InterpolateString(val.String(), vars)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(out))
			return nil
		}
		if val.CanAddr() {
			return interpolateValue(val.Addr(), vars)
		}
		return interpolateValue(val, vars)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if !field.CanSet() && field.Kind() != reflect.Struct && field.Kind() != reflect.Pointer && field.Kind() != reflect.Slice && field.Kind() != reflect.Map && field.Kind() != reflect.Interface {
				continue
			}
			if err := interpolateValue(field, vars); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err := interpolateValue(v.Index(i), vars); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			mapKey := iter.Key()
			mapVal := iter.Value()
			replaced, err := interpolateMapValue(mapVal, vars)
			if err != nil {
				return err
			}
			v.SetMapIndex(mapKey, replaced)
		}
		return nil
	case reflect.String:
		out, err := InterpolateString(v.String(), vars)
		if err != nil {
			return err
		}
		v.SetString(out)
		return nil
	default:
		return nil
	}
}

func interpolateMapValue(v reflect.Value, vars map[string]string) (reflect.Value, error) {
	if !v.IsValid() {
		return v, nil
	}

	switch v.Kind() {
	case reflect.String:
		out, err := InterpolateString(v.String(), vars)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(out).Convert(v.Type()), nil
	case reflect.Interface:
		if v.IsNil() {
			return v, nil
		}
		inner := v.Elem()
		replaced, err := interpolateMapValue(inner, vars)
		if err != nil {
			return reflect.Value{}, err
		}
		return replaced, nil
	case reflect.Map:
		cloned := reflect.MakeMap(v.Type())
		iter := v.MapRange()
		for iter.Next() {
			replaced, err := interpolateMapValue(iter.Value(), vars)
			if err != nil {
				return reflect.Value{}, err
			}
			cloned.SetMapIndex(iter.Key(), replaced)
		}
		return cloned, nil
	case reflect.Slice:
		cloned := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			replaced, err := interpolateMapValue(v.Index(i), vars)
			if err != nil {
				return reflect.Value{}, err
			}
			cloned.Index(i).Set(replaced)
		}
		return cloned, nil
	default:
		return v, nil
	}
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
