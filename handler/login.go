package handler

import (
	"fmt"
	"net/http"
	"github.com/labstack/echo"
	"zdog/model"
)

const (
)

func (h *Handler) Login(c echo.Context) (err error) {
	data := map[string]interface{} {
		"Title" : "Login",
		"Messages" : make([][]string, 0),
	}

	u := GetUser(c)
	if u != nil {
		message := []string {"error", "already login " + u.Name + " !", }
		data["Messages"] = append(data["Messages"].([][]string), message)
		data["User"] = u.ToJson()
		return c.Render(http.StatusOK, "login", data)
	}
	
	u = &model.User{}
	if err = c.Bind(u); err != nil {
		message := []string {"error", err.Error(), }
		data["Messages"] = append(data["Messages"].([][]string), message)
		return c.Render(http.StatusBadRequest, "login", data)
	}
	if !u.Check() {
		message := []string {"error", fmt.Sprintf("invalid input [name -> %s] or [password -> %s]", u.Name, u.Password), }
		data["Messages"] = append(data["Messages"].([][]string), message)
		return c.Render(http.StatusBadRequest, "login", data)
	}

	u_m := u.Find()
	if u_m == nil {
		u.Save()
	} else {
		if u_m.Password != u.Password {
			message := []string {"error", fmt.Sprintf("incorret [name -> %s] or [password -> %s]", u.Name, u.Password), }
			data["Messages"] = append(data["Messages"].([][]string), message)
			return c.Render(http.StatusBadRequest, "login", data)
		}
	}
	
	// CreateToken
	token := u.GenToken()	
	if len(token) == 0 {
		message := []string {"error", "error oops while creating token", }
		data["Messages"] = append(data["Messages"].([][]string), message)
		return c.Render(http.StatusBadRequest, "login", data)
	}
	SetCookie(c, "token", token)

	message := []string {"success", "login success !", }
	data["Messages"] = append(data["Messages"].([][]string), message)
	// set redirect
	data["From"] = "/login" 
	data["To"] = "/index" 
	data["Time"] = "0.5" 
	return c.Render(http.StatusOK, "redirect", data)
}

func (h *Handler) Logout(c echo.Context) error {
	data := map[string]interface{} {
		"Title" : "Logout",
		"Messages" : make([][]string, 0),
	}
	u := GetUser(c)
	data["From"] = "/logout"
	data["Time"] = "0.5"
	if u == nil {
		message := []string {"error", "need login before logout !", }
		data["Messages"] = append(data["Messages"].([][]string), message)
		// set redirect
		data["To"] = "/login" 
	} else {
		SetCookie(c, "token", "")
		data["To"] = "/index"
		message := []string {"success", "logout success !", }
		data["Messages"] = append(data["Messages"].([][]string), message)
	}
	return c.Render(http.StatusOK, "redirect", data)
}

func (h *Handler) GetLogin(c echo.Context) error {
	return h.Render(c, "login")
 }
