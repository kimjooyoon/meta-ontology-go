package main

import (
	"fmt"
	"path"
	"slices"
	"sort"
)

func buildObjects(files []trackedFile) map[string]*storedObject {
	objects := make(map[string]*storedObject)
	for _, file := range files {
		if _, exists := objects[file.objectSHA]; exists {
			continue
		}
		extension := ".blob"
		if objectClass(file.kind, file.language) == "source" {
			extension = ".gooo"
		}
		objects[file.objectSHA] = &storedObject{id: file.objectSHA, ext: extension, data: file.data}
	}
	return objects
}

func assignBacking(objects map[string]*storedObject) error {
	group := make([]*storedObject, 0, len(objects))
	for _, object := range objects {
		group = append(group, object)
	}
	sort.Slice(group, func(i, j int) bool { return group[i].id < group[j].id })
	return assignGroup(group, nil, 0)
}

func assignGroup(group []*storedObject, prefix []string, depth int) error {
	if len(group) <= 10 {
		for _, object := range group {
			parts := append([]string{"objects"}, prefix...)
			object.backing = path.Join(append(parts, object.id+object.ext)...)
		}
		return nil
	}
	partitions := make(map[byte][]*storedObject)
	for _, object := range group {
		key := decimalKey(object.id)
		if depth >= len(key) {
			return fmt.Errorf("object radix collision at %s", object.id)
		}
		partitions[key[depth]] = append(partitions[key[depth]], object)
	}
	digits := make([]byte, 0, len(partitions))
	for digit := range partitions {
		digits = append(digits, digit)
	}
	slices.Sort(digits)
	for _, digit := range digits {
		next := append(append([]string{}, prefix...), string(digit))
		if err := assignGroup(partitions[digit], next, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func decimalKey(hexID string) string {
	key := make([]byte, 0, len(hexID)*2)
	for _, digit := range []byte(hexID) {
		value := digit - '0'
		if digit >= 'a' {
			value = digit - 'a' + 10
		}
		key = append(key, '0'+value/10, '0'+value%10)
	}
	return string(key)
}
