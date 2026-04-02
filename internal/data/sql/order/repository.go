package order

import (
	coreDB "github.com/aryanwalia/synapse/internal/core/db"
	"github.com/aryanwalia/synapse/internal/core/domain"
)

type SQLOrderRepository struct {
	db coreDB.DB
}

func NewSQLOrderRepository(db coreDB.DB) domain.OrderRepository {
	return &SQLOrderRepository{db: db}
}
