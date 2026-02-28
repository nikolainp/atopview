package storage

import (
	"fmt"
	"strings"
)

type updateQuery struct {
	query      string
	queryWhere []string
	args       []any

	isExecuted bool

	data *Storage
}

// Update ...
func (obj *Storage) Update(table string, args ...any) interface {
	SetFilter(filter ...string)
	Execute()
} {

	result := new(updateQuery)

	result.data = obj
	result.args = make([]any, 0, len(args))

	fields := make([]any, 0, len(args))
	for i := range args {
		if i%2 == 0 {
			fields = append(fields, args[i])
		} else {
			result.args = append(result.args, args[i])
		}
	}

	result.query = obj.metadata.GetUpdateSQL(table, fields...)

	return result
}

// /////////////////////////////////////////////////////////////////////////////

func (obj *updateQuery) SetFilter(filter ...string) {
	if len(filter) != 0 {
		obj.queryWhere = append(obj.queryWhere, filter[0])
	}
	if len(filter) == 2 {
		obj.args = append(obj.args, filter[1])
	}
}

func (obj *updateQuery) Execute() {
	var err error

	makeQuery := func() string {
		query := obj.query
		if len(obj.queryWhere) != 0 {
			query = query + " WHERE (" + strings.Join(obj.queryWhere, ") AND (") + ")"
		}

		return query
	}

	if obj.isExecuted {
		panic(fmt.Errorf("\nreExecute query: %s\nerror: %w", obj.query, err))
	}

	obj.query = makeQuery()
	_, err = obj.data.db.Exec(obj.query, obj.args...)
	if err != nil {
		panic(fmt.Errorf("\nquery: %s\nerror: %w", obj.query, err))
	}
	obj.isExecuted = true
}
