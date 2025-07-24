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

		// Безопасное получение User-Agent
		if agents := md.Get("x-client-agent"); len(agents) > 0 {
			userAgent = agents[0]
		}

	}

	// Добавляем в контекст
	newCtx := context.WithValue(ctx, models.IPKey, ip)
	newCtx = context.WithValue(newCtx, models.UserAgentKey, userAgent)

	return handler(newCtx, req)
}
