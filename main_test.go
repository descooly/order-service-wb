package main

import (
	"net/http"
	"net/http/httptest"
	"project/internal"
	my_cache "project/internal/cache"
	"testing"
)

func TestHTTPHandler(t *testing.T) {
	c := my_cache.New()
	c.Set(internal.OrderStruct{OrderUID: "TEST123"})
	service := OrderService{cache: c}

	req := httptest.NewRequest("GET", "/order?order_uid=TEST123", nil)
	rr := httptest.NewRecorder()

	service.HTTPHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got: %v", rr.Code)
	}

	req2 := httptest.NewRequest("GET", "/order?order_uid=MISSING", nil)
	rr2 := httptest.NewRecorder()

	service.HTTPHandler(rr2, req2)

	if rr2.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %v", rr2.Code)
	}

}
