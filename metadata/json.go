package metadata

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"go.starlark.net/starlark"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func ValueToJson(starlarkVal starlark.Value) (string, error) {
	goVal, err := ValueToGoType(starlarkVal)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(goVal)
	return string(b), err
}

func ValueToGoType(v starlark.Value) (interface{}, error) {
	switch v.Type() {
	case "NoneType":
		return nil, nil
	case "string", "bool", "float":
		return v, nil
	case "int":
		i, ok := v.(starlark.Int).Int64()
		if !ok {
			return nil, fmt.Errorf("Cannot cast %v to int64", v)
		}
		return i, nil
	case "set", "list", "tuple":
		seq := v.(starlark.Sequence)
		vals := make([]interface{}, seq.Len())
		iter := seq.Iterate()
		defer iter.Done()
		var item starlark.Value
		for i := 0; iter.Next(&item); i++ {
			goVal, err := ValueToGoType(item)
			if err != nil {
				return nil, fmt.Errorf("Cannot convert item %v of %v into golang value: %v", item, seq, err)
			}
			vals[i] = goVal
		}
		return vals, nil
	case "dict":
		dict := v.(*starlark.Dict)
		goMap := make(map[interface{}]interface{}, dict.Len())
		for _, key := range dict.Keys() {
			goKey, err := ValueToGoType(key)
			if err != nil {
				return nil, fmt.Errorf("Cannot convert key %v of dict %v to golang value: %v", key, dict, err)
			}

			val, _, err := dict.Get(key)
			if err != nil {
				return nil, fmt.Errorf("Cannot get value for key %v from dict %v: %v", key, dict, err)
			}

			goVal, err := ValueToGoType(val)
			if err != nil {
				return nil, fmt.Errorf("Cannot convert value %v in dict %v to golang value: %v", val, dict, err)
			}

			goMap[goKey] = goVal
		}

		return goMap, nil
	default:
		return "", fmt.Errorf("Do not know how to convert %v", v)
	}
}
