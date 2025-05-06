package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data any, c echo.Context) error {
	if (c.Request().Header.Get("HX-Request") == "true") {
		return t.templates.ExecuteTemplate(w, name, data)
	}

	// Non htmx request render using htmx-template
    if data == nil {
        data = map[string]string{}
    }

    data.(map[string]string)["ContentName"] = name

	return t.templates.ExecuteTemplate(w, "htmx-template", data)
}


func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
}
