package configserver

import (
	"html/template"
	"io"
	"log"
)

type LayoutData struct {
	NavItems []NavItem
	ActiveNavItem NavItem
	BodyData any
}

type NavItem struct {
	Path string
	Title string
}

var navItems = []NavItem{
	{Path: "/", Title: "Connect"},
	{Path: "/devices", Title: "Devices"},
}

func render(wr io.Writer, templateName string, activePath string, data any) error {
	tmpl := template.Must(template.ParseGlob("configserver/templates/*"))

	var activeNavItem NavItem

	for _, i := range navItems {
		if i.Path == activePath {
			activeNavItem = i
			break
		}
	}

	data = LayoutData{
		NavItems:            navItems,
		ActiveNavItem:   	 activeNavItem,
		BodyData:            data,
	}
	log.Default().Printf("Rendering template %s with data %v", templateName, data)

	return tmpl.ExecuteTemplate(wr, templateName, data)
}
