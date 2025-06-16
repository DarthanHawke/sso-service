package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type contextKey string

const (
	IPKey        contextKey = "ip"
	UserAgentKey contextKey = "user_agent"
)

func IPUserAgentInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Получаем IP из peer
	var ip string
	if p, ok := peer.FromContext(ctx); ok {
		ip = p.Addr.String()
	}

	// Получаем User-Agent из метаданных
	var userAgent string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ua := md.Get("user-agent"); len(ua) > 0 {
			userAgent = strings.Join(ua, ", ")
		}
	}

	// Добавляем в контекст
	newCtx := context.WithValue(ctx, IPKey, ip)
	newCtx = context.WithValue(newCtx, UserAgentKey, userAgent)

	return handler(newCtx, req)
}
