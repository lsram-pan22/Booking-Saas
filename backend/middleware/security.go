package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
)

const MaxRequestBody int64 = 1 << 20

func SafeRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("request_panic route=%s request_id=%s type=%T", c.FullPath(), c.Writer.Header().Get("X-Request-ID"), recovered)
				c.AbortWithStatusJSON(500, gin.H{"success": false, "message": "ระบบไม่สามารถดำเนินการได้ในขณะนี้"})
			}
		}()
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000")
		}
		c.Next()
	}
}

func LimitRequestBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > MaxRequestBody {
			c.AbortWithStatusJSON(413, gin.H{"success": false, "message": "ข้อมูลคำขอมีขนาดใหญ่เกินไป"})
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxRequestBody)
		}
		c.Next()
	}
}

// Keep validation equivalent to Gin binding, but reject unknown fields and
// trailing JSON. Do not expose decoder errors or submitted credentials.
func DecodeJSON(c *gin.Context, dest any) error {
	media, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || !(media == "application/json" || (strings.HasPrefix(media, "application/") && strings.HasSuffix(media, "+json"))) {
		return errors.New("JSON content type required")
	}
	decoder := json.NewDecoder(c.Request.Body)
	var object json.RawMessage
	if err = decoder.Decode(&object); err != nil {
		return err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return errors.New("exactly one JSON object required")
	}
	object = bytes.TrimSpace(object)
	if len(object) == 0 || object[0] != '{' {
		return errors.New("JSON object required")
	}
	strict := json.NewDecoder(bytes.NewReader(object))
	strict.DisallowUnknownFields()
	if err = strict.Decode(dest); err != nil {
		return err
	}
	return binding.Validator.ValidateStruct(dest)
}

// Audit only route patterns and verified actor IDs. Never record bodies,
// passwords, tokens, search parameters or arbitrary client request IDs.
func SecurityAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		var random [16]byte
		if _, err := rand.Read(random[:]); err != nil {
			c.AbortWithStatus(500)
			return
		}
		requestID := hex.EncodeToString(random[:])
		c.Header("X-Request-ID", requestID)
		c.Next()
		path := c.FullPath()
		if path == "" {
			path = "(unmatched)"
		}
		sensitive := strings.Contains(path, "/login") || strings.Contains(path, "/register") || strings.Contains(path, "/password/") || strings.Contains(path, "/logout")
		if !sensitive && c.Writer.Status() < 400 {
			return
		}
		event := struct {
			Event     string `json:"event"`
			RequestID string `json:"request_id"`
			Method    string `json:"method"`
			Route     string `json:"route"`
			Status    int    `json:"status"`
			IP        string `json:"ip"`
			Actor     string `json:"actor_id,omitempty"`
		}{Event: "security_request", RequestID: requestID, Method: c.Request.Method, Route: path, Status: c.Writer.Status(), IP: c.ClientIP()}
		if claims, ok := AuthClaims(c); ok {
			event.Actor = claims.UserID
		}
		encoded, err := json.Marshal(event)
		if err == nil {
			log.Print(string(encoded))
		}
	}
}
