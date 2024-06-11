package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/jlucaspains/github-charts/models"
)

func (h Handlers) GetIterations(w http.ResponseWriter, r *http.Request) {
	iterations, err := h.Queries.GetIterations(r.Context())

	if err != nil {
		log.Printf("Error getting iteration data: %s", err)

		h.JSON(w, http.StatusInternalServerError, &models.ErrorResult{
			Errors: []string{"Could not get iteration data. Please try again later."},
		})
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
		log.Printf("Error getting burndown data: %s", err)
		h.JSON(w, http.StatusInternalServerError, &models.ErrorResult{
			Errors: []string{"Could not get burndown data. Please try again later."},
		})
	} else {
		h.JSON(w, http.StatusOK, burndown)
	}
}
