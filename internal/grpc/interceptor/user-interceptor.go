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
		userInfo := md.Get("user-agent")
		ip = userInfo[0]
		userAgent = userInfo[1]

	}

	// Добавляем в контекст
	newCtx := context.WithValue(ctx, models.IPKey, ip)
	newCtx = context.WithValue(newCtx, models.UserAgentKey, userAgent)

	return handler(newCtx, req)
}
