package middleware

import (
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/response"
	"github.com/gin-gonic/gin"
)

// MTLSMiddleware verifies that the request includes a valid client certificate.
// The certificate chain verification is handled by the server's TLS config
// (tls.VerifyClientCertIfGiven with the CA cert pool).
// This middleware ensures a client cert was actually provided for the protected route.
func MTLSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS == nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("TLS connection required")))
			c.Abort()
			return
		}

		if len(c.Request.TLS.PeerCertificates) == 0 {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Client certificate required")))
			c.Abort()
			return
		}

		clientCert := c.Request.TLS.PeerCertificates[0]
		c.Set("clientCN", clientCert.Subject.CommonName)

		c.Next()
	}
}
