package storage

import (
	"fmt"
	"strings"
)

type metaData interface {
	InitDB(isCache bool) []string
	SaveAll(schema string) []string
	PostLoad() []string
	CalcPivot() []metaPivot

	GetInsertValueSQL(table string) string
	GetUpdateSQL(table string, fields ...any) string
	//	SetIdByGroup(table string, column, group string) string

	SelectColumnsSQL(table string, columns ...string) string
	GetFilterSQL(table string) string
}

type implMetaData struct {
	tables map[string]metaTable
}

type metaTable struct {
	name    string
	columns []metaColumn
	indexes []string

	//insertStm *sql.Stmt
	isCache  bool
	postLoad []string
	pivot    metaPivot

	columnTimeFrom, columnTimeTo string
}

type metaColumn struct {
	name      string
	datatype  string
	value     string
	isNotNull bool

	isService bool

	isTimeFrom bool
	isTimeTo   bool
}

type metaPivot struct {
	columns string
	create  string
	calc    string
}

///////////////////////////////////////////////////////////////////////////////

// newMetadata ...
func newMetadata() metaData {
	obj := new(implMetaData)
	obj.init()
	return obj
}

func (obj *implMetaData) InitDB(isCache bool) []string {
	queries := make([]string, 0, len(obj.tables))

	for _, table := range obj.tables {
		if table.isCache && !isCache {
			continue
		}

		queries = append(queries, table.getCreateSQL())
		if table.indexes != nil {
			queries = append(queries, table.indexes...)
		}
	}

	return queries
}

func (obj *implMetaData) SaveAll(schema string) []string {
	queries := make([]string, 0, len(obj.tables))

	for _, table := range obj.tables {
		if table.isCache {
			continue
		}
		queries = append(queries, table.getInsertSelectSQL(schema))
	}

	return queries
}

func (obj *implMetaData) PostLoad() []string {
	queries := make([]string, 0, len(obj.tables))

	for _, table := range obj.tables {
		if table.postLoad != nil {
			queries = append(queries, table.postLoad...)
		}
	}

	return queries
}

func (obj *implMetaData) CalcPivot() []metaPivot {
	queries := make([]metaPivot, 0, len(obj.tables))

	for _, table := range obj.tables {
		if table.pivot != (metaPivot{}) {
			queries = append(queries, table.pivot)
		}
	}

	return queries
}

func (obj *implMetaData) GetInsertValueSQL(table string) string {
	return obj.tables[table].getInsertValueSQL()
}

func (obj *implMetaData) SelectColumnsSQL(table string, columns ...string) (query string) {

	var queryColumns string

	if len(columns) == 0 {
		names := make([]string, 0, 10)
		metaColumns := obj.tables[table].columns
		for i := range metaColumns {
			names = append(names, metaColumns[i].name)
		}
		queryColumns = strings.Join(names, ", ")
	} else {
		queryColumns = strings.Join(columns, ", ")
	}

	query = fmt.Sprintf("SELECT DISTINCT %s FROM %s", queryColumns, table)

	return
}

func (obj *implMetaData) GetFilterSQL(table string) (filter string) {

	metaTable := obj.tables[table]
	return fmt.Sprintf("(%[1]s = '0001-01-01 00:00:00+00:00' or %[1]s >= ?) and %[2]s <= ?", metaTable.columnTimeTo, metaTable.columnTimeFrom)
}

func (obj *implMetaData) GetUpdateSQL(table string, fields ...any) (query string) {

	set := make([]string, 0, len(fields))
	for i := range fields {
		set = append(set, fmt.Sprintf("%s = ?", fields[i]))
	}

	query = fmt.Sprintf("UPDATE %s SET %s", table, strings.Join(set, ", "))

	return
}

///////////////////////////////////////////////////////////////////////////////

func (obj *implMetaData) init() {
	obj.tables = getDataStructure()

	for tableName, table := range obj.tables {
		for _, column := range table.columns {
			if column.isTimeFrom {
				table.columnTimeFrom = column.name
			}
			if column.isTimeTo {
				table.columnTimeTo = column.name
			}
		}
		obj.tables[tableName] = table
	}

	//	return
}

func (obj *metaTable) getCreateSQL() string {

	queryColumns := make([]string, 0, len(obj.columns))
	for i := range obj.columns {
		meta := obj.columns[i]

		column := fmt.Sprintf("%s %s", meta.name, meta.datatype)
		if meta.value != "" {
			column += " DEFAULT " + meta.value
		}
		if meta.isNotNull {
			column += " NOT NULL"
		}

		queryColumns = append(queryColumns, column)
	}

	return fmt.Sprintf("CREATE TABLE %s (%s)", obj.name, strings.Join(queryColumns, ","))
}

func (obj metaTable) getInsertValueSQL() string {
	queryColumns := make([]string, 0, len(obj.columns))
	queryValues := make([]string, 0, len(obj.columns))

	for i := range obj.columns {
		if obj.columns[i].isService {
			continue
		}
		queryColumns = append(queryColumns, obj.columns[i].name)
		queryValues = append(queryValues, "?")
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", obj.name,
		strings.Join(queryColumns, ","),
		strings.Join(queryValues, ","),
	)
}

func (obj *metaTable) getInsertSelectSQL(schema string) string {
	queryColumns := make([]string, 0, len(obj.columns))

	for i := range obj.columns {
		queryColumns = append(queryColumns, obj.columns[i].name)
	}

	return fmt.Sprintf("INSERT INTO %s.%s (%s) SELECT %s FROM %s",
		schema,
		obj.name, strings.Join(queryColumns, ","),
		strings.Join(queryColumns, ","), obj.name,
	)
}
