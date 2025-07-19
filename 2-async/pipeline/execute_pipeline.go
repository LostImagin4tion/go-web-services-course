package pipeline

import (
	"fmt"
	"sync"
)

type Job func(in, out chan interface{})

func ExecutePipeline(jobs ...Job) {
	var inputChannel chan interface{}
	var outputChannel chan interface{}

	var waitGroup = &sync.WaitGroup{}

	for i, job := range jobs {
		waitGroup.Add(1)

		fmt.Println("Job", i)

		inputChannel = outputChannel
		outputChannel = make(chan interface{}, 100)

		var goroutine = func(
			job Job,
			inputChannel chan interface{},
			outputChannel chan interface{},
		) {
			defer waitGroup.Done()
			defer close(outputChannel)

			job(inputChannel, outputChannel)
		}

		go goroutine(job, inputChannel, outputChannel)
	}

	fmt.Println("ExecutePipeline waiting for the end")
	waitGroup.Wait()
	fmt.Println("ExecutePipeline completed")
}
