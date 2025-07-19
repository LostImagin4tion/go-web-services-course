package services

import (
	"google.golang.org/grpc"
	"stepikGoWebServices/events"
	"stepikGoWebServices/generated/service"
	"time"
)

var (
	AdminServiceMethods = []string{
		service.Admin_Logging_FullMethodName,
		service.Admin_Statistics_FullMethodName,
	}
)

type AdminService struct {
	service.AdminServer

	eventSubscribers *events.EventSubscribersManager
}

func NewAdminService(
	eventSubscribers *events.EventSubscribersManager,
) *AdminService {
	return &AdminService{
		eventSubscribers: eventSubscribers,
	}
}

func (as *AdminService) Logging(
	_ *service.Nothing,
	outputStream grpc.ServerStreamingServer[service.Event],
) error {
	var id, eventChan = as.eventSubscribers.NewSub()
	defer as.eventSubscribers.RemoveSub(id)

	for event := range eventChan {
		var err = outputStream.Send(event)
		if err != nil {
			return err
		}
	}

	return nil
}

func (as *AdminService) Statistics(
	interval *service.StatInterval,
	outputStream grpc.ServerStreamingServer[service.Stat],
) error {
	var id, eventsChan = as.eventSubscribers.NewSub()
	defer as.eventSubscribers.RemoveSub(id)

	var statCollector = events.NewStatCollector()
	var ticker = time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case e, ok := <-eventsChan:
			if ok {
				statCollector.Update(e)
			} else {
				return nil
			}

		case <-ticker.C:
			var err = outputStream.Send(statCollector.Collect())
			if err != nil {
				return err
			}
		}
	}
}
