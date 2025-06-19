package main

import (
	"fmt"
	"stepikGoWebServices/pipeline"
	"sync/atomic"
	"testing"
	"time"
)

/*
	Тест, предложенный одним из учащихся курса, Ilya Boltnev

	В чем его преимущество по сравнению с TestPipeline?
	1. Он проверяет то, что все функции действительно выполнились
	2. Он дает представление о влиянии time.Sleep в одном из звеньев конвейера на время работы

	возможно кому-то будет легче с ним
	при правильной реализации ваш код конечно же должен его проходить
*/

func testByIlia(t *testing.T) {

	var received uint32
	freeFlowJobs := []Job{
		Job(func(in, out chan interface{}) {
			out <- uint32(1)
			out <- uint32(3)
			out <- uint32(4)
		}),
		Job(func(in, out chan interface{}) {
			for val := range in {
				out <- val.(uint32) * 3
				time.Sleep(time.Millisecond * 100)
			}
		}),
		Job(func(in, out chan interface{}) {
			for val := range in {
				fmt.Println("collected", val)
				atomic.AddUint32(&received, val.(uint32))
			}
		}),
	}

	start := time.Now()

	pipeline.ExecutePipeline(freeFlowJobs...)

	end := time.Since(start)

	expectedTime := time.Millisecond * 350

	if end > expectedTime {
		t.Errorf("execition too long\nGot: %s\nExpected: <%s", end, expectedTime)
	}

	if received != (1+3+4)*3 {
		t.Errorf("f3 have not collected inputs, received = %d", received)
	}
}
