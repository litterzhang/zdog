package handler

import (
	"time"
	"strings"
	"net/http"
	"zdog/model"
	"github.com/labstack/echo"
)

type Handler struct {
}

func New() * Handler {
	return &Handler{}
}

func (h *Handler) Render(c echo.Context, name string) error {
	data := map[string]interface{} {}
	u := GetUser(c)
	if u != nil {
		data["User"] = u.ToJson()
	}

	title := ""
	if len(name) == 0 {
		title = "Index"
		name = "index"
	} else {
		title = strings.ToUpper(name[0:1]) + name[1:]
	}

	data["Title"] = title
    return c.Render(http.StatusOK, name, data)
}

func (h *Handler) RouteRender(c echo.Context) error {
	name := c.Param("route")
	return h.Render(c, name)
}

func ErrorHandler(err error, c echo.Context) {
	c.Logger().Error(err)
	c.Redirect(http.StatusFound, "/error")
}

func SetCookie(c echo.Context, k string, v string) {
	cookie := new(http.Cookie)
	cookie.Name = k
	cookie.Value = v
	cookie.Expires = time.Now().Add(24 * time.Hour)
	c.SetCookie(cookie)
}

func GetCookie(c echo.Context, k string) (v string) {
	cookie, err := c.Cookie(k)
	if err != nil {
		return 
	}
	return cookie.Value
}

func GetUser(c echo.Context) *model.User {
	var u *model.User = nil

	_token := GetCookie(c, "token")
	u = model.UserFromToken(_token)
	return u
}
