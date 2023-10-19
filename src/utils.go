package main

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func filter[T any](list []T, test func(T) bool) (ret []T) {
	for _, el := range list {
		if test(el) {
			ret = append(ret, el)
		}
	}
	return
}

func fold[T any, R any](list []T, base R, combine func(R, T) R) (ret R) {
	for _, el := range list {
		ret = combine(ret, el)
	}
	return
}
