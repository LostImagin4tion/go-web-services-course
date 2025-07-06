package explorer

import (
	"fmt"
	"net/http"
)

func (d *DbExplorer) handleRootPath(r *http.Request) (any, error) {
	if r.Method != "GET" {
		return nil, apiError{
			ResponseCode: http.StatusMethodNotAllowed,
			Err:          fmt.Errorf("wrong method"),
		}
	}

	return d.selectExistingTables()
}
