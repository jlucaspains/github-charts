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

func TestGetProjects(t *testing.T) {
	expected := []db.Project{}
	expected = append(expected, db.Project{
		ID:   1,
		Name: "Project 1",
	})
	handlers := new(Handlers)
	handlers.Queries = &MockQuerier{GetProjectsResult: expected}

	router := http.NewServeMux()
	router.HandleFunc("GET /api/projects", handlers.GetProjects)

	code, body, _, err := makeRequest[[]*models.Project](router, "GET", "/api/projects", nil)

	assert.Nil(t, err)
	assert.Equal(t, 200, code)
	assert.Len(t, *body, 1)
	assert.Equal(t, expected[0].Name, (*body)[0].Title)
}

func TestGetProjectsError(t *testing.T) {
	handlers := new(Handlers)
	handlers.Queries = &MockQuerier{GetProjectsError: fmt.Errorf("error")}

	router := http.NewServeMux()
	router.HandleFunc("GET /api/projects", handlers.GetProjects)

	code, body, _, _ := makeRequest[models.ErrorResult](router, "GET", "/api/projects", nil)

	assert.Equal(t, "Unknown error", (*body).Errors[0])
	assert.Equal(t, 500, code)
}

func TestGetProjectBurnup(t *testing.T) {
	expected := []db.GetProjectBurnupRow{}
	expected = append(expected, db.GetProjectBurnupRow{
		Status:     "Done",
		ProjectDay: pgtype.Date{Time: time.Now(), Valid: true},
		Qty:        pgtype.Numeric{Int: big.NewInt(10), Valid: true},
	})
	expected = append(expected, db.GetProjectBurnupRow{
		Status:     "Done",
		ProjectDay: pgtype.Date{Time: time.Now().AddDate(0, 0, 1), Valid: true},
		Qty:        pgtype.Numeric{Int: big.NewInt(11), Valid: true},
	})
	expected = append(expected, db.GetProjectBurnupRow{
		Status:     "In Progress",
		ProjectDay: pgtype.Date{Time: time.Now(), Valid: true},
		Qty:        pgtype.Numeric{Int: big.NewInt(5), Valid: true},
	})
	expected = append(expected, db.GetProjectBurnupRow{
		Status:     "In Progress",
		ProjectDay: pgtype.Date{Time: time.Now().AddDate(0, 0, 1), Valid: true},
		Qty:        pgtype.Numeric{Int: big.NewInt(3), Valid: true},
	})
	handlers := new(Handlers)
	handlers.Queries = &MockQuerier{GetProjectBurnupResult: expected}

	router := http.NewServeMux()
	router.HandleFunc("GET /api/projects/{id}/burnup", handlers.GetBurnup)

	code, body, _, err := makeRequest[[]*models.BurnupItem](router, "GET", "/api/projects/1/burnup", nil)

	expectedQty, _ := expected[0].Qty.Float64Value()
	assert.Nil(t, err)
	assert.Equal(t, 200, code)
	assert.Len(t, *body, 4)
	assert.Equal(t, expected[0].ProjectDay.Time.Format("2006-01-02"), (*body)[0].ProjectDay.Format("2006-01-02"))
	assert.Equal(t, expectedQty.Float64, (*body)[0].Qty)
	assert.Equal(t, expected[0].Status, (*body)[0].Status)
}

func TestGetProjectBurnupError(t *testing.T) {
	handlers := new(Handlers)
	handlers.Queries = &MockQuerier{GetProjectBurnupError: fmt.Errorf("error")}

	router := http.NewServeMux()
	router.HandleFunc("GET /api/projects/{id}/burnup", handlers.GetBurnup)

	code, body, _, _ := makeRequest[models.ErrorResult](router, "GET", "/api/projects/1/burnup", nil)

	assert.Equal(t, "Unknown error", (*body).Errors[0])
	assert.Equal(t, 500, code)
}
