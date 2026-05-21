package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCases holds the test configuration and expected results.
type TestCase struct {
	name            string
	config          *Config
	method          string
	requestHeaders  map[string]string
	expectedCode    int
	expectedHeaders map[string]string
}

func setupRouter(config *Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Middleware(config))
	router.Any("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	return router
}

func runTestCase(t *testing.T, tc TestCase) {
	router := setupRouter(tc.config)
	w := httptest.NewRecorder()
	req, err := http.NewRequest(tc.method, "/test", nil)
	require.NoError(t, err)

	// Set request headers
	for key, value := range tc.requestHeaders {
		req.Header.Set(key, value)
	}

	router.ServeHTTP(w, req)

	// Check status code
	assert.Equal(t, tc.expectedCode, w.Code, "Status code mismatch for test: %s", tc.name)

	// Check response headers
	for key, expectedValue := range tc.expectedHeaders {
		assert.Equal(t, expectedValue, w.Header().Get(key),
			"Header %s mismatch for test: %s", key, tc.name)
	}
}

func TestCorsMiddleware_DefaultConfig(t *testing.T) {
	tc := TestCase{
		name:   "Default configuration",
		config: DefaultConfig(),
		method: "GET",
		requestHeaders: map[string]string{
			"Origin": "http://example.com",
		},
		expectedCode: http.StatusOK,
		expectedHeaders: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "false",
			"Access-Control-Expose-Headers":    "Content-Length",
			"Vary":                             "Origin",
		},
	}
	runTestCase(t, tc)
}

func TestCorsMiddleware_MultipleOrigins(t *testing.T) {
	config := &Config{
		AllowedOrigins:   []string{"http://example1.com", "http://example2.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           10 * time.Minute,
	}

	testCases := []TestCase{
		{
			name:   "Allowed origin 1",
			config: config,
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://example1.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example1.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:   "Allowed origin 2",
			config: config,
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://example2.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example2.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:   "Disallowed origin",
			config: config,
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://example3.com",
			},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

func TestCorsMiddleware_PreflightRequests(t *testing.T) {
	config := &Config{
		AllowedOrigins:   []string{"http://example.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           10 * time.Minute,
	}

	testCases := []TestCase{
		{
			name:   "Valid preflight request",
			config: config,
			method: "OPTIONS",
			requestHeaders: map[string]string{
				"Origin":                         "http://example.com",
				"Access-Control-Request-Method":  "POST",
				"Access-Control-Request-Headers": "Authorization",
			},
			expectedCode: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Methods":     "GET, POST, PUT",
				"Access-Control-Allow-Headers":     "Authorization, Content-Type",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "600",
			},
		},
		{
			name:   "Preflight with disallowed origin",
			config: config,
			method: "OPTIONS",
			requestHeaders: map[string]string{
				"Origin":                        "http://unauthorized.com",
				"Access-Control-Request-Method": "POST",
			},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

func TestCorsMiddleware_WildcardValidation(t *testing.T) {
	assert.Panics(t, func() {
		config := &Config{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
		}
		config.validate()
	}, "Should panic when using wildcard with credentials")

	assert.NotPanics(t, func() {
		config := &Config{
			AllowedOrigins:   []string{"http://example.com"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"ContentType", "ContentLength"},
			ExposeHeaders:    []string{"ContentType"},
			AllowCredentials: true,
		}
		config.validate()
	}, "Should not panic with specific origin and credentials")
}

func TestCorsMiddleware_HeaderValidation(t *testing.T) {
	testCases := []struct {
		name           string
		config         *Config
		shouldPanic    bool
		expectedConfig *Config
	}{
		{
			name: "Default values with no credentials",
			config: &Config{
				AllowCredentials: false,
			},
			shouldPanic: false,
			expectedConfig: &Config{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"*"},
				AllowedHeaders:   []string{"*"},
				ExposeHeaders:    []string{"*"},
				AllowCredentials: false,
				MaxAge:           24 * time.Hour,
			},
		},
		{
			name: "Custom values with credentials",
			config: &Config{
				AllowedOrigins:   []string{"http://example.com"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"Authorization"},
				ExposeHeaders:    []string{"Content-Length"},
				AllowCredentials: true,
				MaxAge:           10 * time.Minute,
			},
			shouldPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				assert.Panics(t, func() {
					tc.config.validate()
				})
			} else {
				assert.NotPanics(t, func() {
					validatedConfig := tc.config.validate()
					if tc.expectedConfig != nil && tc.expectedConfig.AllowedOrigins != nil {
						assert.Equal(
							t,
							tc.expectedConfig.AllowedOrigins,
							validatedConfig.AllowedOrigins,
						)
					}
				})
			}
		})
	}
}

func TestCorsMiddleware_MethodValidation(t *testing.T) {
	config := &Config{
		AllowedOrigins:   []string{"http://example.com"},
		AllowedMethods:   []string{"GET", "POST"}, // Only GET and POST are allowed
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           10 * time.Minute,
	}

	testCases := []TestCase{
		{
			name:   "Allowed method - GET",
			config: config,
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://example.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:   "Allowed method - POST",
			config: config,
			method: "POST",
			requestHeaders: map[string]string{
				"Origin": "http://example.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:   "Disallowed method - PUT",
			config: config,
			method: "PUT",
			requestHeaders: map[string]string{
				"Origin": "http://example.com",
			},
			expectedCode: http.StatusMethodNotAllowed,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:   "Disallowed method - DELETE",
			config: config,
			method: "DELETE",
			requestHeaders: map[string]string{
				"Origin": "http://example.com",
			},
			expectedCode: http.StatusMethodNotAllowed,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:   "Preflight with allowed method",
			config: config,
			method: "OPTIONS",
			requestHeaders: map[string]string{
				"Origin":                        "http://example.com",
				"Access-Control-Request-Method": "POST",
			},
			expectedCode: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Methods":     "GET, POST",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "600",
			},
		},
		{
			name: "Wildcard methods test",
			config: &Config{
				AllowedOrigins:   []string{"http://example.com"},
				AllowedMethods:   []string{"*"},
				AllowCredentials: false,
			},
			method: "PUT", // Should be allowed because of wildcard
			requestHeaders: map[string]string{
				"Origin": "http://example.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Credentials": "false",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

func TestCorsMiddleware_MethodValidationEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		config      *Config
		shouldPanic bool
	}{
		{
			name: "Empty methods list",
			config: &Config{
				AllowedOrigins:   []string{"http://example.com"},
				AllowedMethods:   []string{},
				AllowCredentials: true,
			},
			shouldPanic: true,
		},
		{
			name: "Invalid HTTP method",
			config: &Config{
				AllowedOrigins:   []string{"http://example.com"},
				AllowedMethods:   []string{"GET", "INVALID_METHOD"},
				AllowCredentials: true,
			},
			shouldPanic: true,
		},
		{
			name: "Duplicate methods",
			config: &Config{
				AllowedOrigins:   []string{"http://example.com"},
				AllowedMethods:   []string{"GET", "GET", "POST"},
				AllowCredentials: true,
			},
			shouldPanic: true,
		},
		{
			name: "Case sensitivity test",
			config: &Config{
				AllowedOrigins:   []string{"http://example.com"},
				AllowedMethods:   []string{"get", "POST"},
				AllowCredentials: true,
			},
			shouldPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				assert.Panics(t, func() {
					tc.config.validate()
				})
			} else {
				assert.NotPanics(t, func() {
					tc.config.validate()
				})
			}
		})
	}
}

func TestCorsMiddleware_AllowedOriginsFunc(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Function allows origin and enables credentials",
			config: &Config{
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					// Allow specific origin with credentials
					return origin == "http://trusted.com", origin == "http://trusted.com"
				},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Authorization"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://trusted.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://trusted.com",
				"Access-Control-Allow-Credentials": "true",
				"Vary":                             "Origin",
			},
		},
		{
			name: "Function allows origin but disables credentials",
			config: &Config{
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					// Allow origin but no credentials
					return origin == "http://public.com", false
				},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://public.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://public.com",
				"Access-Control-Allow-Credentials": "false",
				"Vary":                             "Origin",
			},
		},
		{
			name: "Function denies origin",
			config: &Config{
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					// Deny all origins except trusted.com
					return origin == "http://trusted.com", true
				},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Authorization"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://untrusted.com",
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "Function with pattern matching - subdomain allowed",
			config: &Config{
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					// Allow all subdomains of example.com
					allowed := origin == "http://app.example.com" ||
						origin == "http://api.example.com" ||
						origin == "http://www.example.com"
					return allowed, allowed
				},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Authorization"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://api.example.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://api.example.com",
				"Access-Control-Allow-Credentials": "true",
				"Vary":                             "Origin",
			},
		},
		{
			name: "Function with pattern matching - subdomain denied",
			config: &Config{
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					// Allow all subdomains of example.com
					allowed := origin == "http://app.example.com" ||
						origin == "http://api.example.com" ||
						origin == "http://www.example.com"
					return allowed, allowed
				},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Authorization"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://malicious.com",
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "Function allows all origins with credentials",
			config: &Config{
				AllowedOriginsFunc: func(_ string) (bool, bool) {
					// Allow any origin with credentials
					return true, true
				},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Authorization"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://any-origin.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://any-origin.com",
				"Access-Control-Allow-Credentials": "true",
				"Vary":                             "Origin",
			},
		},
		{
			name: "Function allows all origins without credentials",
			config: &Config{
				AllowedOriginsFunc: func(_ string) (bool, bool) {
					// Allow any origin without credentials
					return true, false
				},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://random-origin.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://random-origin.com",
				"Access-Control-Allow-Credentials": "false",
				"Vary":                             "Origin",
			},
		},
		{
			name: "Function with preflight request - allowed",
			config: &Config{
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					return origin == "http://trusted.com", true
				},
				AllowedMethods: []string{"GET", "POST", "PUT"},
				AllowedHeaders: []string{"Authorization", "Content-Type"},
				ExposeHeaders:  []string{"Content-Length"},
				MaxAge:         5 * time.Minute,
			},
			method: "OPTIONS",
			requestHeaders: map[string]string{
				"Origin":                        "http://trusted.com",
				"Access-Control-Request-Method": "POST",
			},
			expectedCode: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://trusted.com",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Allow-Methods":     "GET, POST, PUT",
				"Access-Control-Allow-Headers":     "Authorization, Content-Type",
				"Access-Control-Max-Age":           "300",
				"Vary":                             "Origin",
			},
		},
		{
			name: "Function with preflight request - denied",
			config: &Config{
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					return origin == "http://trusted.com", true
				},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Authorization"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "OPTIONS",
			requestHeaders: map[string]string{
				"Origin":                        "http://untrusted.com",
				"Access-Control-Request-Method": "POST",
			},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

func TestCorsMiddleware_AllowedOriginsFunc_IgnoresConfigProperties(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Function ignores AllowedOrigins slice",
			config: &Config{
				// These origins should be ignored
				AllowedOrigins: []string{"http://should-be-ignored.com"},
				// Function has different logic
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					return origin == "http://function-allowed.com", true
				},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://function-allowed.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://function-allowed.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name: "Function ignores AllowedOrigins - origin from slice denied",
			config: &Config{
				// This origin is in the slice but should be denied by function
				AllowedOrigins: []string{"http://in-slice.com"},
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					// Only allow different origin
					return origin == "http://function-allowed.com", false
				},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://in-slice.com",
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "Function ignores AllowCredentials config property",
			config: &Config{
				// AllowCredentials is true but function returns false for credentials
				AllowCredentials: true,
				AllowedOrigins:   []string{"http://example.com"},
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					// Allow origin but disable credentials (second return value)
					return origin == "http://example.com", false
				},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Authorization"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://example.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Credentials": "false",
			},
		},
		{
			name: "Function enables credentials despite config being false",
			config: &Config{
				// AllowCredentials is false but function returns true for credentials
				AllowCredentials: false,
				AllowedOrigins:   []string{"http://example.com"},
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					// Allow origin and enable credentials (second return value)
					return origin == "http://example.com", true
				},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Authorization"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://example.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name: "Function ignores wildcard in AllowedOrigins",
			config: &Config{
				// Wildcard should be ignored
				AllowedOrigins: []string{"*"},
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					// Only specific origin allowed
					return origin == "http://specific.com", true
				},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://random.com",
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "Function allows origin that is not in AllowedOrigins slice",
			config: &Config{
				AllowedOrigins: []string{"http://origin1.com", "http://origin2.com"},
				AllowedOriginsFunc: func(origin string) (bool, bool) {
					// Allow origin that is NOT in the AllowedOrigins slice
					return origin == "http://origin3.com", true
				},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
				ExposeHeaders:  []string{"Content-Length"},
			},
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://origin3.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://origin3.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

func TestCorsMiddleware_AllowedOriginsFunc_DynamicBehavior(t *testing.T) {
	// Test that the function can implement complex dynamic logic

	t.Run("Environment-based origin validation", func(t *testing.T) {
		// Simulate checking origins based on environment
		allowedEnvOrigins := map[string]bool{
			"http://prod.example.com":    true,
			"http://staging.example.com": true,
		}

		config := &Config{
			AllowedOriginsFunc: func(origin string) (bool, bool) {
				allowed := allowedEnvOrigins[origin]
				return allowed, allowed
			},
			AllowedMethods: []string{"GET", "POST"},
			AllowedHeaders: []string{"Authorization"},
			ExposeHeaders:  []string{"Content-Length"},
		}

		// Test allowed origin
		tc1 := TestCase{
			config: config,
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://prod.example.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://prod.example.com",
				"Access-Control-Allow-Credentials": "true",
			},
		}
		runTestCase(t, tc1)

		// Test denied origin
		tc2 := TestCase{
			config: config,
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://dev.example.com",
			},
			expectedCode: http.StatusForbidden,
		}
		runTestCase(t, tc2)
	})

	t.Run("Conditional credentials based on origin", func(t *testing.T) {
		// Different credentials settings for different origins
		config := &Config{
			AllowedOriginsFunc: func(origin string) (bool, bool) {
				// Trusted origins get credentials, public origins don't
				if origin == "http://trusted.com" {
					return true, true
				}
				if origin == "http://public.com" {
					return true, false
				}
				return false, false
			},
			AllowedMethods: []string{"GET", "POST"},
			AllowedHeaders: []string{"Authorization"},
			ExposeHeaders:  []string{"Content-Length"},
		}

		// Test trusted origin with credentials
		tc1 := TestCase{
			config: config,
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://trusted.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://trusted.com",
				"Access-Control-Allow-Credentials": "true",
			},
		}
		runTestCase(t, tc1)

		// Test public origin without credentials
		tc2 := TestCase{
			config: config,
			method: "GET",
			requestHeaders: map[string]string{
				"Origin": "http://public.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://public.com",
				"Access-Control-Allow-Credentials": "false",
			},
		}
		runTestCase(t, tc2)
	})
}

func TestCorsMiddleware_NoOrigin(t *testing.T) {
	t.Run("Request without Origin header skips CORS handling", func(t *testing.T) {
		config := &Config{
			AllowedOrigins:   []string{"http://example.com"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}

		router := setupRouter(config)
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err)

		// Do NOT set Origin header
		// This should skip CORS processing and just handle the request normally

		router.ServeHTTP(w, req)

		// Should succeed with OK status
		assert.Equal(t, http.StatusOK, w.Code)

		// CORS headers should NOT be set (except Vary which is always set)
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Expose-Headers"))

		// Vary header should still be set
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})

	t.Run("Request without Origin header with AllowedOriginsFunc", func(t *testing.T) {
		config := &Config{
			AllowedOriginsFunc: func(_ string) (bool, bool) {
				// This should not be called when no origin is present
				t.Error("AllowedOriginsFunc should not be called when no Origin header is present")
				return false, false
			},
			AllowedMethods: []string{"GET", "POST"},
			AllowedHeaders: []string{"Authorization"},
			ExposeHeaders:  []string{"Content-Length"},
		}

		router := setupRouter(config)
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err)

		// Do NOT set Origin header

		router.ServeHTTP(w, req)

		// Should succeed without calling the function
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
	})
}
