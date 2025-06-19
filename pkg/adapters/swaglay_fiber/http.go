package swaglay_fiber

import (
	"github.com/KoNekoD/swaglay/pkg"
	"github.com/gofiber/fiber/v3"
	"net/http"
)

type Opts struct {
	Out          any
	Use          fiber.Handler
	Uses         []fiber.Handler
	UseWithInput bool
}

func wrapBodyInputMiddleware[In any](opts []Opts) []Opts {
	uses := make([]fiber.Handler, len(opts[0].Uses)+1)
	uses[0] = func(ctx fiber.Ctx) error {
		var input In
		setCtxIfNeeded(&input, ctx)

		if err := ctx.Bind().JSON(&input); err != nil {
			return ctx.Status(http.StatusUnprocessableEntity).JSON(NewResponseErrorBody(ctx, err))
		}
		ctx.Locals("input", &input)
		return ctx.Next()
	}

	for i, use := range opts[0].Uses {
		uses[i+1] = use
	}

	opts[0].Uses = uses
	return opts
}

func wrapSatisfyQueryInputMiddleware[In any](opts []Opts) []Opts {
	uses := make([]fiber.Handler, len(opts[0].Uses)+1)

	uses[0] = func(ctx fiber.Ctx) error {
		input := satisfyQuery[In](ctx)

		if input != nil {
			setCtxIfNeeded(input, ctx)
		}

		ctx.Locals("input", input)

		return ctx.Next()
	}

	for i, use := range opts[0].Uses {
		uses[i+1] = use
	}

	opts[0].Uses = uses

	return opts
}

func assertUnsupportedUseWithInput(opts []Opts) {
	if len(opts) > 0 && opts[0].UseWithInput {
		panic("UseWithInput cannot be used with methods that don't have input")
	}
}

func getMiddlewares(opts []Opts) []fiber.Handler {
	if len(opts) == 0 {
		return nil
	}

	var middlewares []fiber.Handler

	if opts[0].Use != nil {
		middlewares = append(middlewares, opts[0].Use)
	}
	middlewares = append(middlewares, opts[0].Uses...)

	return middlewares
}

func setCtxIfNeeded(input any, ctx fiber.Ctx) {
	if c, ok := input.(AwareCtx); ok {
		c.SetCtx(ctx)
	}
}

func Get(apiResource, url string, fn HandleFn, name string, opts ...Opts) {
	assertUnsupportedUseWithInput(opts)
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerVarO(apiResource, fullPath(url), http.MethodGet, name, opts[0].Out)
	} else {
		swaglay.RegisterHandler(apiResource, fullPath(url), http.MethodGet, name)
	}

	action := func(ctx fiber.Ctx) error {
		handle(ctx, fn)

		return nil
	}

	Fiber.Get(replacePath(url), action, getMiddlewares(opts)...)
}

func GetI[In any](apiResource, url string, fn HandleFnI[In], name string, opts ...Opts) {
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerIVarO[In](apiResource, fullPath(url), http.MethodGet, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerI[In](apiResource, fullPath(url), http.MethodGet, name)
	}

	var action fiber.Handler

	if len(opts) > 0 && opts[0].UseWithInput {
		opts = wrapSatisfyQueryInputMiddleware[In](opts)
		action = func(ctx fiber.Ctx) error {
			if input := ctx.Locals("input").(*In); input != nil {
				handleI(input, ctx, fn)
			}

			return nil
		}
	} else {
		action = func(ctx fiber.Ctx) error {
			if input := satisfyQuery[In](ctx); input != nil {
				setCtxIfNeeded(input, ctx)

				handleI(input, ctx, fn)
			}

			return nil
		}
	}

	Fiber.Get(replacePath(url), action, getMiddlewares(opts)...)
}

func GetO[Out any](apiResource, url string, fn HandleFnO[Out], name string, opts ...Opts) {
	assertUnsupportedUseWithInput(opts)
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerVarO(apiResource, fullPath(url), http.MethodGet, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerO[Out](apiResource, fullPath(url), http.MethodGet, name)
	}

	action := func(ctx fiber.Ctx) error {
		handleO(ctx, fn)

		return nil
	}

	Fiber.Get(replacePath(url), action, getMiddlewares(opts)...)
}

func GetIO[In any, Out any](apiResource, url string, fn HandleFnIO[In, Out], name string, opts ...Opts) {
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerIVarO[In](apiResource, fullPath(url), http.MethodGet, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerIO[In, Out](apiResource, fullPath(url), http.MethodGet, name)
	}

	var action fiber.Handler

	if len(opts) > 0 && opts[0].UseWithInput {
		opts = wrapSatisfyQueryInputMiddleware[In](opts)
		action = func(ctx fiber.Ctx) error {
			if input := ctx.Locals("input").(*In); input != nil {
				handleIO(input, ctx, fn)
			}

			return nil
		}
	} else {
		action = func(ctx fiber.Ctx) error {
			if input := satisfyQuery[In](ctx); input != nil {
				setCtxIfNeeded(input, ctx)
				handleIO(input, ctx, fn)
			}
			return nil
		}
	}

	Fiber.Get(replacePath(url), action, getMiddlewares(opts)...)
}

func Post(apiResource, url string, fn HandleFn, name string, opts ...Opts) {
	assertUnsupportedUseWithInput(opts)
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerVarO(apiResource, fullPath(url), http.MethodPost, name, opts[0].Out)
	} else {
		swaglay.RegisterHandler(apiResource, fullPath(url), http.MethodPost, name)
	}

	action := func(ctx fiber.Ctx) error {
		handle(ctx, fn)

		return nil
	}

	Fiber.Post(replacePath(url), action, getMiddlewares(opts)...)
}

func PostI[In any](apiResource, url string, fn HandleFnI[In], name string, opts ...Opts) {
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerIVarO[In](apiResource, fullPath(url), http.MethodPost, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerI[In](apiResource, fullPath(url), http.MethodPost, name)
	}

	var action fiber.Handler

	if len(opts) > 0 && opts[0].UseWithInput {
		opts = wrapBodyInputMiddleware[In](opts)
		action = func(ctx fiber.Ctx) error {
			if input := ctx.Locals("input").(*In); input != nil {
				handleI(input, ctx, fn)
			}

			return nil
		}
	} else {
		action = func(ctx fiber.Ctx) error {
			var input In

			setCtxIfNeeded(&input, ctx)

			if err := ctx.Bind().JSON(&input); err != nil {
				return ctx.Status(http.StatusUnprocessableEntity).JSON(NewResponseErrorBody(ctx, err))
			}

			handleI(&input, ctx, fn)

			return nil
		}
	}

	Fiber.Post(replacePath(url), action, getMiddlewares(opts)...)
}

func PostO[Out any](apiResource, url string, fn HandleFnO[Out], name string, opts ...Opts) {
	assertUnsupportedUseWithInput(opts)
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerVarO(apiResource, fullPath(url), http.MethodPost, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerO[Out](apiResource, fullPath(url), http.MethodPost, name)
	}

	action := func(ctx fiber.Ctx) error {
		handleO(ctx, fn)
		return nil
	}

	Fiber.Post(replacePath(url), action, getMiddlewares(opts)...)
}

func PostIO[In any, Out any](apiResource, url string, fn HandleFnIO[In, Out], name string, opts ...Opts) {
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerIVarO[In](apiResource, fullPath(url), http.MethodPost, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerIO[In, Out](apiResource, fullPath(url), http.MethodPost, name)
	}

	var action fiber.Handler

	if len(opts) > 0 && opts[0].UseWithInput {
		opts = wrapBodyInputMiddleware[In](opts)
		action = func(ctx fiber.Ctx) error {
			if input := ctx.Locals("input").(*In); input != nil {
				handleIO(input, ctx, fn)
			}
			return nil
		}
	} else {
		action = func(ctx fiber.Ctx) error {
			var input In
			setCtxIfNeeded(&input, ctx)
			if err := ctx.Bind().JSON(&input); err != nil {
				return ctx.Status(http.StatusUnprocessableEntity).JSON(NewResponseErrorBody(ctx, err))
			}
			handleIO(&input, ctx, fn)
			return nil
		}
	}

	Fiber.Post(replacePath(url), action, getMiddlewares(opts)...)
}

func Put(apiResource, url string, fn HandleFn, name string, opts ...Opts) {
	assertUnsupportedUseWithInput(opts)
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerVarO(apiResource, fullPath(url), http.MethodPut, name, opts[0].Out)
	} else {
		swaglay.RegisterHandler(apiResource, fullPath(url), http.MethodPut, name)
	}

	action := func(ctx fiber.Ctx) error {
		handle(ctx, fn)
		return nil
	}

	Fiber.Put(replacePath(url), action, getMiddlewares(opts)...)
}

func PutI[In any](apiResource, url string, fn HandleFnI[In], name string, opts ...Opts) {
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerIVarO[In](apiResource, fullPath(url), http.MethodPut, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerI[In](apiResource, fullPath(url), http.MethodPut, name)
	}

	var action fiber.Handler

	if len(opts) > 0 && opts[0].UseWithInput {
		opts = wrapBodyInputMiddleware[In](opts)
		action = func(ctx fiber.Ctx) error {
			if input := ctx.Locals("input").(*In); input != nil {
				handleI(input, ctx, fn)
			}
			return nil
		}
	} else {
		action = func(ctx fiber.Ctx) error {
			var input In
			setCtxIfNeeded(&input, ctx)
			if err := ctx.Bind().JSON(&input); err != nil {
				return ctx.Status(http.StatusUnprocessableEntity).JSON(NewResponseErrorBody(ctx, err))
			}
			handleI(&input, ctx, fn)
			return nil
		}
	}

	Fiber.Put(replacePath(url), action, getMiddlewares(opts)...)
}

func PutO[Out any](apiResource, url string, fn HandleFnO[Out], name string, opts ...Opts) {
	assertUnsupportedUseWithInput(opts)
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerVarO(apiResource, fullPath(url), http.MethodPut, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerO[Out](apiResource, fullPath(url), http.MethodPut, name)
	}

	action := func(ctx fiber.Ctx) error {
		handleO(ctx, fn)
		return nil
	}

	Fiber.Put(replacePath(url), action, getMiddlewares(opts)...)
}

func PutIO[In any, Out any](apiResource, url string, fn HandleFnIO[In, Out], name string, opts ...Opts) {
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerIVarO[In](apiResource, fullPath(url), http.MethodPut, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerIO[In, Out](apiResource, fullPath(url), http.MethodPut, name)
	}

	var action fiber.Handler

	if len(opts) > 0 && opts[0].UseWithInput {
		opts = wrapBodyInputMiddleware[In](opts)
		action = func(ctx fiber.Ctx) error {
			if input := ctx.Locals("input").(*In); input != nil {
				handleIO(input, ctx, fn)
			}
			return nil
		}
	} else {
		action = func(ctx fiber.Ctx) error {
			var input In
			setCtxIfNeeded(&input, ctx)
			if err := ctx.Bind().JSON(&input); err != nil {
				return ctx.Status(http.StatusUnprocessableEntity).JSON(NewResponseErrorBody(ctx, err))
			}
			handleIO(&input, ctx, fn)
			return nil
		}
	}

	Fiber.Put(replacePath(url), action, getMiddlewares(opts)...)
}

func Delete(apiResource, url string, fn HandleFn, name string, opts ...Opts) {
	assertUnsupportedUseWithInput(opts)
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerVarO(apiResource, fullPath(url), http.MethodDelete, name, opts[0].Out)
	} else {
		swaglay.RegisterHandler(apiResource, fullPath(url), http.MethodDelete, name)
	}

	action := func(ctx fiber.Ctx) error {
		handle(ctx, fn)
		return nil
	}

	Fiber.Delete(replacePath(url), action, getMiddlewares(opts)...)
}

func DeleteI[In any](apiResource, url string, fn HandleFnI[In], name string, opts ...Opts) {
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerIVarO[In](apiResource, fullPath(url), http.MethodDelete, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerI[In](apiResource, fullPath(url), http.MethodDelete, name)
	}

	var action fiber.Handler

	if len(opts) > 0 && opts[0].UseWithInput {
		opts = wrapSatisfyQueryInputMiddleware[In](opts)
		action = func(ctx fiber.Ctx) error {
			if input := ctx.Locals("input").(*In); input != nil {
				handleI(input, ctx, fn)
			}
			return nil
		}
	} else {
		action = func(ctx fiber.Ctx) error {
			if input := satisfyQuery[In](ctx); input != nil {
				setCtxIfNeeded(input, ctx)

				handleI(input, ctx, fn)
			}

			return nil
		}
	}

	Fiber.Delete(replacePath(url), action, getMiddlewares(opts)...)
}

func DeleteO[Out any](apiResource, url string, fn HandleFnO[Out], name string, opts ...Opts) {
	assertUnsupportedUseWithInput(opts)
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerVarO(apiResource, fullPath(url), http.MethodDelete, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerO[Out](apiResource, fullPath(url), http.MethodDelete, name)
	}

	action := func(ctx fiber.Ctx) error {
		handleO(ctx, fn)
		return nil
	}

	Fiber.Delete(replacePath(url), action, getMiddlewares(opts)...)
}

func DeleteIO[In any, Out any](apiResource, url string, fn HandleFnIO[In, Out], name string, opts ...Opts) {
	swaglay.MustEmptyOrOneLength(opts)

	if len(opts) > 0 && opts[0].Out != nil {
		swaglay.RegisterHandlerIVarO[In](apiResource, fullPath(url), http.MethodDelete, name, opts[0].Out)
	} else {
		swaglay.RegisterHandlerIO[In, Out](apiResource, fullPath(url), http.MethodDelete, name)
	}

	var action fiber.Handler

	if len(opts) > 0 && opts[0].UseWithInput {
		opts = wrapSatisfyQueryInputMiddleware[In](opts)
		action = func(ctx fiber.Ctx) error {
			if input := ctx.Locals("input").(*In); input != nil {
				handleIO(input, ctx, fn)
			}
			return nil
		}
	} else {
		action = func(ctx fiber.Ctx) error {
			if input := satisfyQuery[In](ctx); input != nil {
				setCtxIfNeeded(input, ctx)

				handleIO(input, ctx, fn)
			}

			return nil
		}
	}

	Fiber.Delete(replacePath(url), action, getMiddlewares(opts)...)
}
