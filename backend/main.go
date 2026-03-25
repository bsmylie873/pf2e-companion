package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"pf2e-companion/backend/database"
	"pf2e-companion/backend/handlers"
	custmw "pf2e-companion/backend/middleware"
	"pf2e-companion/backend/repositories"
	"pf2e-companion/backend/services"
)

func main() {
	db := database.Connect()

	e := echo.New()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(custmw.CSRF())

	corsOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:5173"
	}
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAccept, "X-CSRF-Token"},
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
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)

	// Services
	authService := services.NewAuthService(userRepo, refreshTokenRepo)
	userService := services.NewUserService(userRepo, authService)
	gameService := services.NewGameService(gameRepo, membershipRepo)
	membershipService := services.NewMembershipService(membershipRepo)
	sessionService := services.NewSessionService(sessionRepo, membershipRepo)
	noteService := services.NewNoteService(noteRepo, membershipRepo)
	characterService := services.NewCharacterService(characterRepo, membershipRepo)
	itemService := services.NewItemService(itemRepo, membershipRepo, characterRepo)
	pinService := services.NewPinService(pinRepo, sessionRepo, membershipRepo)

	e.Static("/uploads", "./uploads")

	// Protected group — all resource routes require a valid JWT
	protected := e.Group("", custmw.RequireAuth(authService))

	// Auth routes (register, login, refresh are public; logout and me are protected)
	loginRateLimiter := custmw.RateLimiter()
	handlers.RegisterAuthRoutes(e, protected, authService, loginRateLimiter)

	// Protected resource routes
	handlers.RegisterUserRoutes(protected, userService)
	handlers.RegisterGameRoutes(protected, gameService)
	handlers.RegisterMembershipRoutes(protected, membershipService)
	handlers.RegisterSessionRoutes(protected, sessionService)
	handlers.RegisterNoteRoutes(protected, noteService)
	handlers.RegisterCharacterRoutes(protected, characterService)
	handlers.RegisterItemRoutes(protected, itemService)
	handlers.RegisterPinRoutes(protected, pinService)
	handlers.RegisterMapImageRoutes(protected, gameRepo, membershipRepo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := os.MkdirAll("./uploads/maps", 0755); err != nil {
		e.Logger.Fatal("failed to create uploads directory: ", err)
	}
	e.Logger.Fatal(e.Start(":" + port))
}
