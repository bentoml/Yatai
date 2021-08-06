package utils

import (
	"reflect"
)

func RemoveDuplicatedStrings(items []string) []string {
	res := make([]string, 0, len(items))
	RemoveDuplicatedElementsUnsafe(items, func(idx int) string {
		return items[idx]
	}, func(idx int) {
		res = append(res, items[idx])
	})
	return res
}

func RemoveDuplicatedElementsUnsafe(items interface{}, getKey func(idx int) string, putElement func(idx int)) {
	if reflect.TypeOf(items).Kind() != reflect.Slice {
		return
	}

	rv := reflect.ValueOf(items)
	seen := make(map[string]struct{}, rv.Len())
	for idx := 0; idx < rv.Len(); idx++ {
		key := getKey(idx)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		putElement(idx)
	}
}
