package util

import (
	"errors"
	"strconv"
	"strings"
)

func GetIpPortFromAddr(addr string) (ip string, port int, err error) {
	split := strings.Split(addr, ":")
	if len(split) != 2 {
		err = errors.New("格式不正确")
	} else {
		ip = split[0]
		port, err = strconv.Atoi(split[1])
	}
	return
}
