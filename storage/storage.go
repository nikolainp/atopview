package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3" //sqlite3
)

// Storage ...
type Storage struct {
	metadata metaData

	cacheInsertValueSQL map[string]string

	db *sql.DB
}

// CreateCache - Storage contructor
func CreateCache() (*Storage, error) {

	var err error

	obj := newStorage()
	if obj.db, err = openDB(""); err != nil {
		return nil, err
	}

	initDB(obj.db, obj.metadata)

	return obj, nil
}

// Open ...
func Open(stroragePath string) (obj *Storage, err error) {

	if _, err := os.Stat(stroragePath); err != nil {
		return nil, fmt.Errorf("open storage: %v", err)
	}

	obj = newStorage()
	if obj.db, err = openDB(stroragePath); err != nil {
		return nil, err
	}

	return obj, nil
}

// FlushAll ...
func (obj *Storage) FlushAll(stroragePath string) error {

	if err := os.Remove(stroragePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear storage: %v", err)
	}

	db, err := openDB(stroragePath)
	if err != nil {
		return err
	}
	initDB(db, obj.metadata)
	db.Close()

	obj.saveAll(stroragePath)

	return nil

}

// WriteRow ...
func (obj *Storage) WriteRow(table string, args ...any) {
	query, ok := obj.cacheInsertValueSQL[table]
	if !ok {
		query = obj.metadata.GetInsertValueSQL(table)
		obj.cacheInsertValueSQL[table] = query
	}

	if _, err := obj.db.Exec(query, args...); err != nil {
		panic(fmt.Errorf("\nquery: %s\nerror: %w", query, err))
	}
}


// func (obj *Storage) SetIdByGroup(table string, column, group string) {
// 	query := obj.metadata.SetIdByGroup(table, column, group)

// 	if _, err := obj.db.Exec(query); err != nil {
// 		panic(fmt.Errorf("\nquery: %s\nerror: %w", query, err))
// 	}

// }

///////////////////////////////////////////////////////////////////////////////

func newStorage() *Storage {
	obj := new(Storage)
	obj.metadata = newMetadata()
	obj.cacheInsertValueSQL = make(map[string]string)

	return obj
}

func openDB(stroragePath string) (*sql.DB, error) {
	var dataSource string
	var err error

	if stroragePath == "" {
		// dataSource = ":memory:?mode=memory&cache=private&nolock=1&psow=1"
		dataSource = ":memory:?_journal_mode=WAL&cache=shared"
	} else {
		//		dataSource = "file:" + stroragePath + "?cache=private&nolock=1&psow=1"
		dataSource = "file:" + stroragePath + "?_journal_mode=WAL&cache=shared"
	}
	db, err := sql.Open("sqlite3", dataSource)
	if err != nil {
		return nil, fmt.Errorf("open storage: %v", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping storage: %v", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	queries := []string{
		`PRAGMA main.journal_mode = MEMORY`,
	}
	for i := range queries {
		if _, err := db.Exec(queries[i]); err != nil {
			panic(err)
		}
	}

	return db, nil
}

func initDB(db *sql.DB, meta metaData) {
	for _, table := range meta.InitDB() {
		if _, err := db.Exec(table); err != nil {
			panic(err)
		}
	}
}

func (obj *Storage) saveAll(dbPath string) {
	if _, err := obj.db.Exec("ATTACH DATABASE '" + dbPath + "' AS datafile"); err != nil {
		panic(err)
	}

	parts := obj.metadata.SaveAll("datafile")
	for _, part := range parts {
		if _, err := obj.db.Exec(part); err != nil {
			panic(fmt.Errorf("query: %s\nerror: %w", part, err))
		}
	}

	if _, err := obj.db.Exec("DETACH datafile"); err != nil {
		panic(err)
	}
}
