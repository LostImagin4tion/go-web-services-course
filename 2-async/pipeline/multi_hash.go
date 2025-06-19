package pipeline

import (
	"fmt"
	"strings"
	"sync"
)

func MultiHash(
	inputChannel chan interface{},
	outputChannel chan interface{},
	dataSignerCrc32 func(data string) string,
) {
	fmt.Println("MultiHash start")
	const threads = 6

	PipelineStage(
		inputChannel,
		outputChannel,
		func(data string) string {
			var results = make([]string, threads)

			var waitGroup = &sync.WaitGroup{}
			var mutex = &sync.Mutex{}

			for i := range threads {
				waitGroup.Add(1)

				go func() {
					defer waitGroup.Done()

					var crcHash = dataSignerCrc32(
						fmt.Sprintf("%v%v", i, data),
					)
					fmt.Println("MultiHash crc32(threads+data)", i, crcHash)

					mutex.Lock()
					results[i] = crcHash
					mutex.Unlock()
				}()
			}

			waitGroup.Wait()

			mutex.Lock()
			var result = strings.Join(results, "")
			mutex.Unlock()

			fmt.Println("MultiHash result", result)
			fmt.Println()

			return result
		},
	)
}
