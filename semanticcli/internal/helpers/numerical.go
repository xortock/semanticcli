package helpers

import "strconv"

func IsNumerical(input string) bool {
	var _, err = strconv.Atoi(input)
	return err == nil
}
