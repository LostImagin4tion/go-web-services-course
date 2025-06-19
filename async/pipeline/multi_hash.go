package pipeline

import (
	"fmt"
	"strconv"
	"strings"
)

func MultiHash(
	inputChannel chan interface{},
	outputChannel chan interface{},
	dataSignerCrc32 func(data string) string,
) {
	fmt.Printf("MultiHash start")

	const th = 6

	var stringBuilder strings.Builder

	for rawData := range inputChannel {
		var data, ok = rawData.(string)
		if !ok {
			panic("Cant convert input data to string")
		}

		fmt.Printf("MultiHash data %s\n", data)

		var results = make([]string, th)
		for i := range th {
			var crcHash = dataSignerCrc32(strconv.Itoa(i) + data)
			fmt.Printf("MultiHash crc32(th+data) %d %s\n", i, crcHash)
			results[i] = crcHash
		}

		for _, result := range results {
			stringBuilder.WriteString(result)
		}

		var result = stringBuilder.String()
		fmt.Printf("MultiHash result %s\n\n", result)

		outputChannel <- result

		stringBuilder.Reset()
	}
	close(outputChannel)
}
