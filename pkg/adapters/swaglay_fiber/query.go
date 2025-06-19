package swaglay_fiber

import (
	"github.com/KoNekoD/go-querymap/pkg/querymap"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"net/http"
	"net/url"
)

func satisfyQuery[DtoType any](ctx fiber.Ctx) *DtoType {
	values := make(url.Values, ctx.RequestCtx().QueryArgs().Len())
	ctx.RequestCtx().QueryArgs().VisitAll(
		func(key, value []byte) {
			keyString := utils.UnsafeString(key)
			valueString := utils.UnsafeString(value)

			if _, ok := values[keyString]; ok {
				values[keyString] = append(values[keyString], valueString)
			} else {
				values[keyString] = []string{valueString}
			}
		},
	)

	dto, err := querymap.FromValuesToStruct[DtoType](values)
	if err != nil {
		err = ctx.Status(http.StatusBadRequest).JSON(NewResponseErrorBody(ctx, err))
		if err != nil {
			OnHandleError(err)
		}

		return nil
	}

	if err = FiberApp.Config().StructValidator.Validate(&dto); err != nil {
		err = ctx.Status(http.StatusUnprocessableEntity).JSON(NewResponseErrorBody(ctx, err))
		if err != nil {
			OnHandleError(err)
		}

		return nil
	}

	return dto
}
