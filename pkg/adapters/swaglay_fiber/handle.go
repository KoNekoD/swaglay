package swaglay_fiber

import (
	"github.com/gofiber/fiber/v3"
)

type HandleFnIO[In any, Out any] func(i *In, ctx fiber.Ctx) (Out, error)
type HandleFnI[In any] func(i *In, ctx fiber.Ctx) error
type HandleFnO[Out any] func(ctx fiber.Ctx) (Out, error)
type HandleFn func(ctx fiber.Ctx) error

func checkErr() {

}

func handleIO[In any, Out any](i *In, ctx fiber.Ctx, fn HandleFnIO[In, Out]) {
	output, err := fn(i, ctx)
	if err != nil {
		OnHandleError(err)

		status, data := NewResponseError(ctx, err)

		err = ctx.Status(status).JSON(data)
		if err != nil {
			OnHandleError(err)
		}

		return
	}

	if err = ctx.JSON(output); err != nil {
		OnHandleError(err)
	}
}

func handleI[In any](i *In, ctx fiber.Ctx, fn HandleFnI[In]) {
	err := fn(i, ctx)
	if err != nil {
		OnHandleError(err)

		status, data := NewResponseError(ctx, err)

		err = ctx.Status(status).JSON(data)
		if err != nil {
			OnHandleError(err)
		}

		return
	}
}

func handleO[Out any](ctx fiber.Ctx, fn HandleFnO[Out]) {
	output, err := fn(ctx)
	if err != nil {
		OnHandleError(err)

		status, data := NewResponseError(ctx, err)

		err = ctx.Status(status).JSON(data)
		if err != nil {
			OnHandleError(err)
		}

		return
	}

	if err = ctx.JSON(output); err != nil {
		OnHandleError(err)
	}
}

func handle(ctx fiber.Ctx, fn HandleFn) {
	err := fn(ctx)
	if err != nil {
		OnHandleError(err)

		status, data := NewResponseError(ctx, err)

		err = ctx.Status(status).JSON(data)
		if err != nil {
			OnHandleError(err)
		}

		return
	}
}
