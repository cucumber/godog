package godog

import (
	"fmt"

	"github.com/cucumber/messages-go/v10"
	"github.com/hashicorp/go-memdb"
)

const (
	writeMode bool = true
	readMode  bool = false

	tablePickle         string = "pickle"
	tablePickleIndexID  string = "id"
	tablePickleIndexURI string = "uri"

	tablePickleStep        string = "pickle_step"
	tablePickleStepIndexID string = "id"

	tablePickleResult              string = "pickle_result"
	tablePickleResultIndexPickleID string = "id"

	tablePickleStepResult                  string = "pickle_step_result"
	tablePickleStepResultIndexPickleStepID string = "id"
	tablePickleStepResultIndexPickleID     string = "pickle_id"
	tablePickleStepResultIndexStatus       string = "status"
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
					tablePickleIndexURI: {
						Name:    tablePickleIndexURI,
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Uri"},
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
			tablePickleResult: {
				Name: tablePickleResult,
				Indexes: map[string]*memdb.IndexSchema{
					tablePickleResultIndexPickleID: {
						Name:    tablePickleResultIndexPickleID,
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "PickleID"},
					},
				},
			},
			tablePickleStepResult: {
				Name: tablePickleStepResult,
				Indexes: map[string]*memdb.IndexSchema{
					tablePickleStepResultIndexPickleStepID: {
						Name:    tablePickleStepResultIndexPickleStepID,
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "PickleStepID"},
					},
					tablePickleStepResultIndexPickleID: {
						Name:    tablePickleStepResultIndexPickleID,
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "PickleID"},
					},
					tablePickleStepResultIndexStatus: {
						Name:    tablePickleStepResultIndexStatus,
						Unique:  false,
						Indexer: &memdb.IntFieldIndex{Field: "Status"},
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

func (s *storage) mustInsertPickle(p *messages.Pickle) {
	txn := s.db.Txn(writeMode)

	if err := txn.Insert(tablePickle, p); err != nil {
		panic(err)
	}

	for _, step := range p.Steps {
		if err := txn.Insert(tablePickleStep, step); err != nil {
			panic(err)
		}
	}

	txn.Commit()
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

func (s *storage) mustGetPickles(uri string) (ps []*messages.Pickle) {
	txn := s.db.Txn(readMode)
	defer txn.Abort()

	it, err := txn.Get(tablePickle, tablePickleIndexURI, uri)
	if err != nil {
		panic(err)
	}

	for v := it.Next(); v != nil; v = it.Next() {
		ps = append(ps, v.(*messages.Pickle))
	}

	return
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

func (s *storage) mustInsertPickleResult(pr pickleResult) {
	txn := s.db.Txn(writeMode)

	if err := txn.Insert(tablePickleResult, pr); err != nil {
		panic(err)
	}

	txn.Commit()
}

func (s *storage) mustInsertPickleStepResult(psr pickleStepResult) {
	txn := s.db.Txn(writeMode)

	if err := txn.Insert(tablePickleStepResult, psr); err != nil {
		panic(err)
	}

	txn.Commit()
}

func (s *storage) mustGetPickleResult(id string) pickleResult {
	pr, err := s.getPickleResult(id)
	if err != nil {
		panic(err)
	}

	return pr
}

func (s *storage) getPickleResult(id string) (_ pickleResult, err error) {
	txn := s.db.Txn(readMode)
	defer txn.Abort()

	v, err := txn.First(tablePickleResult, tablePickleResultIndexPickleID, id)
	if err != nil {
		return
	} else if v == nil {
		err = fmt.Errorf("Couldn't find pickle result with ID: %s", id)
		return
	}

	return v.(pickleResult), nil
}

func (s *storage) mustGetPickleResults() (prs []pickleResult) {
	txn := s.db.Txn(readMode)
	defer txn.Abort()

	it, err := txn.Get(tablePickleResult, tablePickleResultIndexPickleID)
	if err != nil {
		panic(err)
	}

	for v := it.Next(); v != nil; v = it.Next() {
		prs = append(prs, v.(pickleResult))
	}

	return prs
}

func (s *storage) mustGetPickleStepResult(id string) pickleStepResult {
	txn := s.db.Txn(readMode)
	defer txn.Abort()

	v, err := txn.First(tablePickleStepResult, tablePickleStepResultIndexPickleStepID, id)
	if err != nil {
		panic(err)
	} else if v == nil {
		panic("Couldn't find pickle step result with ID: " + id)
	}

	return v.(pickleStepResult)
}

func (s *storage) mustGetPickleStepResultsByPickleID(pickleID string) (psrs []pickleStepResult) {
	txn := s.db.Txn(readMode)
	defer txn.Abort()

	it, err := txn.Get(tablePickleStepResult, tablePickleStepResultIndexPickleID, pickleID)
	if err != nil {
		panic(err)
	}

	for v := it.Next(); v != nil; v = it.Next() {
		psrs = append(psrs, v.(pickleStepResult))
	}

	return psrs
}

func (s *storage) mustGetPickleStepResultsByStatus(status stepResultStatus) (psrs []pickleStepResult) {
	txn := s.db.Txn(readMode)
	defer txn.Abort()

	it, err := txn.Get(tablePickleStepResult, tablePickleStepResultIndexStatus, status)
	if err != nil {
		panic(err)
	}

	for v := it.Next(); v != nil; v = it.Next() {
		psrs = append(psrs, v.(pickleStepResult))
	}

	return psrs
}
