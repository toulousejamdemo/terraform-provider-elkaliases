package provider

func isMapKeyInArray(list []any, key string, value any) bool {
	for _, element := range list {
		element := element.(map[string]any)
		if element[key] == value {
			return true
		}
	}
	return false
}
