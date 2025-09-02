package main

import (
	"fmt"
	"slices"
)

func main() {
	input := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	fmt.Println(groupAnagrams(input))
}

func groupAnagrams(strs []string) map[string][]string {
	group := make(map[string][]string)

	for _, s := range strs {
		sorted := sortStr(s)
		group[sorted] = append(group[sorted], s)
	}

	for key, words := range group {
		if len(words) < 2 {
			delete(group, key)
		} else {
			slices.Sort(words)
		}
	}

	return group
}

func sortStr(s string) string {
	b := []rune(s)
	slices.Sort(b)
	return string(b)
}
