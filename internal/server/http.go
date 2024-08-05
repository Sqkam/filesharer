package server

import (
	v1 "filesharer/api/file/v1"
	"filesharer/internal/conf"
	"filesharer/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	transhttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.FileService, logger log.Logger) *transhttp.Server {
	var opts = []transhttp.ServerOption{

		transhttp.Filter(handlers.CORS(
			//handlers.AllowedOrigins([]string{"*","http://localhost:5173"}),
			//handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}),
			//handlers.AllowedHeaders( []string{"Origin", "Content-Length", "Content-Type"}),

			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST"}),
			handlers.AllowedHeaders([]string{"Origin", "Content-Length", "Content-Type", "Authorization", "Host", "Date"}),

			//handlers.AllowCredentials(),
		)),
		transhttp.Middleware(
			recovery.Recovery(),
			validate.Validator(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, transhttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, transhttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, transhttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := transhttp.NewServer(opts...)
	v1.RegisterFileHTTPServer(srv, greeter)
	return srv
}
