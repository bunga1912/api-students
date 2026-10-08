package helper

import (
	"net/http"
	"strconv"
)

func QueryString(
	r *http.Request,
	key string,
	defaultValue string,
) string {

	value := r.URL.Query().Get(key)

	if value == "" {
		return defaultValue
	}

	return value
}

func QueryInt(
	r *http.Request,
	key string,
	defaultValue int,
) int {

	value := r.URL.Query().Get(key)

	if value == "" {
		return defaultValue
	}

	result, err := strconv.Atoi(value)

	if err != nil {
		return defaultValue
	}

	return result
}
