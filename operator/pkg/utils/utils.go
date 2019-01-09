package utils

import (
	"fmt"
)

type StringStringMap map[string]string

func (m StringStringMap) ReplaceKeys(old, new []string) StringStringMap {

	for i, oldKey := range old {
		if v, exist := m[oldKey]; exist {
			newKey := new[i]
			delete(m, oldKey)
			m[newKey] = v
		}
	}

	return m
}

// GetNamespacedNameKey returns string "namespaced/name"
func GetNamespacedNameKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
