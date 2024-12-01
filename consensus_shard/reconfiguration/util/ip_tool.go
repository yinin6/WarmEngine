package util

import (
	"strconv"
	"strings"
)

func UpdateIpTable(ip_nodeTable map[uint64]map[uint64]string) map[uint64]map[uint64]string {
	res := make(map[uint64]map[uint64]string)
	for i, v := range ip_nodeTable {
		res[i] = make(map[uint64]string)
		for j, x := range v {
			res[i][j] = x
			if j == 2147483647 {
				continue
			}
			addr := strings.Split(x, ":")[0]
			s := strings.Split(x, ":")[1]
			port, _ := strconv.Atoi(s)
			port += 10000
			res[i][j] = addr + ":" + strconv.Itoa(port)
		}
	}
	return res
}
