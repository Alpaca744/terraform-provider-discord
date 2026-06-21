package conns

import (
	"net/http"
	"testing"
)

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		status     int
		idempotent bool
		want       bool
	}{
		{http.StatusTooManyRequests, false, true}, // 429 always retried
		{http.StatusTooManyRequests, true, true},
		{http.StatusUnauthorized, true, false}, // invalid creds never retried
		{http.StatusForbidden, true, false},    // missing perms never retried
		{http.StatusInternalServerError, true, true},
		{http.StatusInternalServerError, false, false}, // unsafe POST not retried on 5xx
		{http.StatusBadGateway, true, true},
		{http.StatusBadRequest, true, false}, // client error not retried
		{http.StatusNotFound, true, false},
		{http.StatusOK, true, false},
	}
	for _, tt := range tests {
		if got := shouldRetry(tt.status, tt.idempotent); got != tt.want {
			t.Errorf("shouldRetry(%d, idempotent=%v) = %v, want %v",
				tt.status, tt.idempotent, got, tt.want)
		}
	}
}

func TestIsIdempotent(t *testing.T) {
	for _, m := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodHead} {
		if !isIdempotent(m) {
			t.Errorf("%s should be idempotent", m)
		}
	}
	for _, m := range []string{http.MethodPost, http.MethodPatch} {
		if isIdempotent(m) {
			t.Errorf("%s should not be idempotent", m)
		}
	}
}
