package explorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (d *DbExplorer) parseQueryParams(r *http.Request) map[string]int {
	var query = r.URL.Query()

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		limit = 5
	}

	offset, err := strconv.Atoi(query.Get("offset"))
	if err != nil {
		offset = 0
	}

	return map[string]int{
		"limit":  limit,
		"offset": offset,
	}
}

func (d *DbExplorer) parseBody(r *http.Request) (map[string]any, error) {
	var body, err = io.ReadAll(r.Body)
	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusBadRequest,
			Err:          fmt.Errorf("failed to read request body"),
		}
	}

	var decoder = json.NewDecoder(bytes.NewBuffer(body))
	decoder.UseNumber()

	var bodyDump interface{}
	err = decoder.Decode(&bodyDump)
	if err != nil {
		return nil, apiError{
			ResponseCode: http.StatusInternalServerError,
			Err:          fmt.Errorf("failed to unmarshal body: %v", err),
		}
	}

	var bodyMap, ok = bodyDump.(map[string]any)
	if !ok {
		return nil, apiError{
			ResponseCode: http.StatusBadRequest,
			Err:          fmt.Errorf("body contains not json"),
		}
	}

	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return bodyMap, nil
}
