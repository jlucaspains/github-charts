package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jlucaspains/github-charts/models"
)

func (h Handlers) GetIterations(w http.ResponseWriter, r *http.Request) {
	iterations, err := h.Queries.GetIterations(r.Context())

	if err != nil {
		slog.Error("Error getting iteration data: %s", err)

		status, body := h.ErrorToHttpResult(err)
		h.JSON(w, status, body)
	} else {
		result := []*models.Iteration{}
		for _, item := range iterations {
			result = append(result, &models.Iteration{
				Id:        strconv.FormatInt(item.ID, 10),
				Title:     item.Name,
				StartDate: item.StartDate.Time,
				EndDate:   item.EndDate.Time,
			})
		}

		h.JSON(w, http.StatusOK, result)
	}
}

func (h Handlers) GetBurndown(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	idInt64, _ := strconv.ParseInt(id, 10, 64)
	burndown, err := h.Queries.GetIterationBurndown(r.Context(), idInt64)

	if err != nil {
		slog.Error("Error getting burndown data: %s", err)
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
