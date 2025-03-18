package render

import (
	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./test-data/views"),
	jet.InDevelopmentMode(),
)

var testRenderer = Render{
	Renderer: "",
	RootPath: "",
	JetViews: views,
	Session:  scs.New(),
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func setupTestRequest(t *testing.T) (*http.Request, *httptest.ResponseRecorder) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	// Use an empty token to start a new session
	ctx, err := testRenderer.Session.Load(r.Context(), "")
	if err != nil {
		t.Fatalf("Failed to load session context: %v", err)
	}
	r = r.WithContext(ctx) // Update request with session context

	// Simulate middleware by writing the session cookie
	testRenderer.Session.LoadAndSave(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w, r)

	return r, w
}
