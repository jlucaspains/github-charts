package handlers

import (
	"net/http"
)

func (h Handlers) GetBurndown(w http.ResponseWriter, r *http.Request) {
	burndown, _ := h.Queries.GetIterationBurndown(r.Context())
	h.JSON(w, http.StatusOK, burndown)
}
