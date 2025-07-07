package utils

func SliceMap[Input any, Output any](arr []Input, fn func(Input) Output) []Output {
	var result []Output

	for _, v := range arr {
		result = append(result, fn(v))
	}

	return result
}
