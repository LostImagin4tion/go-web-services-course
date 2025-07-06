package queries

import (
	"fmt"
	"io"
	"os"
)

const (
	queriesDir = "./queries/"

	CreateSampleDbQuery     = "sample_db.sql"
	SelectTablesQuery       = "select_tables.sql"
	SelectTableColumnsQuery = "select_table_columns.sql"
	SelectFromTableQuery    = "select_from_table.sql"
	SelectAllFromTableQuery = "select_all_from_table.sql"
	CreateNewItemQuery      = "create_new_item.sql"
	UpdateItemQuery         = "update_item.sql"
	DeleteItemQuery         = "delete_item.sql"
)

var (
	queries = []string{
		CreateSampleDbQuery,
		SelectTablesQuery,
		SelectTableColumnsQuery,
		SelectFromTableQuery,
		SelectAllFromTableQuery,
		CreateNewItemQuery,
		UpdateItemQuery,
		DeleteItemQuery,
	}
)

func NewQueriesMap() map[string]string {
	var queriesMap = make(map[string]string)

	for _, query := range queries {
		file, err := os.Open(fmt.Sprintf("%s%s", queriesDir, query))
		if err != nil {
			panic(err)
		}

		fileContent, err := io.ReadAll(file)
		if err != nil {
			panic(err)
		}

		queriesMap[query] = string(fileContent)
	}

	return queriesMap
}
