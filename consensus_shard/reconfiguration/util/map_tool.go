package util

func DeepCopyIpNodeTable(original map[uint64]map[uint64]string) map[uint64]map[uint64]string {
	c := make(map[uint64]map[uint64]string)
	for key, innerMap := range original {
		innerCopy := make(map[uint64]string)
		for innerKey, value := range innerMap {
			innerCopy[innerKey] = value
		}
		c[key] = innerCopy
	}
	return c
}
