package explorer

import (
	"fmt"
	"log"
	"net/http"
	"stepikGoWebServices/queries"
)

type tableColumn struct {
	Name       string
	DataType   string
	IsNullable bool
}

func (d *DbExplorer) selectTableColumns(table string) (map[string]tableColumn, error) {
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
		var columnName string
		var dataTypeRaw string
		var isNullableRaw string

		err = rows.Scan(&columnName, &dataTypeRaw, &isNullableRaw)
		if err != nil {
			log.Println("Failed to scan tables row")
			continue
		}

		var dataType string
		switch dataTypeRaw {
		case "varchar", "text":
			dataType = "string"
		default:
			dataType = dataTypeRaw
		}

		tables[columnName] = tableColumn{
			Name:       columnName,
			DataType:   dataType,
			IsNullable: isNullableRaw == "YES",
		}
	}

	return tables, nil
}
