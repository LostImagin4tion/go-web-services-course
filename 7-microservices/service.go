package main

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"stepikGoWebServices/acl"
	"stepikGoWebServices/events"
	"stepikGoWebServices/generated/service"
	"stepikGoWebServices/middleware"
	"stepikGoWebServices/services"
)

func StartMyMicroservice(
	ctx context.Context,
	address string,
	aclRaw string,
) error {
	var aclData, err = acl.NewAclData(aclRaw)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	var eventSubscribersManager = events.NewEventSubscribersManager()

	serviceMiddleware := middleware.NewMiddleware(
		aclData,
		eventSubscribersManager,
	)

	var server = grpc.NewServer(serviceMiddleware.ServerOptions...)

	service.RegisterAdminServer(server, services.NewAdminService(eventSubscribersManager))
	service.RegisterBusinessLogicServer(server, services.NewBusinessLogicService())

	go server.Serve(listener)

	go func(ctx context.Context, server *grpc.Server) {
		var out = ctx.Done()
		if out != nil {
			<-ctx.Done()
			eventSubscribersManager.RemoveAll()
			server.GracefulStop()

		}
	}(ctx, server)

	return nil
}
