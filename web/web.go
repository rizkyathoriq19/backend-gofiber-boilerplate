package web

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

var templates *template.Template

func init() {
	templates = template.Must(template.ParseFS(templateFS, "templates/*.html"))
}

// PageData holds data for template rendering
type PageData struct {
	AppName string
}

// RenderIndex renders the index/welcome page
func RenderIndex(appName string) (string, error) {
	return render("index.html", PageData{AppName: appName})
}

// RenderHealth renders the health check page
func RenderHealth(appName string) (string, error) {
	return render("health.html", PageData{AppName: appName})
}

// Render404 renders the 404 error page
func Render404() (string, error) {
	return render("404.html", nil)
}

// RenderNotFound renders the resource not found page
func RenderNotFound() (string, error) {
	return render("not_found.html", nil)
}

func render(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
