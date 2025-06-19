package pipeline

import "fmt"

type Job func(in, out chan interface{})

func ExecutePipeline(jobs ...Job) {
	var inputChannel = make(chan interface{}, 2)
	var outputChannel = make(chan interface{}, 2)

	for i, job := range jobs {
		fmt.Printf("Job #%d\n", i)
		go job(inputChannel, outputChannel)

		inputChannel = outputChannel
		outputChannel = make(chan interface{}, 2)
	}
}
