package pipeline

import (
	"fmt"
	"sort"
	"strings"
)

func CombineResults(
	inputChannel chan interface{},
	outputChannel chan interface{},
) {
	fmt.Printf("CombineResults start")

	var results = make([]string, 0)
	for rawData := range inputChannel {
		var data, ok = rawData.(string)
		if !ok {
			panic("Cant convert input data to string")
		}
		results = append(results, data)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})

	var result = strings.Join(results, "_")
	fmt.Printf("Combine results %s\n", result)

	outputChannel <- result
	close(outputChannel)
}
