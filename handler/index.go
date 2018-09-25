package handler

import (
	"net/http"
	"github.com/labstack/echo"
)

const (
	TEMPLATE_NAME = "index"
)

func (h *Handler) Index(c echo.Context) error {
	data := map[string]interface{} {
         "Title" : "Index",
    }
	u := GetUser(c)
	if u != nil {
		data["User"] = u.ToJson()
	}
    return c.Render(http.StatusOK, TEMPLATE_NAME, data)
}
