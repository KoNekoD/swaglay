package controllers

import (
	"fiber/pkg/repositories"
	"fiber/pkg/services"
	swaglay "github.com/KoNekoD/swaglay/pkg"
	"github.com/KoNekoD/swaglay/pkg/adapters/swaglay_fiber"
	"github.com/gofiber/fiber/v3"
)

func InitAllControllers(r *fiber.App) {
	swaglay.SetupApi("Test Application")
	swaglay_fiber.Fiber = r
	swaglay_fiber.FiberApp = r

	var userRepository *repositories.UserRepository
	var userManager *services.UserManager

	// ...

	NewUserController(userRepository, userManager).Init(r)
}
