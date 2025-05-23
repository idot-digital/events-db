package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor returns a new unary server interceptor for authentication
func AuthInterceptor(authToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth if no token is configured
		if authToken == "" {
			return handler(ctx, req)
		}

		// Get metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		// Get authorization token from metadata
		auth := md.Get("authorization")
		if len(auth) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		// Check if token matches
		if auth[0] != authToken {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization token")
		}

		return handler(ctx, req)
	}
}

// StreamAuthInterceptor returns a new stream server interceptor for authentication
func StreamAuthInterceptor(authToken string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip auth if no token is configured
		if authToken == "" {
			return handler(srv, ss)
		}

		// Get metadata from context
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		// Get authorization token from metadata
		auth := md.Get("authorization")
		if len(auth) == 0 {
			return status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		// Check if token matches
		if auth[0] != authToken {
			return status.Error(codes.Unauthenticated, "invalid authorization token")
		}

		return handler(srv, ss)
	}
}
