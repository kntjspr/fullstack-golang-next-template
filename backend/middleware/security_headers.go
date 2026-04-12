package middleware

import "net/http"

const (
	headerXContentTypeOptions = "X-Content-Type-Options"
	headerXFrameOptions       = "X-Frame-Options"
	headerXXSSProtection      = "X-XSS-Protection"
	headerReferrerPolicy      = "Referrer-Policy"
	headerCSP                 = "Content-Security-Policy"
	headerPermissionsPolicy   = "Permissions-Policy"
	headerHSTS                = "Strict-Transport-Security"
)

const (
	valueXContentTypeOptions = "nosniff"
	valueXFrameOptions       = "DENY"
	valueXXSSProtection      = "1; mode=block"
	valueReferrerPolicy      = "strict-origin-when-cross-origin"
	valueCSP                 = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'"
	valuePermissionsPolicy   = "camera=(), microphone=(), geolocation=()"
	valueHSTS                = "max-age=31536000; includeSubDomains"
)

// SecurityHeaders applies OWASP-recommended security headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set(headerXContentTypeOptions, valueXContentTypeOptions)
		headers.Set(headerXFrameOptions, valueXFrameOptions)
		headers.Set(headerXXSSProtection, valueXXSSProtection)
		headers.Set(headerReferrerPolicy, valueReferrerPolicy)
		headers.Set(headerCSP, valueCSP)
		headers.Set(headerPermissionsPolicy, valuePermissionsPolicy)

		if r.TLS != nil {
			headers.Set(headerHSTS, valueHSTS)
		}

		next.ServeHTTP(w, r)
	})
}
