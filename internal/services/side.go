package services

func Union[T comparable](slice1, slice2 []T) []T {
	uniqueMap := make(map[T]bool)
	unionSlice := []T{}

	// Add elements from the first slice
	for _, val := range slice1 {
		if !uniqueMap[val] {
			uniqueMap[val] = true
			unionSlice = append(unionSlice, val)
		}
	}

	// Add elements from the second slice
	for _, val := range slice2 {
		if !uniqueMap[val] {
			uniqueMap[val] = true
			unionSlice = append(unionSlice, val)
		}
	}

	return unionSlice
}

func ApplyLimitAndOffset[T any](arr []T, limit, offset int) []T {
	n := len(arr)

	// If offset is greater than the array length, adjust it to return the last possible elements
	if offset > n {
		offset = n - limit
		if offset < 0 {
			offset = 0
		}
	}

	// Calculate the end index for slicing
	end := offset + limit
	if end > n {
		end = n
	}

	// Return the slice from offset to end
	return arr[offset:end]
}