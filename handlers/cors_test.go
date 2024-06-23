package handlers

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSCheckURL(t *testing.T) {
	handlers := new(Handlers)
	handlers.CORSOrigins = "http://localhost:5173"

	router := http.NewServeMux()
	router.HandleFunc("OPTIONS /api/iterations", handlers.CORS)

	code, body, headers, err := makeRequest[string](router, "OPTIONS", "/api/iterations", nil)

	assert.Nil(t, err)
	assert.Equal(t, 200, code)
	assert.Empty(t, body)
	assert.Equal(t, "http://localhost:5173", headers["Access-Control-Allow-Origin"][0])
}
