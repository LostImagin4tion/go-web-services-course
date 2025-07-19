package explorer

import (
	"fmt"
	"log"
	"net/http"
	"stepikGoWebServices/queries"
)

func (d *DbExplorer) selectExistingTables() ([]string, error) {
	var rows, err = d.Db.Query(d.QueriesMap[queries.SelectTablesQuery])
	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusInternalServerError,
			Err:          fmt.Errorf("failed to select tables: %v", err),
		}
	}
	defer rows.Close()

	var tables = make([]string, 0)

	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			log.Println("Failed to scan tables row")
			continue
		}
		tables = append(tables, table)
	}

	return tables, nil
}
