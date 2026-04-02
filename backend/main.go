package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"pf2e-companion/backend/database"
	"pf2e-companion/backend/handlers"
	custmw "pf2e-companion/backend/middleware"
	"pf2e-companion/backend/ot"
	"pf2e-companion/backend/repositories"
	"pf2e-companion/backend/services"
)

func main() {
	db := database.Connect()

	e := echo.New()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())

	corsOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:5173"
	}
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAccept},
		AllowCredentials: true,
	}))

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	gameRepo := repositories.NewGameRepository(db)
	membershipRepo := repositories.NewMembershipRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	noteRepo := repositories.NewNoteRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	itemRepo := repositories.NewItemRepository(db)
	pinRepo := repositories.NewPinRepository(db)
	pinGroupRepo := repositories.NewPinGroupRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	preferenceRepo := repositories.NewPreferenceRepository(db)
	folderRepo := repositories.NewFolderRepository(db)
	mapRepo := repositories.NewMapRepository(db)

	// Services
	authService := services.NewAuthService(userRepo, refreshTokenRepo)
	userService := services.NewUserService(userRepo, authService)
	gameService := services.NewGameService(gameRepo, membershipRepo)
	preferenceService := services.NewPreferenceService(preferenceRepo, membershipRepo)
	membershipService := services.NewMembershipService(membershipRepo, preferenceService)
	sessionService := services.NewSessionService(sessionRepo, membershipRepo)
	noteService := services.NewNoteService(noteRepo, membershipRepo, folderRepo)
	folderService := services.NewFolderService(folderRepo, membershipRepo)
	characterService := services.NewCharacterService(characterRepo, membershipRepo)
	itemService := services.NewItemService(itemRepo, membershipRepo, characterRepo)
	pinService := services.NewPinService(pinRepo, sessionRepo, membershipRepo, pinGroupRepo, mapRepo)
	pinGroupService := services.NewPinGroupService(pinGroupRepo, pinRepo, noteRepo, membershipRepo, mapRepo)
	mapService := services.NewMapService(mapRepo, membershipRepo)

	hub := handlers.NewGameEventHub()
	otStore := ot.NewDocumentStore()
	go ot.StartPersistenceLoop(otStore, func(entityID uuid.UUID, content json.RawMessage, version int) error {
		_, err := noteRepo.Update(entityID, map[string]interface{}{"content": content, "version": version})
		return err
	})

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.Static("/uploads", "./uploads")
	e.GET("/games/:id/ws", handlers.GameWebSocket(hub, otStore))

	// Protected group — all resource routes require a valid JWT
	protected := e.Group("", custmw.RequireAuth(authService))

	// Auth routes (register, login, refresh are public; logout and me are protected)
	loginRateLimiter := custmw.RateLimiter()
	handlers.RegisterAuthRoutes(e, protected, authService, loginRateLimiter)

	// Protected resource routes
	handlers.RegisterUserRoutes(protected, userService)
	handlers.RegisterGameRoutes(protected, gameService)
	handlers.RegisterMembershipRoutes(protected, membershipService)
	handlers.RegisterSessionRoutes(protected, sessionService, hub)
	handlers.RegisterNoteRoutes(protected, noteService, hub)
	handlers.RegisterFolderRoutes(protected, folderService)
	handlers.RegisterCharacterRoutes(protected, characterService)
	handlers.RegisterItemRoutes(protected, itemService)
	handlers.RegisterPinRoutes(protected, pinService, hub)
	handlers.RegisterPinGroupRoutes(protected, pinGroupService, hub)
	handlers.RegisterPreferenceRoutes(protected, preferenceService)
	handlers.RegisterMapRoutes(protected, mapService, hub)

	// Background cleanup: hard-delete maps archived more than 24 hours ago
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := mapService.CleanupExpiredMaps(); err != nil {
				e.Logger.Errorf("map cleanup error: %v", err)
			}
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := os.MkdirAll("./uploads/maps", 0755); err != nil {
		e.Logger.Fatal("failed to create uploads directory: ", err)
	}
	e.Logger.Fatal(e.Start(":" + port))
}
