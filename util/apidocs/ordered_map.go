package apidocs

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
)

// OrderedUnmarshalJSON deserializes JSON data while preserving the order of keys.
func OrderedUnmarshalJSON(b []byte, element interface{}, rank interface{}) error {
	if err := json.Unmarshal(b, element); err != nil {
		return err
	}

	elemObject := reflect.ValueOf(element).Elem()
	index := make(map[string]int, elemObject.Len())
	keys := make([]string, 0, elemObject.Len())
	for _, key := range elemObject.MapKeys() {
		keys = append(keys, key.String())
		nk, _ := json.Marshal(key.String()) // Escape the key
		index[key.String()] = bytes.Index(b, nk)
	}

	// Sort keys based on their occurrence index in the JSON data.
	sort.Slice(keys, func(i, j int) bool {
		return index[keys[i]] < index[keys[j]]
	})

	rankObject := reflect.ValueOf(rank).Elem()
	if rankObject.IsNil() {
		rankObject.Set(reflect.MakeMap(rankObject.Type()))
	}

	// Assign rank values to keys based on their sorted order.
	for idx, key := range keys {
		rankObject.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(idx+1))
	}

	return nil
}

// OrderedMarshalJSON serializes data to JSON while preserving the order of keys.
func OrderedMarshalJSON(element interface{}, rank map[string]int) ([]byte, error) {
	if len(rank) == 0 {
		return json.Marshal(element)
	}

	var keys []string
	elemObject := reflect.ValueOf(element)
	for _, key := range elemObject.MapKeys() {
		keys = append(keys, key.String())
	}

	// Sort keys based on their rank values.
	sort.Slice(keys, func(i, j int) bool {
		return rank[keys[i]] < rank[keys[j]]
	})

	buf := &bytes.Buffer{}
	buf.WriteRune('{')
	l := len(keys)
	for idx, key := range keys {
		k, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(k)
		buf.WriteRune(':')

		v, err := json.Marshal(elemObject.MapIndex(reflect.ValueOf(key)).Interface())
		if err != nil {
			return nil, err
		}
		buf.Write(v)

		if idx != l-1 {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune('}')
	return buf.Bytes(), nil
}
