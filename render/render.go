package render

import (
	"io"
	"html/template"
	"github.com/labstack/echo"
)

const (
	TEMPLATE_PATH = "./template/*.html"
)

func New() *Template {
	return &Template {
		tmpls: template.Must(template.ParseGlob(TEMPLATE_PATH)),
	}
}

type Template struct {
	tmpls *template.Template;
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.tmpls.ExecuteTemplate(w, name, data)
}
