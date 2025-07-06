package explorer

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

func (d *DbExplorer) validateTable(r *http.Request) error {
	var endpoint = r.URL.Path
	if endpoint != "/" {
		var split = strings.Split(endpoint, "/")

		var table = split[0]

		var tables, err = d.selectExistingTables()

		if err != nil {
			return err
		}
		if !slices.Contains(tables, table) {
			return apiError{
				ResponseCode: http.StatusNotFound,
				Err:          fmt.Errorf("unknown_table"),
			}
		}
	}
	return nil
}

func (d *DbExplorer) validateColumnTables(r *http.Request) error {
	var endpoint = r.URL.Path
	if endpoint != "/" && (r.Method == http.MethodPost || r.Method == http.MethodPut) {
		var split = strings.Split(endpoint, "/")

		var table = split[0]

		bodyColumns, err := d.parseBody(r)
		if err != nil {
			return err
		}

		columns, err := d.selectTableColumns(table)
		if err != nil {
			return err
		}

		for name, value := range bodyColumns {
			var column, exists = columns[name]
			if !exists {
				return apiError{
					ResponseCode: http.StatusBadRequest,
					Err:          fmt.Errorf("wrong column"),
				}
			}

		}
	}
	return nil
}
