package sdk

import (
	"fmt"
	"strconv"
)

var u Utility

type Utility struct{}

func (u Utility) Transform(target any) map[string]any {
	if target0, ok := target.(map[string]any); ok {
		return target0
	} else if target0, ok := target.([]any); ok {
		// -
		result := make(map[string]any)
		// -
		for itemIndex, item := range target0 {
			result[strconv.Itoa(itemIndex)] = item
		}
		// -
		return result
	} else {
		return nil
	}
}

func (u Utility) PileDown(target any, keyPrefix string) map[string]any {
	// -
	result := make(map[string]any)
	// -
	for key, value := range u.Transform(target) {
		// 构建新的前缀
		// > -
		var prefixedKey string
		// > -
		if keyPrefix == "" {
			prefixedKey = key
		} else {
			prefixedKey = fmt.Sprintf("%s.%s", keyPrefix, key)
		}

		// 处理子元素
		// > -
		target0 := u.Transform(value)
		// > -
		if target0 == nil {
			result[prefixedKey] = value
		} else {
			// -
			result0 := u.PileDown(target0, prefixedKey)
			// -
			for key0, value0 := range result0 {
				result[key0] = value0
			}
		}
	}
	// -
	return result
}
