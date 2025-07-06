package explorer

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"slices"
	"stepikGoWebServices/queries"
	"strconv"
	"strings"
)

func (d *DbExplorer) handleArbitraryPath(r *http.Request) (any, error) {
	var apiPath = r.URL.Path
	var split = strings.Split(apiPath, "/")

	var table = split[0]

	var id = -1
	var err error

	if len(split) > 1 {
		id, err = strconv.Atoi(split[1])
		if err != nil {
			return nil, fmt.Errorf("failed to convert id to int: %v", err)
		}
	}

	switch r.Method {
	case http.MethodGet:
		return d.handleGetRequest(r, table, id)

	case http.MethodPut:
		return d.handlePutRequest(r, table, id)

	case http.MethodPost:
		return d.handlePostRequest(r, table, id)

	case http.MethodDelete:
		return d.handleDeleteRequest(r, table, id)

	default:
		return nil, apiError{
			ResponseCode: http.StatusBadRequest,
			Err:          fmt.Errorf("wrong method"),
		}
	}
}

func (d *DbExplorer) handleGetRequest(
	r *http.Request,
	table string,
	id int,
) (any, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if id != -1 {
		rows, err = d.Db.Query(
			d.QueriesMap[queries.SelectFromTableQuery],
			table,
			id,
		)
	} else {
		var queryParams = d.parseQueryParams(r)
		var limit = queryParams["limit"]
		var offset = queryParams["offset"]

		rows, err = d.Db.Query(
			d.QueriesMap[queries.SelectAllFromTableQuery],
			table,
			limit,
			offset,
		)
	}

	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusBadRequest,
			Err:          fmt.Errorf("failed to select from table: %v", err),
		}
	}
	defer rows.Close()

	var items = make([]any, 0)

	for rows.Next() {
		var item any
		err = rows.Scan(item)
		if err != nil {
			log.Println("Failed to scan tables row")
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (d *DbExplorer) handlePutRequest(
	r *http.Request,
	table string,
	id int,
) (any, error) {

}

func (d *DbExplorer) handlePostRequest(
	r *http.Request,
	table string,
	id int,
) (any, error) {

}

func (d *DbExplorer) handleDeleteRequest(
	r *http.Request,
	table string,
	id int,
) (any, error) {

}
