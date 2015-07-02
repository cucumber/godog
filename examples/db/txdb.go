package main

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
)

// Register a txdb sql driver which can be used to open
// a single transaction based database connection pool
func Register(drv, dsn string) {
	sql.Register("txdb", &txDriver{dsn: dsn, drv: drv})
}

// txDriver is an sql driver which runs on single transaction
// when the Close is called, transaction is rolled back
type txDriver struct {
	sync.Mutex
	tx *sql.Tx

	drv string
	dsn string
	db  *sql.DB
}

func (d *txDriver) Open(dsn string) (driver.Conn, error) {
	// first open a real database connection
	var err error
	if d.db == nil {
		db, err := sql.Open(d.drv, d.dsn)
		if err != nil {
			return d, err
		}
		d.db = db
	}
	if d.tx == nil {
		d.tx, err = d.db.Begin()
	}
	return d, err
}

func (d *txDriver) Close() error {
	err := d.tx.Rollback()
	d.tx = nil
	return err
}

func (d *txDriver) Begin() (driver.Tx, error) {
	return d, nil
}

func (d *txDriver) Commit() error {
	return nil
}

func (d *txDriver) Rollback() error {
	return nil
}

func (d *txDriver) Prepare(query string) (driver.Stmt, error) {
	return &stmt{drv: d, query: query}, nil
}

type stmt struct {
	query string
	drv   *txDriver
}

func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	s.drv.Lock()
	defer s.drv.Unlock()

	st, err := s.drv.tx.Prepare(s.query)
	if err != nil {
		return nil, err
	}
	defer st.Close()
	var iargs []interface{}
	for _, arg := range args {
		iargs = append(iargs, arg)
	}
	return st.Exec(iargs...)
}

func (s *stmt) NumInput() int {
	return -1
}

func (s *stmt) Close() error {
	return nil
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	s.drv.Lock()
	defer s.drv.Unlock()

	st, err := s.drv.tx.Prepare(s.query)
	if err != nil {
		return nil, err
	}
	// do not close the statement here, Rows need it
	var iargs []interface{}
	for _, arg := range args {
		iargs = append(iargs, arg)
	}
	rs, err := st.Query(iargs...)
	return &rows{rs: rs}, err
}

type rows struct {
	err error
	rs  *sql.Rows
}

func (r *rows) Columns() (cols []string) {
	cols, r.err = r.rs.Columns()
	return
}

func (r *rows) Next(dest []driver.Value) error {
	if r.err != nil {
		return r.err
	}
	if r.rs.Err() != nil {
		return r.rs.Err()
	}
	if !r.rs.Next() {
		return io.EOF
	}
	values := make([]interface{}, len(dest))
	for i := range values {
		values[i] = new(interface{})
	}
	if err := r.rs.Scan(values...); err != nil {
		return err
	}
	for i, val := range values {
		dest[i] = *(val.(*interface{}))
	}
	return r.rs.Err()
}

func (r *rows) Close() error {
	return r.rs.Close()
}
