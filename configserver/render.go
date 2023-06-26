package configserver

import (
	"html/template"
	"io"
	"log"
)

type LayoutData struct {
	NavItems      []NavItem
	ActiveNavItem NavItem
	BodyData      any
	ErrorMsg      string
}

type NavItem struct {
	Path      string
	Title     string
	IconClass string
}

var navItems = []NavItem{
	{Path: "/", Title: "Overview", IconClass: "bi-house"},
	{Path: "/app", Title: "App Config", IconClass: "bi-cloud"},
	{Path: "/uploads", Title: "Data Uploads", IconClass: "bi-cloud-upload"},
	{Path: "/connect-audio", Title: "Connect Audio", IconClass: "bi-telephone"},
	{Path: "/cad", Title: "CAD", IconClass: "bi-database"},
	//{Path: "/devices", Title: "Devices", IconClass: ""},
}

func renderWithError(wr io.Writer, templateName string, activePath string, data any, err string) error {
	tmpl := template.Must(template.ParseGlob("configserver/templates/*"))

	var activeNavItem NavItem

	for _, i := range navItems {
		if i.Path == activePath {
			activeNavItem = i
			break
		}
	}

	data = LayoutData{
		NavItems:      navItems,
		ActiveNavItem: activeNavItem,
		BodyData:      data,
		ErrorMsg:      err,
	}

	log.Default().Printf("Rendering template %s with data %v", templateName, data)

	return tmpl.ExecuteTemplate(wr, templateName, data)
}

func render(wr io.Writer, templateName string, activePath string, data any) error {
	return renderWithError(wr, templateName, activePath, data, "")
}
