package swaglay_fiber

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"net/http"
)

var Fiber fiber.Router
var FiberApp *fiber.App

var NewResponseErrorBody = func(ctx fiber.Ctx, err error) any {
	return map[string]string{"error": err.Error()}
}

var NewResponseError = func(ctx fiber.Ctx, err error) (int, any) {
	status := http.StatusInternalServerError
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		status = http.StatusUnprocessableEntity
	}

	return status, NewResponseErrorBody(ctx, err)
}

var OnHandleError = func(err error) {}
