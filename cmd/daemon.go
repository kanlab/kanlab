package cmd

import (
	"log"
	"net/http"

	"github.com/go-macaron/bindata"
	"github.com/go-macaron/binding"
	"github.com/kanlab/sockets"
	"github.com/spf13/cobra"
	"github.com/kanlab/kanlab"
	"github.com/kanlab/kanlab/templates"
	"github.com/kanlab/kanlab/web"
	"github.com/kanlab/kanlab/ws"
	"gopkg.in/macaron.v1"

	"github.com/kanlab/kanlab/modules/auth"
	"github.com/kanlab/kanlab/modules/setting"

	"github.com/kanlab/kanlab/models"
	"github.com/kanlab/kanlab/modules/middleware"
	"github.com/kanlab/kanlab/routers"
	"github.com/kanlab/kanlab/routers/board"
	"github.com/kanlab/kanlab/routers/user"

	"github.com/spf13/viper"
)

// DaemonCmd is implementation of command to run application in daemon mode
var DaemonCmd = cobra.Command{
	Use:   "server",
	Short: "Starts LeanLabs Kanban board application",
	Long: `Start LeanLabs Kanban board application.

Please refer to http://kanban.leanlabs.io/docs/ for full documentation.

Report bugs to <support@leanlabs.io> or https://gitter.im/leanlabsio/kanban.
        `,
	Run: daemon,
}

func init() {
	DaemonCmd.Flags().String(
		"server-listen",
		"0.0.0.0:80",
		"IP:PORT to listen on",
	)
	DaemonCmd.Flags().String(
		"server-hostname",
		"http://localhost",
		"URL on which Leanlabs Kanban will be reachable",
	)
	DaemonCmd.Flags().String(
		"security-secret",
		"qwerty",
		"This string is used to generate user auth tokens",
	)
	DaemonCmd.Flags().String(
		"gitlab-url",
		"https://gitlab.com",
		"Your GitLab host URL",
	)
	DaemonCmd.Flags().String(
		"gitlab-client",
		"qwerty",
		"Your GitLab OAuth client ID",
	)
	DaemonCmd.Flags().String(
		"gitlab-secret",
		"qwerty",
		"Your GitLab OAuth client secret key",
	)
	DaemonCmd.Flags().String(
		"redis-addr",
		"127.0.0.1:6379",
		"Redis server address - IP:PORT",
	)
	DaemonCmd.Flags().String(
		"redis-password",
		"",
		"Redis server password, empty string if none",
	)
	DaemonCmd.Flags().Int64(
		"redis-db",
		0,
		"Redis server database numeric index, from 0 to 16",
	)
	DaemonCmd.Flags().Bool(
		"enable-signup",
		true,
		"Enable signup",
	)
	DaemonCmd.Flags().Bool(
		"auto-comments",
		true,
		"Comment add if stage changed",
	)
}

func daemon(c *cobra.Command, a []string) {
	m := macaron.New()
	setting.NewContext(c)
	db := setting.NewDbClient()
	m.Map(db)

	err := models.NewEngine(db)

	m.Use(middleware.Contexter())
	m.Use(macaron.Recovery())
	m.Use(macaron.Logger())
	m.Use(macaron.Renderer(
		macaron.RenderOptions{
			Directory: "templates",
			TemplateFileSystem: bindata.Templates(bindata.Options{
				Asset:      templates.Asset,
				AssetDir:   templates.AssetDir,
				AssetNames: templates.AssetNames,
				AssetInfo:  templates.AssetInfo,
				Prefix:     "",
			}),
		},
	))
	m.Use(macaron.Static("web/images",
		macaron.StaticOptions{
			Prefix: "images",
			FileSystem: bindata.Static(bindata.Options{
				Asset:      web.Asset,
				AssetDir:   web.AssetDir,
				AssetNames: web.AssetNames,
				AssetInfo:  web.AssetInfo,
				Prefix:     "web/images",
			}),
		},
	))
	m.Use(macaron.Static("web/template",
		macaron.StaticOptions{
			Prefix: "template",
			FileSystem: bindata.Static(bindata.Options{
				Asset:      web.Asset,
				AssetDir:   web.AssetDir,
				AssetNames: web.AssetNames,
				AssetInfo:  web.AssetInfo,
				Prefix:     "web/template",
			}),
		},
	))

	m.Use(macaron.Static("web",
		macaron.StaticOptions{
			FileSystem: bindata.Static(bindata.Options{
				Asset:      web.Asset,
				AssetDir:   web.AssetDir,
				AssetNames: web.AssetNames,
				AssetInfo:  web.AssetInfo,
				Prefix:     "web",
			}),
			Prefix: viper.GetString("version"),
		},
	))

	m.Get("/assets/html/user/views/oauth.html", user.OauthHandler)
	m.Combo("/api/oauth").
		Get(user.OauthUrl).
		Post(binding.Json(auth.Oauth2{}), user.OauthLogin)

	m.Post("/api/login", binding.Json(auth.SignIn{}), user.SignIn)
	m.Post("/api/register", binding.Json(auth.SignUp{}), user.SignUp)

	m.Group("/api", func() {
		m.Get("/labels/:project", middleware.Datasource(), board.ListLabels)
		m.Put("/labels/:project", middleware.Datasource(), binding.Json(models.LabelRequest{}), board.EditLabel)
		m.Delete("/labels/:project/:label", middleware.Datasource(), board.DeleteLabel)
		m.Post("/labels/:project", middleware.Datasource(), binding.Json(models.LabelRequest{}), board.CreateLabel)

		m.Group("/boards", func() {
			m.Get("", board.ListBoards)
			m.Get("/starred", board.ListStarredBoards)
			m.Post("/configure", binding.Json(models.BoardRequest{}), board.Configure)

			m.Group("/:board", func() {
				m.Combo("/connect").
					Get(board.ListConnectBoard).
					Post(binding.Json(models.BoardRequest{}), board.CreateConnectBoard).
					Delete(board.DeleteConnectBoard)

				m.Post("/upload", binding.MultipartForm(models.UploadForm{}), board.UploadFile)
			})
		}, middleware.Datasource())

		m.Get("/board", middleware.Datasource(), board.ItemBoard)

		m.Get("/cards", middleware.Datasource(), board.ListCards)
		m.Combo("/milestones").
			Get(middleware.Datasource(), board.ListMilestones).
			Post(middleware.Datasource(), binding.Json(models.MilestoneRequest{}), board.CreateMilestone)

		m.Get("/users", middleware.Datasource(), board.ListMembers)
		m.Combo("/comments").
			Get(middleware.Datasource(), board.ListComments).
			Post(middleware.Datasource(), binding.Json(models.CommentRequest{}), board.CreateComment)

		m.Group("/card/:board", func() {
			m.Combo("").
				Post(binding.Json(models.CardRequest{}), board.CreateCard).
				Put(binding.Json(models.CardRequest{}), board.UpdateCard).
				Delete(binding.Json(models.CardRequest{}), board.DeleteCard)

			m.Put("/move", binding.Json(models.CardRequest{}), board.MoveToCard)
			m.Post("/move/:projectId", binding.Json(models.CardRequest{}), board.ChangeProjectForCard)

		}, middleware.Datasource())
	}, middleware.Auther())
	m.Get("/*", routers.Home)
	m.Get("/ws/", sockets.Messages(), ws.ListenAndServe)

	listen := viper.GetString("server.listen")
	log.Printf("Listen: %s", listen)
	err = http.ListenAndServe(listen, m)

	if err != nil {
		log.Fatalf("Failed to start: %s", err)
	}
}
