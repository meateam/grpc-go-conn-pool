package options

import (
	"errors"

	"github.com/meateam/grpc-go-conn-pool/grpc/types"
	"google.golang.org/grpc"
)

// DialSettings holds information needed to establish a connection with a service.
type DialSettings struct {
	Endpoint         string
	GRPCDialOpts     []grpc.DialOption
	GRPCConn         *grpc.ClientConn
	GRPCConnPool     types.ConnPool
	GRPCConnPoolSize int
	NoAuth           bool
	SkipValidation   bool
}

// A ClientOption is an option for a Google API client.
type ClientOption interface {
	Apply(*DialSettings)
}

// WithGRPCConn returns a ClientOption that specifies the gRPC client
// connection to use as the basis of communications. This option may only be
// used with services that support gRPC as their communication transport. When
// used, the WithGRPCConn option takes precedent over all other supplied
// options.
func WithGRPCConn(conn *grpc.ClientConn) ClientOption {
	return withGRPCConn{conn}
}

type withGRPCConn struct{ conn *grpc.ClientConn }

func (w withGRPCConn) Apply(o *DialSettings) {
	o.GRPCConn = w.conn
}

// WithGRPCDialOption returns a ClientOption that appends a new grpc.DialOption
// to an underlying gRPC dial. It does not work with WithGRPCConn.
func WithGRPCDialOption(opt grpc.DialOption) ClientOption {
	return withGRPCDialOption{opt}
}

type withGRPCDialOption struct{ opt grpc.DialOption }

func (w withGRPCDialOption) Apply(o *DialSettings) {
	o.GRPCDialOpts = append(o.GRPCDialOpts, w.opt)
}

// WithGRPCConnectionPool returns a ClientOption that creates a pool of gRPC
// connections that requests will be balanced between.
//
// This is an EXPERIMENTAL API and may be changed or removed in the future.
func WithGRPCConnectionPool(size int) ClientOption {
	return withGRPCConnectionPool(size)
}

type withGRPCConnectionPool int

func (w withGRPCConnectionPool) Apply(o *DialSettings) {
	o.GRPCConnPoolSize = int(w)
}

// Validate reports an error if ds is invalid.
func (ds *DialSettings) Validate() error {
	if ds.SkipValidation {
		return nil
	}

	if ds.GRPCConnPool != nil {
		return errors.New("WithGRPCConn missing")
	}

	return nil
}

// WithEndpoint returns a ClientOption that defines the grpc service endpoint
func WithEndpoint(endpoint string) ClientOption {
	return withEndpoint(endpoint)
}

type withEndpoint string

func (w withEndpoint) Apply(o *DialSettings) {
	o.Endpoint = string(w)
}

// GetEndpoint returns the endpoint of a DialSettings.
func GetEndpoint(s *DialSettings) string {
	return s.Endpoint
}
