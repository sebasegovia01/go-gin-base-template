package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebasegovia01/base-template-go-gin/errors"
	"github.com/sebasegovia01/base-template-go-gin/middleware"
	"github.com/stretchr/testify/assert"
)

func TestTraceabilityMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid headers",
			headers: map[string]string{
				"Consumer-Sys-Code":          "CHL-HB-WEB",
				"Consumer-Enterprise-Code":   "BANCORIPLEY-CHL",
				"Consumer-Country-Code":      "CHL",
				"Trace-Client-Req-Timestamp": time.Now().Format("2006-01-02 15:04:05.000000-0700"),
				"Trace-Source-Id":            uuid.New().String(),
				"Channel-Name":               "PWA",
				"Channel-Mode":               "PRESENCIAL",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Missing headers",
			headers: map[string]string{
				"Consumer-Sys-Code": "CHL-HB-WEB",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing required headers: Consumer-Enterprise-Code, Consumer-Country-Code, Trace-Client-Req-Timestamp, Trace-Source-Id, Channel-Name, Channel-Mode",
		},
		{
			name: "Invalid timestamp format",
			headers: map[string]string{
				"Consumer-Sys-Code":          "CHL-HB-WEB",
				"Consumer-Enterprise-Code":   "BANCORIPLEY-CHL",
				"Consumer-Country-Code":      "CHL",
				"Trace-Client-Req-Timestamp": "invalid-timestamp",
				"Trace-Source-Id":            uuid.New().String(),
				"Channel-Name":               "PWA",
				"Channel-Mode":               "PRESENCIAL",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid headers: Trace-Client-Req-Timestamp (Invalid timestamp format. Expected: yyyy-MM-dd HH:mm:ss.SSSSSSZ)",
		},
		{
			name: "Invalid UUID format",
			headers: map[string]string{
				"Consumer-Sys-Code":          "CHL-HB-WEB",
				"Consumer-Enterprise-Code":   "BANCORIPLEY-CHL",
				"Consumer-Country-Code":      "CHL",
				"Trace-Client-Req-Timestamp": time.Now().Format("2006-01-02 15:04:05.000000-0700"),
				"Trace-Source-Id":            "invalid-uuid",
				"Channel-Name":               "PWA",
				"Channel-Mode":               "PRESENCIAL",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid headers: Trace-Source-Id (Invalid UUID format)",
		},
		{
			name: "Invalid Enterprise Code",
			headers: map[string]string{
				"Consumer-Sys-Code":          "CHL-HB-WEB",
				"Consumer-Enterprise-Code":   "INVALID",
				"Consumer-Country-Code":      "CHL",
				"Trace-Client-Req-Timestamp": time.Now().Format("2006-01-02 15:04:05.000000-0700"),
				"Trace-Source-Id":            uuid.New().String(),
				"Channel-Name":               "PWA",
				"Channel-Mode":               "PRESENCIAL",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid headers: Consumer-Enterprise-Code (Invalid value. Expected: BANCORIPLEY-CHL or BANCORIPLEY-PER)",
		},
		{
			name: "Invalid Country Code",
			headers: map[string]string{
				"Consumer-Sys-Code":          "CHL-HB-WEB",
				"Consumer-Enterprise-Code":   "BANCORIPLEY-CHL",
				"Consumer-Country-Code":      "INVALID",
				"Trace-Client-Req-Timestamp": time.Now().Format("2006-01-02 15:04:05.000000-0700"),
				"Trace-Source-Id":            uuid.New().String(),
				"Channel-Name":               "PWA",
				"Channel-Mode":               "PRESENCIAL",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid headers: Consumer-Country-Code (Invalid value. Expected: CHL or PER)",
		},
		{
			name: "Invalid Channel Mode",
			headers: map[string]string{
				"Consumer-Sys-Code":          "CHL-HB-WEB",
				"Consumer-Enterprise-Code":   "BANCORIPLEY-CHL",
				"Consumer-Country-Code":      "CHL",
				"Trace-Client-Req-Timestamp": time.Now().Format("2006-01-02 15:04:05.000000-0700"),
				"Trace-Source-Id":            uuid.New().String(),
				"Channel-Name":               "PWA",
				"Channel-Mode":               "INVALID",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid headers: Channel-Mode (Invalid value. Expected: PRESENCIAL or NO-PRESENCIAL)",
		},
		{
			name: "Inconsistent Channel Name",
			headers: map[string]string{
				"Consumer-Sys-Code":          "CHL-HB-WEB",
				"Consumer-Enterprise-Code":   "BANCORIPLEY-CHL",
				"Consumer-Country-Code":      "CHL",
				"Trace-Client-Req-Timestamp": time.Now().Format("2006-01-02 15:04:05.000000-0700"),
				"Trace-Source-Id":            uuid.New().String(),
				"Channel-Name":               "INVALID",
				"Channel-Mode":               "PRESENCIAL",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid headers: Channel-Name (Inconsistent with Consumer-Sys-Code. Expected: PWA)",
		},
		{
			name: "Invalid Consumer-Sys-Code",
			headers: map[string]string{
				"Consumer-Sys-Code":          "INVALID",
				"Consumer-Enterprise-Code":   "BANCORIPLEY-CHL",
				"Consumer-Country-Code":      "CHL",
				"Trace-Client-Req-Timestamp": time.Now().Format("2006-01-02 15:04:05.000000-0700"),
				"Trace-Source-Id":            uuid.New().String(),
				"Channel-Name":               "PWA",
				"Channel-Mode":               "PRESENCIAL",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid headers: Consumer-Sys-Code (Invalid value)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			// Simular WithTraceability
			router.Use(func(c *gin.Context) {
				middleware.TraceabilityMiddleware()(c)
				if c.IsAborted() {
					// Manejar CustomError
					if len(c.Errors) > 0 {
						err := c.Errors.Last()
						if customErr, ok := err.Err.(*errors.CustomError); ok {
							c.JSON(customErr.StatusCode, gin.H{"error": customErr.Message})
						}
					}
					return
				}
				c.Next()
			})

			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			if tt.expectedStatus == http.StatusOK {
				assert.NotEmpty(t, w.Header().Get("Trace-Req-Timestamp"))
				assert.NotEmpty(t, w.Header().Get("Trace-Source-Id"))
				assert.NotEmpty(t, w.Header().Get("Local-Transaction-Id"))
				assert.NotEmpty(t, w.Header().Get("Trace-Rsp-Timestamp"))
			}
		})
	}
}
