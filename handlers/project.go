package handlers

import (
	"log/slog"
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
		slog.Error("Error getting iteration data", "error", err)

		status, body := h.ErrorToHttpResult(err)
		h.JSON(w, status, body)
	} else {
		result := []*models.Project{}
		for _, item := range projects {
			result = append(result, &models.Project{
				Id:    strconv.Itoa(int(item.ID)),
				Title: item.Name,
			})
		}

		h.JSON(w, http.StatusOK, result)
	}
}

func (h Handlers) GetBurnup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("projectId")
	idInt, _ := strconv.Atoi(id)
	burnup, err := h.Queries.GetProjectBurnup(r.Context(), db.GetProjectBurnupParams{
		ProjectID: int32(idInt),
		Column2:   pgtype.Timestamp{Time: time.Now().AddDate(0, -1, 0), Valid: true},
	})

	if err != nil {
		slog.Error("Error getting burnup data", "error", err)
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

func (h Handlers) GetIterations(w http.ResponseWriter, r *http.Request) {
	projectId := r.PathValue("projectId")
	projectIdInt, _ := strconv.Atoi(projectId)
	iterations, err := h.Queries.GetIterations(r.Context(), int32(projectIdInt))

	if err != nil {
		slog.Error("Error getting iteration data: %s", err)

		status, body := h.ErrorToHttpResult(err)
		h.JSON(w, status, body)
	} else {
		result := []*models.Iteration{}
		for _, item := range iterations {
			result = append(result, &models.Iteration{
				Id:        strconv.FormatInt(int64(item.ID), 10),
				Title:     item.Name,
				StartDate: item.StartDate.Time,
				EndDate:   item.EndDate.Time,
			})
		}

		h.JSON(w, http.StatusOK, result)
	}
}

func (h Handlers) GetBurndown(w http.ResponseWriter, r *http.Request) {
	iterationId := r.PathValue("iterationId")
	iterationIdInt, _ := strconv.Atoi(iterationId)
	burndown, err := h.Queries.GetIterationBurndown(r.Context(), int32(iterationIdInt))

	if err != nil {
		slog.Error("Error getting burndown data", "error", err)
		status, body := h.ErrorToHttpResult(err)
		h.JSON(w, status, body)
	} else {
		result := []*models.BurndownItem{}
		for _, item := range burndown {
			remaining, _ := item.Remaining.Float64Value()
			ideal, _ := item.Ideal.Float64Value()
			result = append(result, &models.BurndownItem{
				IterationDay: item.IterationDay.Time,
				Remaining:    remaining.Float64,
				Ideal:        ideal.Float64,
			})
		}

		h.JSON(w, http.StatusOK, result)
	}
}
