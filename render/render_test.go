package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var pageData = []struct {
	name          string
	renderer      string
	templateName  string
	errorExpected bool
	errorMessage  string
}{
	{"go_page", "go", "home", false, "error rendering go template"},
	{"go_page", "go", "non-existing", true, "no error rendering non-existent go page when one is expected"},
	{"jet_page", "jet", "home", false, "error rendering jet template"},
	{"jet_page", "jet", "non-existing", true, "no error rendering non-existent jet page when one is expected"},
	{"invalid_render_engine", "foo", "home", true, "No error rendering with invalid rendering engine"},
}

func TestRender_Page(t *testing.T) {
	for _, e := range pageData {
		r, err := http.NewRequest("GET", "/some-url", nil)
		if err != nil {
			t.Error(err)
		}

		w := httptest.NewRecorder()

		testRenderer.RootPath = "./test-data"
		testRenderer.Renderer = e.renderer

		err = testRenderer.Page(w, r, e.templateName, nil, nil)
		if e.errorExpected {
			if err == nil {
				t.Errorf("%s: %s", e.name, e.errorMessage)
			}
		} else {
			if err != nil {
				t.Errorf("%s: %s: %s", e.name, e.errorMessage, err.Error())
			}
		}
	}
}

func TestRender_GoPage(t *testing.T) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()

	testRenderer.Renderer = "go"

	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering go page", err)
	}
	err = testRenderer.Page(w, r, "non-existent", nil, nil)
	if err == nil {
		t.Error("Error rendering non-existent go template", err)
	}

}

func TestRender_JetPage(t *testing.T) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()

	testRenderer.Renderer = "jet"

	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering jet page", err)
	}
	err = testRenderer.Page(w, r, "non-existent", nil, nil)
	if err == nil {
		t.Error("Error rendering non-existent jet template", err)
	}

}
