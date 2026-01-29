package handlers

import (
	"connect-four-backend/services"
	"encoding/json"
	"net/http"
)

func HandleLeaderboard(w http.ResponseWriter, r *http.Request, gameService *services.GameService) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	leaderboard, err := gameService.GetLeaderboard()
	if err != nil {
		http.Error(w, "Failed to get leaderboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}
