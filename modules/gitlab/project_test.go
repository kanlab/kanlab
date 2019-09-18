package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var projectCollection = []byte(`[{
			"id": 1,
			"description": "My project",
			"name":"my project",
			"name_with_namespace": "admin / my-project",
			"path_with_namespace": "admin/my-project"
		}, {
			"id": 2,
			"description": "My project2",
			"name":"my project2",
			"name_with_namespace": "admin / my-project2",
			"path_with_namespace": "admin/my-project2"
		}]`)

func TestClient_ListProjectsSuccess(t *testing.T) {
	c, teardown, _ := setup()
	defer teardown()

	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Write(projectCollection)
	})

	pr, collOpt, err := c.ListProjects(&ProjectListOptions{})

	assert.Nil(t, err)
	assert.NotEmpty(t, pr)
	assert.NotEmpty(t, collOpt)
	assert.Equal(t, int64(1), (*pr[0]).Id)
	assert.Equal(t, int64(2), (*pr[1]).Id)
	assert.Equal(t, "my project", (*pr[0]).Name)
}

func TestClient_ListProjectsWithErr(t *testing.T) {
	c, teardown, _ := setup()
	defer teardown()

	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{[}`))
	})
	pr, collOpt, err := c.ListProjects(&ProjectListOptions{})
	assert.NotNil(t, err)
	assert.Empty(t, pr)
	assert.Empty(t, collOpt)
}

func TestClient_ListProjectsRequestParams(t *testing.T) {
	c, teardown, _ := setup()
	defer teardown()

	opt := &ProjectListOptions{
		Archived: "false",
		Search:   "qwerty",
		ListOptions: ListOptions{
			Page:    "1",
			PerPage: "1",
		},
	}

	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, opt.Page, r.URL.Query().Get("page"))
		assert.Equal(t, opt.PerPage, r.URL.Query().Get("per_page"))
		assert.Equal(t, opt.Archived, r.URL.Query().Get("archived"))
		assert.Equal(t, opt.Search, r.URL.Query().Get("search"))
	})

	c.ListProjects(opt)
}

func TestClient_StarredProjectsSuccess(t *testing.T) {
	c, teardown, _ := setup()
	defer teardown()

	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Write(projectCollection)
	})

	pr, collOpt, err := c.StarredProjects(&ProjectListOptions{})

	assert.Nil(t, err)
	assert.NotEmpty(t, pr)
	assert.NotEmpty(t, collOpt)
	assert.Equal(t, int64(1), (*pr[0]).Id)
	assert.Equal(t, int64(2), (*pr[1]).Id)
	assert.Equal(t, "my project", (*pr[0]).Name)
}

func TestClient_ListProjectsErr(t *testing.T) {
	c, teardown, _ := setup()
	defer teardown()

	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{[}`))
	})

	pr, opt, err := c.StarredProjects(&ProjectListOptions{})
	assert.NotNil(t, err)
	assert.Empty(t, pr)
	assert.Empty(t, opt)
}

func TestClient_StarredProjectsRequestCheck(t *testing.T) {
	c, teardown, _ := setup()
	defer teardown()

	opt := &ProjectListOptions{
		Archived: "false",
		Search:   "false",
		Starred:  "true",
		ListOptions: ListOptions{
			Page:    "1",
			PerPage: "10",
		},
	}

	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, opt.Archived, r.URL.Query().Get("archived"))
		assert.Equal(t, opt.Page, r.URL.Query().Get("page"))
		assert.Equal(t, opt.PerPage, r.URL.Query().Get("per_page"))
		assert.Equal(t, opt.Search, r.URL.Query().Get("search"))
		assert.Equal(t, opt.Starred, r.URL.Query().Get("starred"))
	})

	c.StarredProjects(opt)
}

func TestClient_ItemProjectSuccess(t *testing.T) {
	c, teardown, _ := setup()
	defer teardown()

	mux.HandleFunc("/projects/admin/my-project", func(w http.ResponseWriter, r *http.Request) {
		data := []byte(`{
			"id": 1,
			"description": "My project",
			"name":"my project",
			"name_with_namespace": "admin / my-project",
			"path_with_namespace": "admin/my-project"
		}`)

		w.Write(data)
	})

	pr, err := c.ItemProject("admin/my-project")

	assert.Nil(t, err)
	assert.NotEmpty(t, pr)
	assert.Equal(t, int64(1), pr.Id)
	assert.Equal(t, "My project", pr.Description)
	assert.Equal(t, "my project", pr.Name)
	assert.Equal(t, "admin / my-project", pr.NamespaceWithName)
	assert.Equal(t, "admin/my-project", pr.PathWithNamespace)
}

func TestClient_ItemProjectError(t *testing.T) {
	c, teardown, _ := setup()
	defer teardown()

	mux.HandleFunc("/projects/admin/my-project", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{]`))
	})

	pr, err := c.ItemProject("admin/my-project")
	assert.NotNil(t, err)
	assert.Empty(t, pr)
}
