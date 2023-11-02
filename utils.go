package lineschema

import (
	"strings"
)

// BaseName 获取最后.后的文本
func BaseName(fullname string) (baseName string) {
	baseName = fullname
	lastDotIndex := strings.LastIndex(baseName, ".")
	if lastDotIndex > -1 {
		baseName = baseName[lastDotIndex+1:]
	}
	return baseName
}

// Namespace 获取最后.前的文本
func Namespace(fullname string) (namespace string) {
	lastDotIndex := strings.LastIndex(fullname, ".")
	if lastDotIndex > -1 {
		namespace = fullname[:lastDotIndex]
	}
	return namespace
}
