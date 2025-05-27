package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

// Compression middleware for response compression
func Compression() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the client accepts gzip encoding
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Check if the response should be compressed
		contentType := c.Writer.Header().Get("Content-Type")
		shouldCompress := strings.Contains(contentType, "text/") ||
			strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "application/xml") ||
			strings.Contains(contentType, "application/javascript")

		if !shouldCompress {
			c.Next()
			return
		}

		// Create a gzip writer
		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestCompression)
		if err != nil {
			c.Next()
			return
		}
		defer gz.Close()

		// Replace the writer with our gzip writer
		c.Writer = &gzipWriter{
			ResponseWriter: c.Writer,
			gzipWriter:     gz,
		}

		// Set the gzip header
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Call the next handler
		c.Next()
	}
}

// gzipWriter wraps a gin.ResponseWriter and writes to a gzip.Writer
type gzipWriter struct {
	gin.ResponseWriter
	gzipWriter *gzip.Writer
}

// Write writes the data to the gzip writer
func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.gzipWriter.Write(data)
}

// WriteString writes a string to the gzip writer
func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.gzipWriter.Write([]byte(s))
}

// DecompressRequest middleware to decompress request body if it's compressed
func DecompressRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the request body is compressed
		if c.Request.Header.Get("Content-Encoding") == "gzip" {
			// Read the body
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Next()
				return
			}
			defer c.Request.Body.Close()

			// Create a gzip reader
			reader, err := gzip.NewReader(bytes.NewReader(body))
			if err != nil {
				c.Next()
				return
			}
			defer reader.Close()

			// Read the decompressed body
			decompressedBody, err := io.ReadAll(reader)
			if err != nil {
				c.Next()
				return
			}

			// Replace the request body with the decompressed body
			c.Request.Body = io.NopCloser(bytes.NewReader(decompressedBody))
			c.Request.ContentLength = int64(len(decompressedBody))
			c.Request.Header.Del("Content-Encoding")
			c.Request.Header.Del("Content-Length")
		}

		c.Next()
	}
}
