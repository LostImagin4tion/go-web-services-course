package explorer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

func (d *DbExplorer) validateTable(r *http.Request) error {
	var endpoint = r.URL.Path
	if endpoint != "/" {
		var split = strings.Split(endpoint, "/")

		var table = split[1]

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

		var table = split[1]

		bodyColumns, err := d.parseBody(r)
		if err != nil {
			return err
		}

		columns, err := d.getTableColumnsMap(table)
		if err != nil {
			return err
		}

		for name, value := range bodyColumns {
			var column, exists = columns[name]
			if !exists {
				continue
			}

			if column.isPrimaryKey && r.Method == http.MethodPost {
				return apiError{
					ResponseCode: http.StatusBadRequest,
					Err:          fmt.Errorf("field %v has invalid type", column.Name),
				}
			}

			if number, isNumber := value.(json.Number); isNumber {
				if _, err = number.Int64(); (err == nil && column.DataType == "int") || column.DataType == "float" {
					return nil
				}
			}
			if value == nil && column.IsNullable {
				return nil
			}
			if fmt.Sprintf("%T", value) != column.DataType {
				return apiError{
					ResponseCode: http.StatusBadRequest,
					Err:          fmt.Errorf("field %v has invalid type", name),
				}
			}
		}
	}
	return nil
}
