package db	

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Z interface {
	Querier
}

type Travia struct {
	db *pgxpool.Pool
	*Queries
}

func NewTravia(db *pgxpool.Pool) Z {
	return &Travia{db: db, Queries: New(db)}
}
