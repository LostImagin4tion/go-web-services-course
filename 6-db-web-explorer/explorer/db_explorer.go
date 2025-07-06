package explorer

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

type DbExplorer struct {
	http.Handler

	Db         *sql.DB
	QueriesMap map[string]string
}

func NewDbExplorer(db *sql.DB, queriesMap map[string]string) (DbExplorer, error) {
	return DbExplorer{
		Db:         db,
		QueriesMap: queriesMap,
	}, nil
}

func (d *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares = [...]func(*http.Request) error{
		d.validateTable,
		d.validateColumnTables,
	}

	for _, middleware := range middlewares {
		var err = middleware(r)
		if err == nil {
			continue
		}

		var apiErr = apiError{}
		var ok = errors.As(err, &apiErr)
		if ok {
			http.Error(w, apiErr.Error(), apiErr.ResponseCode)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	var (
		err    error
		output any
	)

	if r.URL.Path == "/" {
		output, err = d.handleRootPath(r)
	} else {
		output, err = d.handleArbitraryPath(r)
	}

	if err != nil {
		var apiErr = apiError{}
		var ok = errors.As(err, &apiErr)
		if ok {
			http.Error(w, apiErr.Error(), apiErr.ResponseCode)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		marshaledOutput, err := json.Marshal(output)
		if err != nil {
			http.Error(w, "failed to marshal output", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(marshaledOutput)
	}
}
