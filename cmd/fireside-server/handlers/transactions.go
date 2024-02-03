package handlers

import (
	"fireside/app"
	"time"

	"github.com/gofiber/fiber/v2"
)

type recentTxRenderData struct {
	Transactions []string
}

func RenderRecentTransactions(c *fiber.Ctx) error {
	sess, err := parseSessionCookie(c.Cookies("session"))
	if err != nil {
		c.ClearCookie("session")
		c.Set("HX-Redirect", "/login")
		return c.SendStatus(fiber.StatusOK)
	}
	since := time.Now().AddDate(0, 0, -30)
	txs, err := app.RecentTransactions(sess.ID, sess.SelectedFile, since)
	if err != nil {
		// TODO: return error
		return c.Render("recent-tx.html", nil)
	}
	data := recentTxRenderData{
		Transactions: app.TxStringify(txs),
	}
	return c.Render("recent-tx.html", data)
}
