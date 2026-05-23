package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxJSONBodyBytes int64 = 1 << 20

func decodeJSONBody(w http.ResponseWriter, req *http.Request, dst any) bool {
	req.Body = http.MaxBytesReader(w, req.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, jsonDecodeErrorMessage(err))
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func jsonDecodeErrorMessage(err error) string {
	var syntaxError *json.SyntaxError
	var typeError *json.UnmarshalTypeError
	var maxBytesError *http.MaxBytesError
	switch {
	case errors.As(err, &maxBytesError):
		return "JSON body is too large"
	case errors.As(err, &syntaxError):
		return "invalid JSON body"
	case errors.As(err, &typeError):
		return "invalid JSON field type"
	case errors.Is(err, io.ErrUnexpectedEOF):
		return "invalid JSON body"
	case errors.Is(err, io.EOF):
		return "JSON body is required"
	default:
		return "invalid JSON body"
	}
}
