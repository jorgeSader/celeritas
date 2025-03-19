package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
)

type Render struct {
	Renderer      string
	RootPath      string
	Secure        bool
	Port          string
	ServerName    string
	JetViews      *jet.Set
	Session       *scs.SessionManager
	TemplateCache map[string]*template.Template
	UseCache      bool
}

type TemplateData struct {
	IsAuthenticated bool
	IntMap          map[string]int
	StringMap       map[string]string
	FloatMap        map[string]float64
	Data            map[string]interface{}
	CSRFToken       string
	Port            string
	ServerName      string
	Secure          bool
}

func (d *Render) defaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Secure = d.Secure
	td.ServerName = d.ServerName
	td.Port = d.Port

	if d.Session != nil && r != nil {
		if d.Session.Exists(r.Context(), "userID") {
			td.IsAuthenticated = true
		}
	}
	return td
}

func (d *Render) Page(w http.ResponseWriter, r *http.Request, view string, data, variables interface{}) error {
	switch strings.ToLower(d.Renderer) {
	case "go":
		return d.GoPage(w, r, view, data)
	case "jet":
		return d.JetPage(w, r, view, data, variables)
	default:
		return errors.New("no rendering engine specified")
	}
}

// GoPage renders a standard Go template using the pre-cached template
func (d *Render) GoPage(w http.ResponseWriter, r *http.Request, view string, data interface{}) error {
	var tc map[string]*template.Template
	var err error

	if d.UseCache {
		tc = d.TemplateCache
	} else {
		tc, err = d.CreateTemplateCache()
		if err != nil {
			log.Printf("Error creating template cache: %v", err)
			return err
		}
	}

	// Get requested template from cache
	tmpl, ok := tc[view+".page.tmpl"]
	if !ok {
		return fmt.Errorf("can't get template %s.page.tmpl from cache", view)
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}
	td = d.defaultData(td, r)

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, td)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return err
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		log.Printf("Error writing template to browser: %v", err)
		return err
	}

	return nil
}

// JetPage renders a template using the Jet templating engine
func (d *Render) JetPage(w http.ResponseWriter, r *http.Request, templateName string, data, variables interface{}) error {
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

	td = d.defaultData(td, r)

	jt, err := d.JetViews.GetTemplate(fmt.Sprintf("%s.jet", templateName))
	if err != nil {
		log.Println(err)
		return err
	}

	err = jt.Execute(w, vars, td)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// CreateTemplateCache initializes a template cache by parsing all template files
func (d *Render) CreateTemplateCache() (map[string]*template.Template, error) {
	templateCache := make(map[string]*template.Template)

	pathToPages := fmt.Sprintf("%s/views/*.page.tmpl", d.RootPath)
	pathToLayouts := fmt.Sprintf("%s/views/layouts/*.layout.tmpl", d.RootPath)

	pages, err := filepath.Glob(pathToPages)
	if err != nil {
		return templateCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return templateCache, err
		}

		layouts, err := filepath.Glob(pathToLayouts)
		if err != nil {
			return templateCache, err
		}

		if len(layouts) > 0 {
			ts, err = ts.ParseGlob(pathToLayouts)
			if err != nil {
				return templateCache, err
			}
		}

		templateCache[name] = ts
	}

	return templateCache, nil
}
