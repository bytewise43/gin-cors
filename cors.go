// Package cors provides a CORS middleware for the Gin web framework.
package cors

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Added this test comment will not be merged

// Config is the configuration for the cors middleware.
type Config struct {
	// All the allowed origins in an array. The default is "*".
	// The default cannot be used when AllowCredentials is true.
	// This array will have no effect if the AllowedOriginsFunc is used.
	// [MDN Web Docs]
	//
	// [MDN Web Docs]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	AllowedOrigins []string
	// A function to dynamically check if an origin is allowed.
	// The first return value indicates if the origin is allowed.
	// The second determines if the AllowCredentials header should be set.
	//
	// This can be used for dynamic client evaluation instead of using the AllowedOrigins slice.
	AllowedOriginsFunc func(origin string) (bool, bool)
	// All the allowed HTTP Methodes. The default is "*"
	// The default cannot be used when AllowCredentials is true
	// [MDN Web Docs]
	//
	// [MDN Web Docs]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
	AllowedMethods []string
	// All the allowed Headers that can be sent from the client. The default is "*"
	// The default cannot be used when AllowCredentials is true
	// Note that the Authorization header cannot be wildcarded and needs to be listed explicitly [MDN Web Docs]
	//
	// [MDN Web Docs]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
	AllowedHeaders []string
	// The headers which should be readable by the client. The default is "*"
	// The default cannot be used when AllowCredentials is true
	// [MDN Web Docs]
	//
	// [MDN Web Docs]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Headers
	ExposeHeaders []string
	// If you allow receiving cookies and Authorization headers. The default is false.
	// This will have no effect if the AllowedOriginsFunc is used.
	// [MDN Web Docs]
	//
	// [MDN Web Docs]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
	AllowCredentials bool
	// The maximum age of your preflight requests. The default is 1 day
	// [MDN Web Docs]
	//
	// [MDN Web Docs]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age
	MaxAge time.Duration
}

// checkCredentials adds wildcard to the array if no value was set.
// Panics if the allowCredentials is true and the header is a wildcard.
func checkCredentials(header []string, allowCredentials bool, headerName string) []string {
	if len(header) <= 0 {
		if allowCredentials {
			panic(fmt.Sprintf("The %s must be set when AllowCredentials is true", headerName))
		}
		return []string{"*"}
	}

	if slices.Contains(header, "*") && allowCredentials {
		panic(
			fmt.Sprintf(
				"The %s cannot contain the \"*\" wildcard when AllowCredentials is true",
				headerName,
			),
		)
	}

	return header
}

func (c *Config) validate() *Config {
	c.AllowedOrigins = checkCredentials(c.AllowedOrigins, c.AllowCredentials, "allowed origins")

	c.AllowedMethods = checkCredentials(c.AllowedMethods, c.AllowCredentials, "allowed methods")

	c.AllowedHeaders = checkCredentials(c.AllowedHeaders, c.AllowCredentials, "allowed headers")

	c.ExposeHeaders = checkCredentials(c.ExposeHeaders, c.AllowCredentials, "expose headers")

	for i, method := range c.AllowedMethods {
		c.AllowedMethods[i] = strings.ToUpper(method)
	}

	c.AllowedMethods = slices.Compact(c.AllowedMethods)

	if c.MaxAge == 0 {
		c.MaxAge = 24 * time.Hour
	}

	return c
}

// DefaultConfig returns a default cors config.
func DefaultConfig() *Config {
	return &Config{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"*"},
		AllowedHeaders:   []string{"Content-Type", "Content-Length"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           24 * time.Hour,
	}
}

// Middleware returns a CORS Middleware for Gin which handles CORS headers and preflight requests.
// Needs a cors config.
func Middleware(config *Config) gin.HandlerFunc {
	config = config.validate()

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		c.Writer.Header().Set("Vary", "Origin")

		// If no origin is set we skip the CORS handling
		if origin == "" {
			c.Next()
			return
		}

		// Check origin and determine credentials based on config or function
		var allowCredentials bool

		if config.AllowedOriginsFunc != nil {
			// Use dynamic function for origin validation
			allowed, funcAllowCredentials := config.AllowedOriginsFunc(origin)
			if !allowed {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			allowCredentials = funcAllowCredentials
		} else {
			// Use static origin list for validation
			if !slices.Contains(config.AllowedOrigins, "*") &&
				!slices.Contains(config.AllowedOrigins, origin) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			// Use wildcard origin if configured
			if slices.Contains(config.AllowedOrigins, "*") {
				origin = "*"
			}

			allowCredentials = config.AllowCredentials
		}

		// Check if method is allowed
		method := strings.ToUpper(c.Request.Method)
		methodAllowed := slices.Contains(config.AllowedMethods, "*") ||
			slices.Contains(config.AllowedMethods, method) ||
			method == "OPTIONS"

		// Check if this is a preflight request
		isPreflight := method == "OPTIONS"

		// Set CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().
			Set("Access-Control-Allow-Credentials", strconv.FormatBool(allowCredentials))
		c.Writer.Header().
			Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))

		if isPreflight {
			// Additional headers for preflight requests
			c.Writer.Header().
				Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			c.Writer.Header().
				Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			c.Writer.Header().
				Set("Access-Control-Max-Age", strconv.FormatInt(int64(config.MaxAge.Seconds()), 10))
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		if !methodAllowed {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}

		c.Next()
	}
}
