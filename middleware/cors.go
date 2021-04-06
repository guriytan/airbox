package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
)

type (
	// CORSConfig defines the config for CORS middleware.
	CORSConfig struct {
		// AllowOrigin defines a list of origins that may access the resource.
		// Optional. Default value []string{"*"}.
		AllowOrigins []string `yaml:"allow_origins"`

		// AllowMethods defines a list methods allowed when accessing the resource.
		// This is used in response to a preflight request.
		// Optional. Default value DefaultCORSConfig.AllowMethods.
		AllowMethods []string `yaml:"allow_methods"`
	}
)

var (
	// DefaultCORSConfig is the default CORS middleware config.
	DefaultCORSConfig = CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}
)

// CORS returns a Cross-Origin Resource Sharing (CORS) middleware.
// See: https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS
func CORS() gin.HandlerFunc {
	allowMethods := strings.Join(DefaultCORSConfig.AllowMethods, ",")
	allowOrigins := strings.Join(DefaultCORSConfig.AllowOrigins, ",")

	return func(c *gin.Context) {
		c.Writer.Header().Set(HeaderAccessControlAllowCredentials, "true")
		c.Writer.Header().Set(HeaderAccessControlAllowOrigin, allowOrigins)
		c.Writer.Header().Set(HeaderAccessControlAllowMethods, allowMethods)
		c.Writer.Header().Set(HeaderAccessControlAllowHeaders, "Origin, X-Requested-With, Content-Type, Accept, Authorization, authorization")
		c.Writer.Header().Set(HeaderAccessControlMaxAge, "3600")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
	}
}
