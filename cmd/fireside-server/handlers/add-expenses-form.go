package handlers

import (
	"bytes"
	"fireside/app"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func RenderAddExpenses(c *fiber.Ctx) error {
	data := fiber.Map{
		"ReloadRecentTx": c.Locals("ReloadRecentTx"),
	}
	return c.Render("add-expenses.html", data)
}

type addExpensesData struct {
	FromAccount string
	Date        []string
	ExpAccount  []string
	Amount      []string
	Description []string
}

func PostAddExpenses(c *fiber.Ctx) error {
	sess, err := parseSessionCookie(c.Cookies("session"))
	if err != nil {
		c.ClearCookie("session")
		c.Set("HX-Redirect", "/login")
		return c.SendStatus(fiber.StatusOK)
	}
	// requires custom parsing because form can be an array of
	// inputs, reusing names
	data, err := parseMultiExpenseForm(c.Body())
	if err != nil {
		return err
	}
	txStr := expenseInputsToPlaintext(data)
	err = app.AppendPlaintext(sess.ID, sess.SelectedFile, txStr)
	if err != nil {
		return err
	}
	c.Locals("ReloadRecentTx", true)
	return RenderAddExpenses(c)
}

func parseMultiExpenseForm(body []byte) (data addExpensesData, err error) {
	for _, input := range bytes.Split(body, []byte{'&'}) {
		kvpair := bytes.Split(input, []byte{'='})
		if len(kvpair) != 2 {
			return data, fmt.Errorf("input not correctly parsed")
		}
		key := string(kvpair[0])
		val, err := url.PathUnescape(string(kvpair[1]))
		if err != nil {
			log.Println(err)
			return data, fmt.Errorf("failed to parse input string")
		}
		switch key {

		case "date":
			data.Date = append(data.Date, val)

		case "expcat":
			if !strings.HasPrefix(val, "expense") {
				val = "expenses:" + val
			}
			data.ExpAccount = append(data.ExpAccount, val)

		case "amount":
			data.Amount = append(data.Amount, val)

		case "desc":
			data.Description = append(data.Description, val)

		case "fromacct":
			data.FromAccount = val
		}
	}
	return data, nil
}

func expenseInputsToPlaintext(data addExpensesData) string {
	s := strings.Builder{}
	for i := 0; i < len(data.Date); i++ {
		s.WriteString(data.Date[i])
		if len(data.Description[i]) > 0 {
			s.WriteByte(' ')
			s.WriteString(data.Description[i])
		}
		s.WriteString("\n\t")
		s.WriteString(data.ExpAccount[i])
		s.WriteString("\t")
		s.WriteString(data.Amount[i])
		s.WriteString("\n\t")
		s.WriteString(data.FromAccount)
		s.WriteString("\n\n")
	}
	return s.String()
}
