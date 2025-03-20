package plugin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// HydrateFromHostConfig loads a json payload passed from Traefik into a struct. All json values passed to the wasm
// plugin from Traefik are assumed to be strings. The spf13/cast library is used for converting the strings to the
// desired type.
func HydrateFromHostConfig[T any](data json.RawMessage) (T, error) {
	var (
		out T
		raw map[string]any
	)
	if err := json.Unmarshal(data, &raw); err != nil {
		return out, err
	}

	if err := setValue(raw, reflect.ValueOf(&out).Elem()); err != nil {
		return out, err
	}

	return out, nil
}

// setValue recursively assigns a raw JSON value (from the decoded map) to dest.
func setValue(raw any, dest reflect.Value) error {
	if dest.Kind() == reflect.Ptr {
		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}

		return setValue(raw, dest.Elem())
	}

	switch dest.Kind() {
	case reflect.Struct:
		return setStruct(raw, dest)
	case reflect.Slice:
		return setSlice(raw, dest)
	case reflect.Array:
		return setArray(raw, dest)
	case reflect.Map:
		return setMap(raw, dest)
	default:
		return setPrimitive(raw, dest)
	}
}

func setStruct(raw any, dest reflect.Value) error {
	if dest.Type() == reflect.TypeOf(time.Time{}) {
		t, err := cast.ToTimeE(raw)
		if err != nil {
			return err
		}

		dest.Set(reflect.ValueOf(t))
		return nil
	}

	rawMap, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map for struct but got %T", raw)
	}

	t := dest.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		fieldVal := dest.Field(i)
		// Skip unexported fields.
		if !fieldVal.CanSet() {
			continue
		}

		fieldRaw, exists := getFieldRaw(&field, rawMap)
		if !exists {
			// If the field is a struct, supply an empty map so that its defaults are processed.
			if fieldVal.Kind() == reflect.Struct {
				fieldRaw = map[string]any{}
			} else if def := field.Tag.Get("default"); def != "" {
				fieldRaw = def
			} else if req := field.Tag.Get("required"); req == "true" {
				return fmt.Errorf("required field %s missing", field.Name)
			} else {
				// Not provided and not required; skip.
				continue
			}
		}

		if err := setValue(fieldRaw, fieldVal); err != nil {
			return fmt.Errorf("error setting field %s: %w", field.Name, err)
		}
	}

	return nil
}

func getFieldRaw(field *reflect.StructField, rawMap map[string]any) (any, bool) {
	keysToTry := []string{
		field.Name,
		strings.ToLower(field.Name),
		strings.ToUpper(field.Name),
		strings.ToTitle(field.Name),
	}

	if jsonKey := field.Tag.Get("json"); jsonKey != "" && jsonKey != "-" {
		keysToTry = append(keysToTry, jsonKey)
	}

	if len(field.Name) > 1 {
		keysToTry = append(keysToTry, strings.ToLower(field.Name[:1])+field.Name[1:])
	}

	for _, k := range keysToTry {
		if fieldRaw, exists := rawMap[k]; exists {
			return fieldRaw, true
		}
	}

	return nil, false
}

func setSlice(raw any, dest reflect.Value) error {
	rawSlice, ok := raw.([]any)
	if !ok {
		return fmt.Errorf("expected slice for value but got %T", raw)
	}

	slice := reflect.MakeSlice(dest.Type(), len(rawSlice), len(rawSlice))
	for i, item := range rawSlice {
		if err := setValue(item, slice.Index(i)); err != nil {
			return fmt.Errorf("error setting slice index %d: %w", i, err)
		}
	}

	dest.Set(slice)
	return nil
}

func setArray(raw any, dest reflect.Value) error {
	rawSlice, ok := raw.([]any)
	if !ok {
		return fmt.Errorf("expected array for value but got %T", raw)
	}

	if len(rawSlice) != dest.Len() {
		return fmt.Errorf("array length mismatch: expected %d, got %d", dest.Len(), len(rawSlice))
	}

	for i := range dest.Len() {
		if err := setValue(rawSlice[i], dest.Index(i)); err != nil {
			return fmt.Errorf("error setting array index %d: %w", i, err)
		}
	}

	return nil
}

func setMap(raw any, dest reflect.Value) error {
	rawMap, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map for value but got %T", raw)
	}

	var (
		mapVal   = reflect.MakeMap(dest.Type())
		keyType  = dest.Type().Key()
		elemType = dest.Type().Elem()
	)
	for k, v := range rawMap {
		if keyType.Kind() != reflect.String {
			return fmt.Errorf("unsupported map key type: %s", keyType.String())
		}

		keyVal := reflect.New(keyType).Elem()
		keyVal.SetString(k)

		elemVal := reflect.New(elemType).Elem()
		if err := setValue(v, elemVal); err != nil {
			return fmt.Errorf("error setting map value for key %s: %w", k, err)
		}

		mapVal.SetMapIndex(keyVal, elemVal)
	}

	dest.Set(mapVal)
	return nil
}

func setPrimitive(raw any, dest reflect.Value) error {
	var (
		converted any
		err       error
	)
	switch dest.Kind() {
	case reflect.Bool:
		converted, err = cast.ToBoolE(raw)
	case reflect.Int:
		converted, err = cast.ToIntE(raw)
	case reflect.Int8:
		converted, err = cast.ToInt8E(raw)
	case reflect.Int16:
		converted, err = cast.ToInt16E(raw)
	case reflect.Int32:
		converted, err = cast.ToInt32E(raw)
	case reflect.Int64:
		if dest.Type() == reflect.TypeOf(time.Duration(0)) {
			converted, err = cast.ToDurationE(raw)
			break
		}

		converted, err = cast.ToInt64E(raw)
	case reflect.Uint:
		converted, err = cast.ToUintE(raw)
	case reflect.Uint8:
		converted, err = cast.ToUint8E(raw)
	case reflect.Uint16:
		converted, err = cast.ToUint16E(raw)
	case reflect.Uint32:
		converted, err = cast.ToUint32E(raw)
	case reflect.Uint64:
		converted, err = cast.ToUint64E(raw)
	case reflect.Float32:
		converted, err = cast.ToFloat32E(raw)
	case reflect.Float64:
		converted, err = cast.ToFloat64E(raw)
	case reflect.String:
		converted, err = cast.ToStringE(raw)
	default:
		err = fmt.Errorf("unsupported type: %s", dest.Type().String())
	}
	if err != nil {
		return err
	}

	val := reflect.ValueOf(converted)
	if !val.Type().AssignableTo(dest.Type()) {
		if !val.Type().ConvertibleTo(dest.Type()) {
			return fmt.Errorf("cannot assign value of type %s to destination type %s", val.Type(), dest.Type())
		}

		val = val.Convert(dest.Type())
	}

	dest.Set(val)
	return nil
}
