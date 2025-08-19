package interceptor

import (
	"context"
	"sso-service/internal/models"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func IPUserAgentInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	var ip string
	var userAgent string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ips := md.Get("x-client-ip"); len(ips) > 0 {
			ip = ips[0]
		}
		if agents := md.Get("x-client-agent"); len(agents) > 0 {
			userAgent = agents[0]
		}
	}

	// Добавляем в контекст
	ctx = context.WithValue(ctx, models.IPKey, ip)
	ctx = context.WithValue(ctx, models.UserAgentKey, userAgent)

	return handler(ctx, req)
}
