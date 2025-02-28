package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRender_Page(t *testing.T) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()

	testRenderer.RootPath = "./test-data"

	testRenderer.Renderer = "go"
	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering go page", err)
	}
	err = testRenderer.Page(w, r, "non-existent", nil, nil)
	if err == nil {
		t.Error("Error rendering non-existent go template", err)
	}

	testRenderer.Renderer = "jet"
	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering jet page", err)
	}
	err = testRenderer.Page(w, r, "non-existent", nil, nil)
	if err == nil {
		t.Error("Error rendering non-existent jet template", err)
	}

	testRenderer.Renderer = ""
	err = testRenderer.Page(w, r, "home", nil, nil)
	if err == nil {
		t.Error("No error returned while rendering with invalid renderer specified", err)
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
