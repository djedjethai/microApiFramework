package commonRest

import (
	"encoding/json"
	"errors"
	e "gitlab.com/grpasr/common/errors/json"
	"io"
	"net/http"
)

func readJson(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("Body must have only a single json value")
	}

	return nil
}

func writeJson(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	// w.Header().Set("Access-Control-Allow-Origin", "*")
	_, err = w.Write(out)
	if err != nil {
		return nil
	}

	return nil
}

func CustomResponseJson(w http.ResponseWriter, ce e.IError) error {

	w.Header().Set("Content-Type", "application/json")

	// automatically handled
	// w.WriteHeader(ce.GetCode())

	json.NewEncoder(w).Encode(ce)
	return nil
}
