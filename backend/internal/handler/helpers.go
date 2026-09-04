package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// decodeJSON decodes a JSON body into dst, rejecting empty bodies.
func decodeJSON(r *http.Request, dst any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MiB max
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if len(body) == 0 {
		return errEmptyBody
	}
	return json.Unmarshal(body, dst)
}

var errEmptyBody = &decodeErr{"empty request body"}

type decodeErr struct{ msg string }

func (e *decodeErr) Error() string { return e.msg }

// queryInt parses an integer query parameter, returning the default if absent/invalid.
func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// queryFloat parses a float query parameter, returning def if absent/invalid.
func queryFloat(r *http.Request, key string, def float64) float64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

// queryCSV splits a comma-separated query parameter into trimmed non-empty parts.
func queryCSV(r *http.Request, key string) []string {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	out := []string{}
	for _, p := range strings.Split(v, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// queryBool parses a boolean query parameter (true/1/yes).
func queryBool(r *http.Request, key string) bool {
	v := strings.ToLower(r.URL.Query().Get(key))
	return v == "true" || v == "1" || v == "yes"
}
