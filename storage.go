package godog

import (
	"github.com/cucumber/messages-go/v10"
	"github.com/hashicorp/go-memdb"
)

const (
	writeMode bool = true
	readMode  bool = false

	tablePickle        string = "pickle"
	tablePickleIndexID string = "id"

	tablePickleStep        string = "pickle_step"
	tablePickleStepIndexID string = "id"
)

type storage struct {
	db *memdb.MemDB
}

func newStorage() *storage {
	// Create the DB schema
	schema := memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			tablePickle: {
				Name: tablePickle,
				Indexes: map[string]*memdb.IndexSchema{
					tablePickleIndexID: {
						Name:    tablePickleIndexID,
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Id"},
					},
				},
			},
			tablePickleStep: {
				Name: tablePickleStep,
				Indexes: map[string]*memdb.IndexSchema{
					tablePickleStepIndexID: {
						Name:    tablePickleStepIndexID,
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Id"},
					},
				},
			},
		},
	}

	// Create a new data base
	db, err := memdb.NewMemDB(&schema)
	if err != nil {
		panic(err)
	}

	return &storage{db}
}

func (s *storage) mustInsertPickle(p *messages.Pickle) (err error) {
	txn := s.db.Txn(writeMode)

	if err = txn.Insert(tablePickle, p); err != nil {
		panic(err)
	}

	for _, step := range p.Steps {
		if err = txn.Insert(tablePickleStep, step); err != nil {
			panic(err)
		}
	}

	txn.Commit()
	return
}

func (s *storage) mustGetPickle(id string) *messages.Pickle {
	txn := s.db.Txn(readMode)
	defer txn.Abort()

	var v interface{}
	v, err := txn.First(tablePickle, tablePickleIndexID, id)
	if err != nil {
		panic(err)
	} else if v == nil {
		panic("Couldn't find pickle with ID: " + id)
	}

	return v.(*messages.Pickle)
}

func (s *storage) mustGetPickleStep(id string) *messages.Pickle_PickleStep {
	txn := s.db.Txn(readMode)
	defer txn.Abort()

	var v interface{}
	v, err := txn.First(tablePickleStep, tablePickleStepIndexID, id)
	if err != nil {
		panic(err)
	} else if v == nil {
		panic("Couldn't find pickle step with ID: " + id)
	}

	return v.(*messages.Pickle_PickleStep)
}
