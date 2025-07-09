package events

import (
	"stepikGoWebServices/generated/service"
	"time"
)

type StatCollector struct {
	stat service.Stat
}

func NewStatCollector() *StatCollector {
	s := &StatCollector{}
	s.Reset()

	return s
}

func (s *StatCollector) Reset() {
	s.stat = service.Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}
}

func (s *StatCollector) Collect() *service.Stat {
	var stat = &service.Stat{
		Timestamp:  time.Now().Unix(),
		ByMethod:   s.stat.ByMethod,
		ByConsumer: s.stat.ByConsumer,
	}
	s.Reset()

	return stat
}

func (s *StatCollector) Update(e *service.Event) {
	s.stat.ByConsumer[e.Consumer] += 1
	s.stat.ByMethod[e.Method] += 1
}
