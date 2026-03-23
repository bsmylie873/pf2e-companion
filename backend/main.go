package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"pf2e-companion/backend/database"
	"pf2e-companion/backend/handlers"
	"pf2e-companion/backend/repositories"
	"pf2e-companion/backend/services"
)

func main() {
	db := database.Connect()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	corsOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{corsOrigin},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Users
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	handlers.RegisterUserRoutes(e, userService)

	// Games
	gameRepo := repositories.NewGameRepository(db)
	membershipRepo := repositories.NewMembershipRepository(db)
	gameService := services.NewGameService(gameRepo, userRepo, membershipRepo)
	handlers.RegisterGameRoutes(e, gameService)

	// Memberships
	membershipService := services.NewMembershipService(membershipRepo)
	handlers.RegisterMembershipRoutes(e, membershipService)

	// Sessions
	sessionRepo := repositories.NewSessionRepository(db)
	sessionService := services.NewSessionService(sessionRepo)
	handlers.RegisterSessionRoutes(e, sessionService)

	// Notes
	noteRepo := repositories.NewNoteRepository(db)
	noteService := services.NewNoteService(noteRepo)
	handlers.RegisterNoteRoutes(e, noteService)

	// Characters
	characterRepo := repositories.NewCharacterRepository(db)
	characterService := services.NewCharacterService(characterRepo)
	handlers.RegisterCharacterRoutes(e, characterService)

	// Items
	itemRepo := repositories.NewItemRepository(db)
	itemService := services.NewItemService(itemRepo)
	handlers.RegisterItemRoutes(e, itemService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
