package explorer

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"stepikGoWebServices/queries"
)

type tableColumn struct {
	Name            string
	DataType        string
	IsNullable      bool
	isPrimaryKey    bool
	isAutoIncrement bool
}

func (d *DbExplorer) getTableColumnsMap(table string) (map[string]tableColumn, error) {
	var rows, err = d.Db.Query(
		d.QueriesMap[queries.SelectTableColumnsQuery],
		table,
	)
	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusInternalServerError,
			Err:          fmt.Errorf("failed to select tables: %v", err),
		}
	}
	defer rows.Close()

	var tables = make(map[string]tableColumn)

	for rows.Next() {
		var name, column = parseRow(rows)
		if column == nil {
			continue
		}
		tables[name] = *column
	}

	return tables, nil
}

func (d *DbExplorer) getTableColumnsList(table string) ([]tableColumn, error) {
	var rows, err = d.Db.Query(
		d.QueriesMap[queries.SelectTableColumnsQuery],
		table,
	)
	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusInternalServerError,
			Err:          fmt.Errorf("failed to select tables: %v", err),
		}
	}
	defer rows.Close()

	var columns = make([]tableColumn, 0)

	for rows.Next() {
		var _, column = parseRow(rows)
		if column == nil {
			continue
		}
		columns = append(
			columns,
			*column,
		)
	}

	return columns, nil
}

func parseRow(rows *sql.Rows) (string, *tableColumn) {
	var columnName string
	var dataTypeRaw string
	var isNullableRaw string
	var columnKey string
	var extra string

	var err = rows.Scan(
		&columnName,
		&dataTypeRaw,
		&isNullableRaw,
		&columnKey,
		&extra,
	)
	if err != nil {
		log.Printf("Failed to scan tables row: %v\n", err)
	}

	var dataType string
	switch dataTypeRaw {
	case "varchar", "text":
		dataType = "string"
	default:
		dataType = dataTypeRaw
	}

	return columnName, &tableColumn{
		Name:            columnName,
		DataType:        dataType,
		IsNullable:      isNullableRaw == "YES",
		isPrimaryKey:    columnKey == "PRI",
		isAutoIncrement: extra == "auto_increment",
	}
}

func findPrimaryKeyColumn(columns map[string]tableColumn) string {
	var primaryKey string
	for _, value := range columns {
		if value.isPrimaryKey {
			primaryKey = value.Name
			break
		}
	}
	return primaryKey
}
