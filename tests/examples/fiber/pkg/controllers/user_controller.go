package controllers

import (
	"fiber/pkg/dtos"
	"fiber/pkg/repositories"
	"fiber/pkg/services"
	. "github.com/KoNekoD/swaglay/pkg/adapters/swaglay_fiber"
	"github.com/gofiber/fiber/v3"
	"strconv"
)

type UserController struct {
	userRepository *repositories.UserRepository
	userManager    *services.UserManager
}

func NewUserController(
	userRepository *repositories.UserRepository,
	userManager *services.UserManager,
) *UserController {
	return &UserController{
		userRepository: userRepository,
		userManager:    userManager,
	}
}

func (c *UserController) Init(r *fiber.App) {
	const api = "users"
	Fiber = r.Group("/api/users") // Add too `.Use(c.authMiddleware.Exec)`
	GetO(api, "/me", c.Me, "Me")
	PostI(api, "/change-email", c.ChangeEmail, "Change email")
	Delete(api, "/delete-account", c.DeleteAccount, "Delete account")
	Delete(api, "/{id}", c.DeleteAccount, "Delete account")
}

func (c *UserController) Me(ctx fiber.Ctx) (*dtos.UserDto, error) {
	userId := 123

	return c.userRepository.GetById(userId)
}

func (c *UserController) ChangeEmail(i *dtos.ChangeEmail, ctx fiber.Ctx) error {
	userId := 123

	return c.userManager.ChangeEmail(*i.Email, userId)
}

func (c *UserController) DeleteAccount(ctx fiber.Ctx) error {
	userIdStr := ctx.Params("id")
	userId, _ := strconv.Atoi(userIdStr)

	return c.userManager.DeleteAccount(userId)
}
