package pipeline

import (
	"fmt"
	"sync"
)

func SingleHash(
	inputChannel chan interface{},
	outputChannel chan interface{},
	dataSignerMd5 func(data string) string,
	dataSignerCrc32 func(data string) string,
) {
	fmt.Println("SingleHash start")

	var waitGroup = &sync.WaitGroup{}
	var mutex = &sync.Mutex{}

	PipelineStage(
		inputChannel,
		outputChannel,
		func(data string) string {
			waitGroup.Add(3)

			var crc32HashChannel = make(chan string, 1)

			var md5HashChannel = make(chan string, 1)
			var crc32Md5HashChannel = make(chan string, 1)

			go func() {
				defer waitGroup.Done()

				var crc32Hash = dataSignerCrc32(data)
				fmt.Println("SingleHash crc32(data)", crc32Hash)
				crc32HashChannel <- crc32Hash
			}()

			go func() {
				defer waitGroup.Done()

				mutex.Lock()
				var md5Hash = dataSignerMd5(data)
				mutex.Unlock()
				fmt.Println("SingleHash md5(data)", md5Hash)
				md5HashChannel <- md5Hash
			}()

			go func() {
				defer waitGroup.Done()

				var md5Hash = <-md5HashChannel
				var crcMd5Hash = dataSignerCrc32(md5Hash)
				fmt.Println("SingleHash crc32(md5(data))", crcMd5Hash)
				crc32Md5HashChannel <- crcMd5Hash
			}()

			waitGroup.Wait()

			var crc32Hash = <-crc32HashChannel
			var crc32Md5Hash = <-crc32Md5HashChannel

			var result = fmt.Sprintf("%s~%s", crc32Hash, crc32Md5Hash)
			fmt.Println("SingleHash result", result)
			fmt.Println()

			return result
		},
	)
}
