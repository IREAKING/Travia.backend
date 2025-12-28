package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Z interface {
	Querier

	// Transaction methods for complex operations
	CreateTourWithDetails(ctx context.Context, params CreateTourWithDetailsParams) (*CreateTourWithDetailsResult, error)
	UpdateTourWithDetails(ctx context.Context, tourID int32, params CreateTourWithDetailsParams) (*CreateTourWithDetailsResult, error)
	CreateSupplierWithUser(ctx context.Context, req CreateSupplierWithUserParams) (*CreateSupplierWithUserResult, error)
}

type Travia struct {
	db *pgxpool.Pool
	*Queries
}

func NewTravia(db *pgxpool.Pool) Z {
	return &Travia{db: db, Queries: New(db)}
}
