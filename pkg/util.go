package pkg

func MustEmptyOrOneLength[List any](list []List) {
	count := len(list)

	if count == 0 {
		return
	}

	if count > 1 {
		panic("list has more than one element")
	}
}
