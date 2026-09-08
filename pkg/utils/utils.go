// Package utils provides small cross-domain helpers for gRPC transports.
package utils

import (
	"context"
	"errors"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToStatusError maps a domain error to the corresponding gRPC status.
//
// All domains share a single "not found" sentinel (errs.ErrNotFound), so a
// no-argument mapping is enough for any of them.
func ToStatusError(err error) error {
	switch {
	case errors.Is(err, errs.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
