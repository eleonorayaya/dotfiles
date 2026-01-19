package util

func MergeStringAnyMap(existing, defaults map[string]any) map[string]any {
	if existing == nil {
		return CopyMap(defaults)
	}
	if defaults == nil {
		return existing
	}

	result := make(map[string]any)

	for k, v := range existing {
		result[k] = v
	}

	for k, defaultVal := range defaults {
		if existingVal, exists := result[k]; exists {
			existingMap, existingIsMap := existingVal.(map[string]any)
			defaultMap, defaultIsMap := defaultVal.(map[string]any)

			if existingIsMap && defaultIsMap {
				result[k] = MergeStringAnyMap(existingMap, defaultMap)
			}
		} else {
			result[k] = CopyValue(defaultVal)
		}
	}

	return result
}

func CopyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any)
	for k, v := range m {
		result[k] = CopyValue(v)
	}
	return result
}

func CopyValue(v any) any {
	if m, ok := v.(map[string]any); ok {
		return CopyMap(m)
	}
	return v
}
