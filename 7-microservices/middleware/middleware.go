package middleware

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"stepikGoWebServices/acl"
	"stepikGoWebServices/events"
	"stepikGoWebServices/generated/service"
	"time"
)

type ServiceMiddleware struct {
	ServerOptions    []grpc.ServerOption
	acl              *acl.Data
	eventSubscribers *events.EventSubscribersManager
}

func NewMiddleware(
	aclData *acl.Data,
	eventSubscribers *events.EventSubscribersManager,
) *ServiceMiddleware {
	mid := &ServiceMiddleware{
		acl:              aclData,
		eventSubscribers: eventSubscribers,
	}

	mid.ServerOptions = []grpc.ServerOption{
		grpc.UnaryInterceptor(mid.unaryInterceptor),
		grpc.StreamInterceptor(mid.streamInterceptor),
	}

	return mid
}

func (m *ServiceMiddleware) intercept(
	ctx context.Context,
	method string,
) error {
	var meta, _ = metadata.FromIncomingContext(ctx)
	var consumerNames = meta.Get("consumer")

	if len(consumerNames) == 0 {
		return status.Error(codes.Unauthenticated, "consumer is empty")
	}

	var consumer = consumerNames[0]

	var host string
	if p, ok := peer.FromContext(ctx); ok {
		host = p.Addr.String()
	}

	m.eventSubscribers.Notify(&service.Event{
		Method:    method,
		Consumer:  consumer,
		Host:      host,
		Timestamp: time.Now().Unix(),
	})

	var err = m.acl.ValidateAcl(consumer, method)
	if err != nil {
		return err
	}

	return nil
}

func (m *ServiceMiddleware) unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if err := m.intercept(ctx, info.FullMethod); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (m *ServiceMiddleware) streamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	if err := m.intercept(ss.Context(), info.FullMethod); err != nil {
		return err
	}

	return handler(srv, ss)
}
