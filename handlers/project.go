package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jlucaspains/github-charts/db"
	"github.com/jlucaspains/github-charts/models"
)

func (h Handlers) GetProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.Queries.GetProjects(r.Context())

	if err != nil {
		log.Printf("Error getting iteration data: %s", err)

		status, body := h.ErrorToHttpResult(err)
		h.JSON(w, status, body)
	} else {
		result := []*models.Project{}
		for _, item := range projects {
			result = append(result, &models.Project{
				Id:    strconv.FormatInt(item.ID, 10),
				Title: item.Name,
			})
		}

		h.JSON(w, http.StatusOK, result)
	}
}

func (h Handlers) GetBurnup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	idInt64, _ := strconv.ParseInt(id, 10, 64)
	burnup, err := h.Queries.GetProjectBurnup(r.Context(), db.GetProjectBurnupParams{
		ProjectID: idInt64,
		Column2:   pgtype.Timestamp{Time: time.Now().AddDate(0, -1, 0), Valid: true},
	})

	if err != nil {
		log.Printf("Error getting burnup data: %s", err)
		status, body := h.ErrorToHttpResult(err)
		h.JSON(w, status, body)
	} else {
		result := []*models.BurnupItem{}
		for _, item := range burnup {
			qty, _ := item.Qty.Float64Value()
			result = append(result, &models.BurnupItem{
				ProjectDay: item.ProjectDay.Time,
				Qty:        qty.Float64,
				Status:     item.Status,
			})
		}

		h.JSON(w, http.StatusOK, result)
	}
}
