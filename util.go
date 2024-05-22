package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func SortArtists(input map[string]artist) []string {
	// Iterate over the map keys and sort them alphabetically
	names := make([]string, 0, len(input))

	for n := range input {
		names = append(names, n)
	}

	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})

	return names
}

func Log(data ...interface{}) error {
	file, err := os.Create("/tmp/debug.log")
	if err != nil {
		return err
	}
	defer file.Close()

	for _, datum := range data {
		_, err = file.WriteString(fmt.Sprintf("%v ", datum))
		if err != nil {
			return err
		}
	}

	return nil
}
