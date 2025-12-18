package ports

import "context"

type AuthRepo interface {
	GetPassword(ctx context.Context) (string, error)
}
