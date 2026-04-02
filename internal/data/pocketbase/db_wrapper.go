package pocketbase

import (
	"context"

	coreDB "github.com/aryanwalia/synapse/internal/core/db"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// PBDatabase implements core/db.DB by wrapping PocketBase.
type PBDatabase struct {
	app core.App
}

func NewPBDatabase(app core.App) coreDB.DB {
	return &PBDatabase{app: app}
}

func (db *PBDatabase) Execute(ctx context.Context, query string, params map[string]any) error {
	_, err := db.app.DB().NewQuery(query).Bind(dbx.Params(params)).Execute()
	return err
} //Runs update , insert , Delete but does not return anything

func (db *PBDatabase) QueryRow(ctx context.Context, query string, params map[string]any, dest any) error {
	return db.app.DB().NewQuery(query).Bind(dbx.Params(params)).One(dest)
} //Returns single row

func (db *PBDatabase) QueryRows(ctx context.Context, query string, params map[string]any, dest any) error {
	return db.app.DB().NewQuery(query).Bind(dbx.Params(params)).All(dest)
} //Returns multiple rows

func (db *PBDatabase) RunInTransaction(ctx context.Context, fn func(tx coreDB.Transaction) error) error {
	return db.app.RunInTransaction(func(pbTx core.App) error {
		txWrapper := &PBTransaction{pbTx: pbTx}
		return fn(txWrapper)
	})
}

// PBTransaction implements core/db.Transaction
type PBTransaction struct {
	pbTx core.App
}

func (tx *PBTransaction) Execute(ctx context.Context, query string, params map[string]any) error {
	_, err := tx.pbTx.DB().NewQuery(query).Bind(dbx.Params(params)).Execute()
	return err
}

func (tx *PBTransaction) QueryRow(ctx context.Context, query string, params map[string]any, dest any) error {
	return tx.pbTx.DB().NewQuery(query).Bind(dbx.Params(params)).One(dest)
}

func (tx *PBTransaction) QueryRows(ctx context.Context, query string, params map[string]any, dest any) error {
	return tx.pbTx.DB().NewQuery(query).Bind(dbx.Params(params)).All(dest)
}
