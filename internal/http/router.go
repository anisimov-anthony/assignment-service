package http

import (
	"net/http"

	"assignment-service/internal/http/handlers"
	"assignment-service/internal/repository/mongodb"
	"assignment-service/internal/service"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func SetupRouter(client *mongodb.Client, logger *zap.Logger) http.Handler {
	// Repos
	userRepo := mongodb.NewUserRepository(client, logger)
	teamRepo := mongodb.NewTeamRepository(client, logger)
	prRepo := mongodb.NewPRRepository(client, logger)

	// Services
	teamService := service.NewTeamService(teamRepo, userRepo, logger)
	userService := service.NewUserService(userRepo, logger)
	prService := service.NewPRService(prRepo, userRepo, logger)
	statsService := service.NewStatsService(prRepo, userRepo, logger)

	// Handlers
	teamHandler := handlers.NewTeamHandler(teamService, logger)
	userHandler := handlers.NewUserHandler(userService, prService, logger)
	prHandler := handlers.NewPRHandler(prService, logger)
	statsHandler := handlers.NewStatsHandler(statsService, logger)
	healthHandler := handlers.NewHealthHandler()

	// Router setup
	router := mux.NewRouter()

	// - Teams
	router.HandleFunc("/team/add", teamHandler.CreateTeam).Methods(http.MethodPost)
	router.HandleFunc("/team/get", teamHandler.GetTeam).Methods(http.MethodGet)

	// - Users
	router.HandleFunc("/users/setIsActive", userHandler.SetIsActive).Methods(http.MethodPost)
	router.HandleFunc("/users/getReview", userHandler.GetReview).Methods(http.MethodGet)

	// - PullRequests
	router.HandleFunc("/pullRequest/create", prHandler.CreatePR).Methods(http.MethodPost)
	router.HandleFunc("/pullRequest/merge", prHandler.MergePR).Methods(http.MethodPost)
	router.HandleFunc("/pullRequest/reassign", prHandler.ReassignReviewer).Methods(http.MethodPost)

	// - Health
	router.HandleFunc("/health", healthHandler.Health).Methods(http.MethodGet)

	// - Stats
	router.HandleFunc("/stats/user", statsHandler.GetUserStats).Methods(http.MethodGet)

	return router
}
