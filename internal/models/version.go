package models

import (
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
	Build int
}

func (version Version) ToString() string {
	var stringBuilder strings.Builder

	stringBuilder.WriteString(strconv.Itoa(version.Major))
	stringBuilder.WriteString(".")
	stringBuilder.WriteString(strconv.Itoa(version.Minor))
	stringBuilder.WriteString(".")
	stringBuilder.WriteString(strconv.Itoa(version.Patch))
	stringBuilder.WriteString(".")
	stringBuilder.WriteString(strconv.Itoa(version.Build))

	return stringBuilder.String()
}
