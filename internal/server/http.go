package server

import (
	v1 "filesharer/api/file/v1"
	"filesharer/internal/conf"
	"filesharer/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/gorilla/handlers"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.FileService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{

		http.Filter(handlers.CORS(
			//handlers.AllowedOrigins([]string{"*","http://localhost:5173"}),
			//handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}),
			//handlers.AllowedHeaders( []string{"Origin", "Content-Length", "Content-Type"}),

			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST"}),
			handlers.AllowedHeaders([]string{"Origin", "Content-Length", "Content-Type", "Authorization", "Host", "Date"}),

			//handlers.AllowCredentials(),
		)),
		http.Middleware(
			recovery.Recovery(),
			validate.Validator(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterFileHTTPServer(srv, greeter)
	return srv
}
