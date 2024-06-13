package handlers

import (
	"fmt"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jlucaspains/github-charts/db"
	"github.com/jlucaspains/github-charts/models"
	"github.com/stretchr/testify/assert"
)

func TestGetIterations(t *testing.T) {
	expected := []db.Iteration{}
	expected = append(expected, db.Iteration{
		ID:        1,
		Name:      "Iteration 1",
		StartDate: pgtype.Date{Time: time.Now(), Valid: true},
		EndDate:   pgtype.Date{Time: time.Now().AddDate(0, 0, 15), Valid: true},
	})
	handlers := new(Handlers)
	handlers.Queries = &MockQuerier{GetIterationsResult: expected}

	router := http.NewServeMux()
	router.HandleFunc("GET /api/iterations", handlers.GetIterations)

	code, body, _, err := makeRequest[[]*models.Iteration](router, "GET", "/api/iterations", nil)

	assert.Nil(t, err)
	assert.Equal(t, 200, code)
	assert.Len(t, *body, 1)
	assert.Equal(t, expected[0].Name, (*body)[0].Title)
	assert.Equal(t, expected[0].StartDate.Time.Format("2006-01-02"), (*body)[0].StartDate.Format("2006-01-02"))
	assert.Equal(t, expected[0].EndDate.Time.Format("2006-01-02"), (*body)[0].EndDate.Format("2006-01-02"))
}

func TestGetIterationsError(t *testing.T) {
	handlers := new(Handlers)
	handlers.Queries = &MockQuerier{GetIterationsResult: nil, GetIterationsError: fmt.Errorf("error")}

	router := http.NewServeMux()
	router.HandleFunc("GET /api/iterations", handlers.GetIterations)

	code, body, _, _ := makeRequest[models.ErrorResult](router, "GET", "/api/iterations", nil)

	assert.Equal(t, "Unknown error", (*body).Errors[0])
	assert.Equal(t, 500, code)
}

func TestGetBurndown(t *testing.T) {
	expected := []db.GetIterationBurndownRow{}
	expected = append(expected, db.GetIterationBurndownRow{
		IterationDay: pgtype.Date{Time: time.Now(), Valid: true},
		Remaining:    pgtype.Numeric{Int: big.NewInt(10), Valid: true},
		Ideal:        pgtype.Numeric{Int: big.NewInt(10), Valid: true},
	})
	expected = append(expected, db.GetIterationBurndownRow{
		IterationDay: pgtype.Date{Time: time.Now().AddDate(0, 0, 1), Valid: true},
		Remaining:    pgtype.Numeric{Int: big.NewInt(9), Valid: true},
		Ideal:        pgtype.Numeric{Int: big.NewInt(8), Valid: true},
	})
	handlers := new(Handlers)
	handlers.Queries = &MockQuerier{GetIterationBurndownResult: expected}

	router := http.NewServeMux()
	router.HandleFunc("GET /api/iterations/1/burndown", handlers.GetBurndown)

	code, body, _, err := makeRequest[[]*models.BurndownItem](router, "GET", "/api/iterations/1/burndown", nil)

	ideal, _ := expected[0].Ideal.Float64Value()
	remaining, _ := expected[0].Remaining.Float64Value()
	assert.Nil(t, err)
	assert.Equal(t, 200, code)
	assert.Len(t, *body, 2)
	assert.Equal(t, ideal.Float64, (*body)[0].Ideal)
	assert.Equal(t, expected[0].IterationDay.Time.Format("2006-01-02"), (*body)[0].IterationDay.Format("2006-01-02"))
	assert.Equal(t, remaining.Float64, (*body)[0].Remaining)
}

func TestGetBurndownError(t *testing.T) {
	handlers := new(Handlers)
	handlers.Queries = &MockQuerier{GetIterationBurndownError: fmt.Errorf("error")}

	router := http.NewServeMux()
	router.HandleFunc("GET /api/iterations/1/burndown", handlers.GetBurndown)

	code, body, _, _ := makeRequest[models.ErrorResult](router, "GET", "/api/iterations/1/burndown", nil)

	assert.Equal(t, "Unknown error", (*body).Errors[0])
	assert.Equal(t, 500, code)
}
