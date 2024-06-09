package handlers

import (
	"net/http"
	"strconv"
)

func (h Handlers) GetBurndown(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	idInt64, _ := strconv.ParseInt(id, 10, 64)
	burndown, _ := h.Queries.GetIterationBurndown(r.Context(), idInt64)
	h.JSON(w, http.StatusOK, burndown)
}
