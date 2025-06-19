package main

import (
	"context"
	"fiber/pkg/constants"
	"fiber/pkg/controllers"
	"github.com/KoNekoD/swaglay/pkg"
	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v3"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	r := fiber.New()

	controllers.InitAllControllers(r)

	includeApiDocs(r)

	srv := &fasthttp.Server{Handler: r.Handler()}

	log.Info("Server started at http://localhost:8080")
	log.Info("OpenAPI json at http://localhost:8080/api/docs/openapi.json")

	go func() {
		err := srv.ListenAndServe(constants.ServerAddr)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)

	<-quit

	log.Info("Shutdown Server with timeout of 5 seconds...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.ShutdownWithContext(ctx); err != nil {
		log.Fatal("Server Shutdown(context deadline):", err)
	}

	log.Info("Server exiting")
}

func includeApiDocs(r *fiber.App) {
	apiSpec, _ := swaglay.Api.Spec()

	swaggerBytes, err := apiSpec.MarshalJSON()
	if err != nil {
		panic(err)
	}

	// Examples:
	// * Stoplight Elements https://stoplight.io/open-source/elements
	// * Redocly https://redocly.com/

	//htmlBytes, err := os.ReadFile("docs/index.html")
	//if err != nil {
	//	panic(err)
	//}
	//
	//jsBytes, err := os.ReadFile("docs/web-components.min.js")
	//if err != nil {
	//	panic(err)
	//}
	//
	//cssBytes, err := os.ReadFile("docs/styles.min.css")
	//if err != nil {
	//	panic(err)
	//}

	getSwagger := func(ctx fiber.Ctx) error {
		ctx.Set("Content-Type", constants.JsonContentType)
		return ctx.Send(swaggerBytes)
	}

	//getDocsHtml := func(ctx fiber.Ctx) error {
	//	ctx.Set("Content-Type", constants.HtmlContentType)
	//	return ctx.Send(htmlBytes)
	//}
	//
	//getDocsJs := func(ctx fiber.Ctx) error {
	//	ctx.Set("Content-Type", constants.JsContentType)
	//	return ctx.Send(jsBytes)
	//}
	//
	//getDocsCss := func(ctx fiber.Ctx) error {
	//	ctx.Set("Content-Type", constants.CssContentType)
	//	return ctx.Send(cssBytes)
	//}

	r.Group("/api/docs").
		//Get("", getDocsHtml).
		//Get("/web-components.min.js", getDocsJs).
		//Get("/styles.min.css", getDocsCss).
		Get("/openapi.json", getSwagger)
}
