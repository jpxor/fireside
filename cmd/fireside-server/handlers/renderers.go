package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func RenderHello(c *fiber.Ctx) error {
	return c.Render("hello.html", fiber.Map{
		"Name": name,
	})
}
