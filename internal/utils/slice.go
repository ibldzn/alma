package utils

func Dedup[T comparable](input []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(input))

	for _, item := range input {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
