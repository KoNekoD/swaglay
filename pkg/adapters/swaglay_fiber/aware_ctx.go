package swaglay_fiber

import "github.com/gofiber/fiber/v3"

type AwareCtx interface {
	GetCtx() fiber.Ctx
	SetCtx(ctx fiber.Ctx)
}

type AwareCtxStruct struct {
	ctx fiber.Ctx
}

func (a *AwareCtxStruct) GetCtx() fiber.Ctx {
	return a.ctx
}

func (a *AwareCtxStruct) SetCtx(ctx fiber.Ctx) {
	a.ctx = ctx
}
