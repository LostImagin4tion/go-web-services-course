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

	var table = split[1]

	var id = -1
	var err error

	if len(split) > 2 && len(split[2]) > 0 {
		id, err = strconv.Atoi(split[2])
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
			Err:          fmt.Errorf("id is not specified or has invalid type"),
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
		return d.handleDeleteRequest(table, id)

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
		var tableColumns map[string]tableColumn
		tableColumns, err = d.getTableColumnsMap(table)
		if err != nil {
			return nil, err
		}

		var primaryKeyColumn = findPrimaryKeyColumn(tableColumns)

		var formattedQuery = fmt.Sprintf(
			d.QueriesMap[queries.SelectFromTableQuery],
			table,
			fmt.Sprintf("`%s` = %v", primaryKeyColumn, id),
		)
		rows, err = d.Db.Query(formattedQuery)
	} else {
		var queryParams = d.parseQueryParams(r)
		var limit = queryParams["limit"]
		var offset = queryParams["offset"]

		var formattedQuery = fmt.Sprintf(
			d.QueriesMap[queries.SelectAllFromTableQuery],
			table,
		)

		rows, err = d.Db.Query(
			formattedQuery,
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

	var items = make([]map[string]any, 0)

	for rows.Next() {
		var columns, _ = rows.Columns()
		var itemFields = make([]any, len(columns))
		var itemFieldsReference = make([]any, len(columns))

		for i := range itemFields {
			itemFieldsReference[i] = &itemFields[i]
		}

		err = rows.Scan(itemFieldsReference...)
		if err != nil {
			log.Println("Failed to scan tables row, ", err)
			continue
		}

		var itemMap = make(map[string]any)
		for i, name := range columns {
			value := itemFields[i]

			if bytes, ok := value.([]byte); ok {
				itemMap[name] = string(bytes)
			} else {
				itemMap[name] = value
			}
		}

		items = append(items, itemMap)
	}

	if len(items) == 0 {
		return nil, apiError{
			ResponseCode: http.StatusNotFound,
			Err:          fmt.Errorf("record not found"),
		}
	}

	return map[string]any{
		"records": items,
	}, nil
}

func (d *DbExplorer) handlePutRequest(
	r *http.Request,
	table string,
) (any, error) {

	body, err := d.parseBody(r)
	if err != nil {
		return nil, err
	}

	tableColumns, err := d.getTableColumnsMap(table)
	if err != nil {
		return nil, err
	}

	var primaryKeyColumn = findPrimaryKeyColumn(tableColumns)
	var columns = make([]string, 0)
	var values = make([]string, 0)

	for columnName, value := range body {
		if _, exists := tableColumns[columnName]; !exists || columnName == primaryKeyColumn {
			continue
		}

		columns = append(columns, columnName)

		values = append(
			values,
			fmt.Sprintf("%v", parseBodyValue(value)),
		)
	}

	var formattedQuery = fmt.Sprintf(
		d.QueriesMap[queries.CreateNewItemQuery],
		table,
		strings.Join(columns, ", "),
		strings.Join(values, ", "),
	)

	result, err := d.Db.Exec(formattedQuery)

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
		primaryKeyColumn: int(lastId),
	}, nil
}

func (d *DbExplorer) handlePostRequest(
	r *http.Request,
	table string,
	id int,
) (any, error) {

	body, err := d.parseBody(r)
	if err != nil {
		return nil, err
	}

	columns, err := d.getTableColumnsMap(table)
	if err != nil {
		return nil, err
	}

	var primaryKeyColumn = findPrimaryKeyColumn(columns)

	var setArgs = make([]string, 0)

	for name, value := range body {
		setArgs = append(
			setArgs,
			fmt.Sprintf("`%s` = %v", name, parseBodyValue(value)),
		)
	}

	var formattedQuery = fmt.Sprintf(
		d.QueriesMap[queries.UpdateItemQuery],
		table,
		strings.Join(setArgs, ", "),
		fmt.Sprintf("%s = %v", primaryKeyColumn, id),
	)

	result, err := d.Db.Exec(formattedQuery)

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
	table string,
	id int,
) (any, error) {

	columns, err := d.getTableColumnsMap(table)
	if err != nil {
		return nil, err
	}

	var primaryKeyColumn = findPrimaryKeyColumn(columns)

	var formattedQuery = fmt.Sprintf(
		d.QueriesMap[queries.DeleteItemQuery],
		table,
		fmt.Sprintf("`%s` = %v", primaryKeyColumn, id),
	)

	result, err := d.Db.Exec(formattedQuery)

	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusBadRequest,
			Err:          fmt.Errorf("failed to delete item: %v", err),
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
		"deleted": int(rowsAffected),
	}, nil
}

func parseBodyValue(value any) any {
	var fmtValue any
	if strValue, ok := value.(string); ok {
		fmtValue = fmt.Sprintf(
			"'%v'",
			strings.ReplaceAll(strValue, "'", "''"),
		)
	} else if value == nil {
		fmtValue = "null"
	} else {
		fmtValue = value
	}
	return fmtValue
}
