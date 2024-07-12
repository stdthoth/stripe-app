package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type templateData struct {
	StringMap            map[string]string
	IntMap               map[string]int
	FloatMap             map[string]float32
	Data                 map[string]interface{}
	CSRFToken            string
	Flash                string
	Warning              string
	Error                string
	IsAuthenticated      int
	API                  string
	CSSVersion           string
	StripeSecretKey      string
	StripePublishableKey string
}

var functions = template.FuncMap{
	"formatCurrency": formatCurrency,
}

func formatCurrency(n int) string {
	f := float32(n) / float32(100)
	return fmt.Sprintf("$%.2f", f)
}

//go:embed templates
var templateFS embed.FS

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	td.API = app.config.api
	td.StripePublishableKey = app.config.stripeInfo.key
	td.StripeSecretKey = app.config.stripeInfo.secret
	return td
}

func (app *application) renderTemplates(w http.ResponseWriter, r *http.Request, page string, td *templateData, partials ...string) error {
	var t *template.Template
	var err error

	toRender := fmt.Sprintf("templates/%s.page.gohtml", page)
	_, templateInMap := app.templateCache[toRender]

	if app.config.env == "production" && templateInMap {
		t = app.templateCache[toRender]
	} else {
		t, err = app.parseTemplate(partials, page, toRender)
		if err != nil {
			app.errorLog.Println(err)
			return err
		}
	}
	if td == nil {
		td = &templateData{}
	}
	td = app.addDefaultData(td, r)

	err = t.Execute(w, td)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}

func (app *application) parseTemplate(partials []string, page, toRerender string) (*template.Template, error) {
	var t *template.Template
	var err error

	//build partials
	if len(partials) > 0 {
		for i, x := range partials {
			partials[i] = fmt.Sprintf("templates/%s.partial.gohtml", x)
		}
	}

	if len(partials) > 0 {
		t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.gohtml", strings.Join(partials, ","), toRerender)
	} else {
		t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.gohtml", toRerender)
	}
	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	app.templateCache[toRerender] = t

	return t, err
}
