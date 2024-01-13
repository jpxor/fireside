package main

import (
	"fireside/cmd/fireside-server/handlers"
	"flag"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
)

var (
	port = flag.String("port", "3000", "Port to listen on")
	// wdir  = flag.String("wdir", ".", "Working directory")
	devel = flag.Bool("devel", false, "Development mode")
)

func main() {
	flag.Parse()

	tmplEngine := html.New("./www/templates", "")
	if *devel {
		// Reload the templates on each render
		tmplEngine.Reload(true)
	}

	app := fiber.New(fiber.Config{
		AppName: "fireside",
		Views:   tmplEngine,
	})

	app.Use(favicon.New())
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(handlers.ResetLoginExpirationMiddleware())

	if *devel {
		// prevent browser from caching replies
		app.Use(handlers.SetNoCacheHeaders)
		app.Get("/debug/users", handlers.DebugListUsers)
	}

	tmpl := app.Group("/render/")
	tmpl.Get("hello", handlers.RenderHello)

	api := app.Group("/api/")
	api.Post("user/create", handlers.UserCreate)
	api.Post("user/verify", handlers.UserVerify)
	api.Post("user/login", handlers.UserLogin)

	app.Static("/assets/", "./www/assets/")
	app.Static("/", "./www/pages/")

	log.Fatal(app.Listen(":" + *port))
}
