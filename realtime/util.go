package realtime

func Contains(array []int, find int) bool {
	for _, value := range array {
		if value == find {
			return true
		}
	}
	return false
}

func Remove(array []int, find int) {
	for i, value := range array {
		if value == find {
			array[i] = -1
		}
	}
}
