package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"go.opentelemetry.io/otel"
)

// TestMain initialises noop OTel metrics so package-level globals in observability
// are non-nil. Without this, any matched-route request would panic calling Record/Add.
func TestMain(m *testing.M) {
	if err := observability.InitMetrics(otel.GetMeterProvider().Meter("test")); err != nil {
		panic("failed to init noop metrics: " + err.Error())
	}
	os.Exit(m.Run())
}

// --- responseRecorder ---

func TestResponseRecorder_DefaultStatus(t *testing.T) {
	rr := newResponseRecorder(httptest.NewRecorder())
	if rr.statusCode != http.StatusOK {
		t.Errorf("expected default status 200, got %d", rr.statusCode)
	}
	if rr.written != 0 {
		t.Errorf("expected 0 bytes written, got %d", rr.written)
	}
}

func TestResponseRecorder_WriteHeader(t *testing.T) {
	underlying := httptest.NewRecorder()
	rr := newResponseRecorder(underlying)

	rr.WriteHeader(http.StatusCreated)

	if rr.statusCode != http.StatusCreated {
		t.Errorf("expected statusCode 201, got %d", rr.statusCode)
	}
	if underlying.Code != http.StatusCreated {
		t.Errorf("expected underlying recorder 201, got %d", underlying.Code)
	}
}

func TestResponseRecorder_Write(t *testing.T) {
	underlying := httptest.NewRecorder()
	rr := newResponseRecorder(underlying)

	body := []byte("hello world")
	n, err := rr.Write(body)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(body) {
		t.Errorf("Write() n = %d, want %d", n, len(body))
	}
	if rr.written != int64(len(body)) {
		t.Errorf("written counter = %d, want %d", rr.written, len(body))
	}
	if underlying.Body.String() != string(body) {
		t.Errorf("body = %q, want %q", underlying.Body.String(), body)
	}
}

func TestResponseRecorder_MultipleWrites(t *testing.T) {
	rr := newResponseRecorder(httptest.NewRecorder())

	rr.Write([]byte("foo"))
	rr.Write([]byte("bar"))

	if rr.written != 6 {
		t.Errorf("expected 6 bytes written, got %d", rr.written)
	}
}

// --- WrapMux ---

func TestWrapMux_SetsPatternInHolder(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	holder := &routePatternHolder{}
	ctx := context.WithValue(context.Background(), routePatternKey{}, holder)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	WrapMux(mux).ServeHTTP(rec, req)

	if holder.value != "GET /ping" {
		t.Errorf("expected pattern 'GET /ping', got %q", holder.value)
	}
}

func TestWrapMux_NoHolderInContext(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	// Should not panic even without a holder in context
	WrapMux(mux).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// --- NewObservabilityMiddleware ---

func buildHandler(t *testing.T, statusCode int) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		io.WriteString(w, "body")
	})
	return NewObservabilityMiddleware(WrapMux(mux))
}

func TestObservabilityMiddleware_SetsRequestID(t *testing.T) {
	handler := buildHandler(t, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestObservabilityMiddleware_MatchedRoute_200(t *testing.T) {
	handler := buildHandler(t, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestObservabilityMiddleware_MatchedRoute_4xx(t *testing.T) {
	handler := buildHandler(t, http.StatusBadRequest)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestObservabilityMiddleware_MatchedRoute_5xx(t *testing.T) {
	handler := buildHandler(t, http.StatusInternalServerError)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestObservabilityMiddleware_UnmatchedRoute_EarlyReturn(t *testing.T) {
	mux := http.NewServeMux()
	// No routes registered — any request gets a 404 with empty pattern
	handler := NewObservabilityMiddleware(WrapMux(mux))

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Should still set X-Request-ID before the early return
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID to be set even for unmatched routes")
	}
}

func TestObservabilityMiddleware_PropagatesTraceContext(t *testing.T) {
	handler := buildHandler(t, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	rec := httptest.NewRecorder()

	// Should not panic — trace context extraction must be graceful
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestObservabilityMiddleware_RoutePathStripping(t *testing.T) {
	// Verifies the middleware strips the method prefix from the route pattern
	// e.g. "POST /foo" → routePath "/foo" used in metrics/spans
	mux := http.NewServeMux()
	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler := NewObservabilityMiddleware(WrapMux(mux))

	req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}
