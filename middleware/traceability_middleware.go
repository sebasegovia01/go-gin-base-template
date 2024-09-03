package middleware

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebasegovia01/base-template-go-gin/errors"
)

var requiredHeaders = []string{
	"Consumer-Sys-Code",
	"Consumer-Enterprise-Code",
	"Consumer-Country-Code",
	"Trace-Client-Req-Timestamp",
	"Trace-Source-Id",
	"Channel-Name",
	"Channel-Mode",
}

var (
	timestampRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{6}[+-]\d{4}$`)
	uuidRegex      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

var validEnterpriseValues = map[string]bool{
	"BANCORIPLEY-CHL": true,
	"BANCORIPLEY-PER": true,
}

var validCountryValues = map[string]bool{
	"CHL": true,
	"PER": true,
}

var validChannelModes = map[string]bool{
	"PRESENCIAL":    true,
	"NO-PRESENCIAL": true,
}

var channelCodes = map[string]string{
	"CHL-SIT-SEG":   "SEGUROS",
	"CHL-HB-WEB":    "PWA",
	"CHL-HB-MOB":    "PWA",
	"CHL-HB-APP":    "PWA",
	"CHL-APP-MOB":   "APP",
	"CHL-TIE-VIR":   "TIENDA-VIRTUAL",
	"CHL-PUB-WEB":   "SITIO-PUBLICO",
	"CHL-PUB-MOB":   "SITIO-PUBLICO",
	"CHL-SCP-WEB":   "SCP",
	"CHL-PTF-WEB":   "PORTABILIDAD FINANCIERA",
	"CHL-REC-EXT":   "RECAUDADORES EXTERNOS",
	"CHL-PWA-TOTEM": "PWA",
	"CHL-CO-WEB":    "CAPTACION-ONLINE",
	"CHL-DG-WEB":    "CHEK",
	"CHL-PUBCK-WEB": "SITIO-PUBLICO-CHEK",
	"CHL-CK-MOB":    "CHEK MOBILE",
	"CHL-CKCOM-WEB": "CHECK COMERCIOS",
	"CHL-SLCK-BCK":  "SHERLOCK CHECK",
}

func TraceabilityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		missingHeaders := []string{}
		invalidHeaders := make(map[string]string)

		for _, header := range requiredHeaders {
			value := c.GetHeader(header)
			if value == "" {
				missingHeaders = append(missingHeaders, header)
			} else {
				switch header {
				case "Trace-Client-Req-Timestamp":
					if !timestampRegex.MatchString(value) {
						invalidHeaders[header] = "Invalid timestamp format. Expected: yyyy-MM-dd HH:mm:ss.SSSSSSZ"
					}
				case "Trace-Source-Id":
					if !uuidRegex.MatchString(value) {
						invalidHeaders[header] = "Invalid UUID format"
					}
				case "Consumer-Enterprise-Code":
					if !validEnterpriseValues[value] {
						invalidHeaders[header] = "Invalid value. Expected: BANCORIPLEY-CHL or BANCORIPLEY-PER"
					}
				case "Consumer-Country-Code":
					if !validCountryValues[value] {
						invalidHeaders[header] = "Invalid value. Expected: CHL or PER"
					}
				case "Channel-Mode":
					if !validChannelModes[value] {
						invalidHeaders[header] = "Invalid value. Expected: PRESENCIAL or NO-PRESENCIAL"
					}
				}
			}
		}

		// Validate consistency between Consumer-Sys-Code and Channel-Name
		consumerSysCode := c.GetHeader("Consumer-Sys-Code")
		channelName := c.GetHeader("Channel-Name")
		if expectedChannelName, exists := channelCodes[consumerSysCode]; exists {
			if channelName != expectedChannelName {
				invalidHeaders["Channel-Name"] = fmt.Sprintf("Inconsistent with Consumer-Sys-Code. Expected: %s", expectedChannelName)
			}
		} else {
			invalidHeaders["Consumer-Sys-Code"] = "Invalid value"
		}

		if len(missingHeaders) > 0 || len(invalidHeaders) > 0 {
			errorMsg := ""
			if len(missingHeaders) > 0 {
				errorMsg += "Missing required headers: " + strings.Join(missingHeaders, ", ")
			}
			if len(invalidHeaders) > 0 {
				if errorMsg != "" {
					errorMsg += ". "
				}
				errorMsg += "Invalid headers: "
				for header, reason := range invalidHeaders {
					errorMsg += fmt.Sprintf("%s (%s), ", header, reason)
				}
				errorMsg = strings.TrimSuffix(errorMsg, ", ")
			}
			c.Error(errors.NewCustomError(400, errorMsg))
			c.Abort()
			return
		}

		// Set response headers
		now := time.Now().Format("2006-01-02 15:04:05.000000-0700")
		c.Header("Trace-Req-Timestamp", now)
		c.Header("Trace-Source-Id", c.GetHeader("Trace-Source-Id"))
		c.Header("Local-Transaction-Id", uuid.New().String())

		c.Next()

		// Set response timestamp after processing
		c.Header("Trace-Rsp-Timestamp", time.Now().Format("2006-01-02 15:04:05.000000-0700"))
	}
}
