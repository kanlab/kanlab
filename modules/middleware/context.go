package middleware

import (
	"encoding/json"
	"github.com/kanlab/kanlab/models"
	"github.com/kanlab/kanlab/ws"
	"github.com/kanlab/kanlab/datasource"
	"gopkg.in/macaron.v1"
)

type Context struct {
	*macaron.Context
	User        *models.User
	IsAdmin     bool
	IsSigned    bool
	IsBasicAuth bool

	DataSource datasource.DataSource

	Provider string
}

// Contexter initializes a classic context for a request.
func Contexter() macaron.Handler {
	return func(c *macaron.Context) {
		ctx := &Context{
			Context: c,
		}

		ctx.Provider = "gitlab"

		c.Map(ctx)
	}
}

// Broadcast sends message via WebSocket to all subscribed to r users
func (*Context) Broadcast(r string, d interface{}) {
	res, _ := json.Marshal(d)
	go ws.Server(r).Broadcast(string(res))
}
