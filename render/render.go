package render

import (
	"errors"
	"fmt"
	"github.com/justinas/nosurf"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
)

type Render struct {
	Renderer   string
	RootPath   string
	Secure     bool
	Port       string
	ServerName string
	JetViews   *jet.Set
	Session    *scs.SessionManager
}

type TemplateData struct {
	IsAuthenticated bool
	IntMap          map[string]int
	StringMap       map[string]string
	FloatMap        map[string]float32
	Data            map[string]interface{}
	CSRFToken       string
	Port            string
	ServerName      string
	Secure          bool
	Error           string
	Flash           string
}

func (b *Render) defaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Secure = b.Secure
	td.ServerName = b.ServerName
	td.CSRFToken = nosurf.Token(r)
	td.Port = b.Port
	if b.Session.Exists(r.Context(), "userID") {
		td.IsAuthenticated = true
	}
	td.Error = b.Session.PopString(r.Context(), "error")
	td.Flash = b.Session.PopString(r.Context(), "flash")
	return td
}

// Page handles the template engine: rendering of the templates (go or jet)
func (b *Render) Page(w http.ResponseWriter, r *http.Request, view string, variables, data interface{}) error {
	switch strings.ToLower(b.Renderer) {
	case "go":
		return b.GoPage(w, r, view, data)
	case "jet":
		return b.JetPage(w, r, view, variables, data)
	default:

	}

	return errors.New("no rendering engine specified")
}

// GoPage renders a standard Go template
func (b *Render) GoPage(w http.ResponseWriter, r *http.Request, view string, data interface{}) error {
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/views/%s.page.tmpl", b.RootPath, view))
	if err != nil {
		return err
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	err = tmpl.Execute(w, &td)
	if err != nil {
		return err
	}

	return nil
}

// JetPage renders a template using the Jet templating engine
func (b *Render) JetPage(w http.ResponseWriter, r *http.Request, templateName string, variables, data interface{}) error {
	// create empty jet map variable
	var vars jet.VarMap

	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	td = b.defaultData(td, r)

	t, err := b.JetViews.GetTemplate(fmt.Sprintf("%s.jet", templateName))
	if err != nil {
		log.Println(err)
		return err
	}

	if err = t.Execute(w, vars, td); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
