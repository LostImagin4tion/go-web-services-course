package explorer

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
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
			return nil, apiError{
				ResponseCode: http.StatusBadRequest,
				Err:          fmt.Errorf("failed to convert id to int: %v", err),
			}
		}
	}
	if id == -1 && (r.Method == http.MethodPost || r.Method == http.MethodDelete) {
		return nil, apiError{
			ResponseCode: http.StatusBadRequest,
			Err:          fmt.Errorf("id has invalid type"),
		}
	}

	switch r.Method {
	case http.MethodGet:
		return d.handleGetRequest(r, table, id)

	case http.MethodPut:
		return d.handlePutRequest(r, table)

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
) (any, error) {

	var body, err = d.parseBody(r)
	var columns = make([]string, len(body))
	var values = make([]string, len(body))

	for column := range body {
		columns = append(columns, column)
	}
	for _, value := range body {
		values = append(values, fmt.Sprintf("%v", value))
	}

	var columnsArg = fmt.Sprintf("'%s'", strings.Join(columns, ","))
	var valuesArg = fmt.Sprintf("'%s'", strings.Join(values, ","))

	result, err := d.Db.Exec(
		d.QueriesMap[queries.CreateNewItemQuery],
		table,
		columnsArg,
		valuesArg,
	)

	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusBadRequest,
			Err:          fmt.Errorf("failed to add new item: %v", err),
		}
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusInternalServerError,
			Err:          fmt.Errorf("failed to get last inserted id: %v", err),
		}
	}

	return map[string]int{
		columns[0]: int(lastId),
	}, nil
}

func (d *DbExplorer) handlePostRequest(
	r *http.Request,
	table string,
	id int,
) (any, error) {

	var body, err = d.parseBody(r)

	var args = make([]any, len(body)+2)
	args = append(args, table)

	for name, value := range body {
		args = append(
			args,
			fmt.Sprintf("`%s` = %v", name, value),
		)
	}

	args = append(args, id)

	result, err := d.Db.Exec(
		d.QueriesMap[queries.UpdateItemQuery],
		args...,
	)

	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusBadRequest,
			Err:          fmt.Errorf("failed to update item: %v", err),
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusInternalServerError,
			Err:          fmt.Errorf("failed to get affected rows: %v", err),
		}
	}

	return map[string]int{
		"updated": int(rowsAffected),
	}, nil
}

func (d *DbExplorer) handleDeleteRequest(
	r *http.Request,
	table string,
	id int,
) (any, error) {

}
