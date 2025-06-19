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
	fmt.Println("CombineResults start")

	var results = make([]string, 0)

	for rawData := range inputChannel {
		var data = fmt.Sprintf("%v", rawData)
		fmt.Println("CombineResults received data", data)
		results = append(results, data)
	}

	sort.Strings(results)

	var result = strings.Join(results, "_")
	fmt.Println("Combine results", result)

	outputChannel <- result
}
