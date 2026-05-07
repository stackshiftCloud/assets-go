package stackshift

import "net/url"

func setQuery(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func stripEmpty(value any) any {
	item, ok := value.(map[string]any)
	if !ok {
		return value
	}
	clean := map[string]any{}
	for key, entry := range item {
		if entry != nil && entry != "" && entry != 0 {
			clean[key] = entry
		}
	}
	return clean
}
