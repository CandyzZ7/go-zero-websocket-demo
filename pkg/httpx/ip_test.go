package httpx

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "Valid X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
			},
			remoteAddr: "198.51.100.1:12345",
			expectedIP: "203.0.113.195",
		},
		{
			name: "Valid X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "198.51.100.10",
			},
			remoteAddr: "198.51.100.1:12345",
			expectedIP: "198.51.100.10",
		},
		{
			name:       "No headers, use RemoteAddr",
			headers:    nil,
			remoteAddr: "198.51.100.1:12345",
			expectedIP: "198.51.100.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			req.RemoteAddr = tt.remoteAddr

			ip := GetClientIP(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestClientPublicIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "Valid public IP in X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
			},
			remoteAddr: "192.168.0.1:12345",
			expectedIP: "203.0.113.195",
		},
		{
			name:       "No headers, RemoteAddr is local",
			headers:    nil,
			remoteAddr: "127.0.0.1:12345",
			expectedIP: "",
		},
		{
			name: "Valid public IP in X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "198.51.100.10",
			},
			remoteAddr: "192.168.0.1:12345",
			expectedIP: "198.51.100.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			req.RemoteAddr = tt.remoteAddr

			ip := GetClientPublicIP(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}
