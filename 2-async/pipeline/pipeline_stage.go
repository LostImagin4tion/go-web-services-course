package pipeline

import (
	"fmt"
	"sync"
)

func PipelineStage(
	inputChannel chan interface{},
	outputChannel chan interface{},
	transform func(string string) string,
) {
	waitGroup := &sync.WaitGroup{}

	for data := range inputChannel {
		waitGroup.Add(1)

		var data = fmt.Sprintf("%v", data)

		var goroutine = func(data string) {
			defer waitGroup.Done()
			outputChannel <- transform(data)
		}

		go goroutine(data)
	}

	waitGroup.Wait()
}
