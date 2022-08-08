package utils

import (
	"github.com/sirupsen/logrus"
)

func DoOrDie(err error) {
	if err != nil {
		logrus.Fatalf("%+v", err)
	}
}

func Pointer(s string) *string {
	return &s
}

func AddKey(dict map[string]bool, key string) map[string]bool {
	out := map[string]bool{}
	for k, v := range dict {
		out[k] = v
	}
	out[key] = true
	return out
}

func StringPrefix(s string, chars int) string {
	if len(s) <= chars {
		return s
	}
	return s[:chars]
}

func CopySlice[A any](s []A) []A {
	newCopy := make([]A, len(s))
	copy(newCopy, s)
	return newCopy
}
