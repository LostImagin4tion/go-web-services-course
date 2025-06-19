package pipeline

import (
	"fmt"
	"strconv"
	"strings"
)

func SingleHash(
	inputChannel chan interface{},
	outputChannel chan interface{},
	dataSignerMd5 func(data string) string,
	dataSignerCrc32 func(data string) string,
) {
	fmt.Printf("SingleHash start")

	var stringBuilder strings.Builder

	var md5HashChannel = make(chan string, 1)
	var crc32Md5HashChannel = make(chan string, 1)
	var crc32HashChannel = make(chan string, 1)

	for rawData := range inputChannel {
		var intData, ok = rawData.(int)
		if !ok {
			panic("Cant convert input data to int")
		}
		var data = strconv.Itoa(intData)

		fmt.Printf("SingleHash data %s\n", data)

		go func() {
			var md5Hash = dataSignerCrc32(data)
			fmt.Printf("SingleHash md5(data) %s\n", md5Hash)
			md5HashChannel <- md5Hash
		}()

		go func() {
			var md5Hash = <-md5HashChannel
			var crcMd5Hash = dataSignerCrc32(md5Hash)
			crc32Md5HashChannel <- dataSignerCrc32(md5Hash)
		}()

		go func() {
			crc32HashChannel <- dataSignerCrc32(data)
		}()

		var crcMd5Hash = dataSignerCrc32(md5Hash)
		fmt.Printf("SingleHash crc32(md5(data)) %s\n", crcMd5Hash)

		var crcHash = dataSignerCrc32(data)
		fmt.Printf("SingleHash crc32(data) %s\n", crcHash)

		stringBuilder.WriteString(crcHash)
		stringBuilder.WriteString("~")
		stringBuilder.WriteString(crcMd5Hash)

		var result = stringBuilder.String()
		fmt.Printf("SingleHash result %s\n\n", result)

		outputChannel <- result

		stringBuilder.Reset()
	}
	close(outputChannel)
}
