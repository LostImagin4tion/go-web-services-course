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
		} else {
			handleApiError(w, err)
			return
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
		handleApiError(w, err)
	} else {
		var response = map[string]any{
			"response": output,
		}
		marshaledResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "failed to marshal output", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(marshaledResponse)
	}
}

func handleApiError(w http.ResponseWriter, err error) {
	var apiErr = apiError{}
	var ok = errors.As(err, &apiErr)

	var jsonError = map[string]string{
		"error": apiErr.Error(),
	}

	marshalledJson, err := json.Marshal(jsonError)
	if err != nil {
		http.Error(w, "failed to marshal output", http.StatusInternalServerError)
		return
	}

	if ok {
		w.WriteHeader(apiErr.ResponseCode)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(marshalledJson)
}
