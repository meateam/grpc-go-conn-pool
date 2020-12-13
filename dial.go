package transport

import (
	"context"

	"google.golang.org/grpc"

	grpcTransport "github.com/meateam/grpc-go-conn-pool/transport/grpc"
	"github.com/meateam/grpc-go-conn-pool/transport/grpc/options"
)

// DialGRPC returns a GRPC connection for use communicating with a GRPC servers
// service, configured with the given ClientOptions.
func DialGRPC(ctx context.Context, opts ...options.ClientOption) (*grpc.ClientConn, error) {
	return grpcTransport.Dial(ctx, opts...)
}
