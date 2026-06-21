package conns

import "net/http"

// shouldRetry decides whether a request that produced the given status should be
// retried. The rules implement the design doc's safety stance:
//
//   - 429 (rate limited) is always retryable after the indicated delay.
//   - 401/403 (invalid credentials / missing permission) are never retried;
//     retrying cannot succeed and only delays a clear diagnostic.
//   - 5xx server errors are retryable, since they are typically transient.
//   - Other 4xx are caller errors and are not retried.
//
// idempotent guards unsafe methods: only retry non-idempotent requests (POST)
// on conditions where Discord semantics make a retry safe (rate limits), never
// on 5xx where the write may have partially applied.
func shouldRetry(status int, idempotent bool) bool {
	switch {
	case status == http.StatusTooManyRequests:
		return true
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return false
	case status >= 500 && status <= 599:
		return idempotent
	default:
		return false
	}
}

// isIdempotent reports whether an HTTP method is safe to retry on transient
// server errors. PUT and DELETE are idempotent by HTTP semantics; GET is safe;
// POST is not.
func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}
