package models

import "context"

type StepParam interface {
	LoadParam(ctx context.Context) (string, error)
}
