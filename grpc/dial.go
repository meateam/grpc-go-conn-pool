package grpc

import (
	"context"

	"github.com/meateam/grpc-go-conn-pool/transport/grpc/options"
	"google.golang.org/grpc"

	// Install grpclb
	_ "google.golang.org/grpc/balancer/grpclb"
)

// Dial returns a GRPC connection for use communicating with a GRPC
// service, configured with the given ClientOptions.
func Dial(ctx context.Context, opts ...options.ClientOption) (*grpc.ClientConn, error) {
	o, err := processAndValidateOpts(opts)
	if err != nil {
		return nil, err
	}
	if o.GRPCConnPool != nil {
		return o.GRPCConnPool.Conn(), nil
	}

	return dial(ctx, false, o)
}

// DialInsecure returns an insecure GRPC connection for use communicating
// with insecure GRPC service
func DialInsecure(ctx context.Context, opts ...options.ClientOption) (*grpc.ClientConn, error) {
	o, err := processAndValidateOpts(opts)
	if err != nil {
		return nil, err
	}
	return dial(ctx, true, o)
}

func processAndValidateOpts(opts []options.ClientOption) (*options.DialSettings, error) {
	var o options.DialSettings
	for _, opt := range opts {
		opt.Apply(&o)
	}
	if err := o.Validate(); err != nil {
		return nil, err
	}

	return &o, nil
}

type connPoolOption struct{ ConnPool }

// DialPool returns a pool of GRPC connections for the given service.
// This differs from the connection pooling implementation used by Dial, which uses a custom GRPC load balancer.
// DialPool should be used instead of Dial when a pool is used by default or a different custom GRPC load balancer is needed.
// The context and options are shared between each Conn in the pool.
// The pool size is configured using the WithGRPCConnectionPool option.
//
// This API is subject to change as we further refine requirements. It will go away if gRPC stubs accept an interface instead of the concrete ClientConn type. See https://github.com/grpc/grpc-go/issues/1287.
func DialPool(ctx context.Context, opts ...options.ClientOption) (ConnPool, error) {
	o, err := processAndValidateOpts(opts)
	if err != nil {
		return nil, err
	}
	if o.GRPCConnPool != nil {
		return o.GRPCConnPool, nil
	}
	poolSize := o.GRPCConnPoolSize
	if o.GRPCConn != nil {
		// WithGRPCConn is technically incompatible with WithGRPCConnectionPool.
		// Always assume pool size is 1 when a grpc.ClientConn is explicitly used.
		poolSize = 1
	}
	o.GRPCConnPoolSize = 0 // we don't *need* to set this to zero, but it's safe to.

	if poolSize == 0 || poolSize == 1 {
		// Fast path for common case for a connection pool with a single connection.
		conn, err := dial(ctx, false, o)
		if err != nil {
			return nil, err
		}
		return &singleConnPool{conn}, nil
	}

	pool := &roundRobinConnPool{}
	for i := 0; i < poolSize; i++ {
		conn, err := dial(ctx, false, o)
		if err != nil {
			defer pool.Close() // NOTE: error from Close is ignored.
			return nil, err
		}
		pool.conns = append(pool.conns, conn)
	}
	return pool, nil
}

func dial(ctx context.Context, insecure bool, o *options.DialSettings) (*grpc.ClientConn, error) {
	if o.GRPCConn != nil {
		return o.GRPCConn, nil
	}

	var grpcOpts []grpc.DialOption
	if insecure {
		grpcOpts = []grpc.DialOption{grpc.WithInsecure()}
	}

	// Add tracing, but before the other options, so that clients can override the
	// gRPC stats handler.
	// This assumes that gRPC options are processed in order, left to right.
	grpcOpts = append(grpcOpts, o.GRPCDialOpts...)

	endpoint := options.GetEndpoint(o)

	return grpc.DialContext(ctx, endpoint, grpcOpts...)
}
