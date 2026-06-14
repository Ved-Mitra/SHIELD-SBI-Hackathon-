package mtls

import (
	"net/http"
	"strings"
)

func ClientIdentity(r *http.Request, headerName string) (string, bool) {
	value := strings.TrimSpace(r.Header.Get(headerName))
	if value == "" {
		return "", false
	}
	return value, true
}

