package handlers

import (
	"net/http"

	"github.com/jlucaspains/github-charts/models"
)

func (h Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.JSON(w, http.StatusOK, &models.HealthResult{
		Healthy:      true,
		Dependencies: []models.HealthResultItem{},
	})
}
