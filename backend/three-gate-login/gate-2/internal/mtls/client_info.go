package mtls

import (
	"net/http"
	"os"
	"strings"
)

// ClientIdentity extracts the mTLS client Distinguished Name injected by Envoy.
//
// In production: Envoy sets the header (default: "x-client-dn") to the value of
// %DOWNSTREAM_PEER_SUBJECT% after verifying the client certificate.
//
// In mock/dev mode (GATE2_MOCK_CLIENT_DN=true): the header is accepted directly
// from the incoming request without Envoy — useful for smoke testing without mTLS.
//
// Returns the DN string and true on success, or ("", false) if the header is absent.
func ClientIdentity(r *http.Request, headerName string) (string, bool) {
	value := strings.TrimSpace(r.Header.Get(headerName))
	if value == "" {
		// In mock mode, fall back to a synthetic identity so tests work without Envoy.
		if os.Getenv("GATE2_MOCK_CLIENT_DN") == "true" || os.Getenv("GATE2_MOCK_CLIENT_DN") == "1" {
			return "CN=mock-client,O=SHIELD,C=IN", true
		}
		return "", false
	}
	return value, true
}
