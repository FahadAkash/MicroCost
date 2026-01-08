package analyzer

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewHTTPDetector(t *testing.T) {
	logger := logrus.New()
	detector := NewHTTPDetector(logger)

	if detector == nil {
		t.Fatal("NewHTTPDetector returned nil")
	}

	if detector.logger != logger {
		t.Error("Logger not set correctly")
	}

	if len(detector.urlPatterns) == 0 {
		t.Error("URL patterns not initialized")
	}
}

func TestExtractServiceFromURL(t *testing.T) {
	logger := logrus.New()
	detector := NewHTTPDetector(logger)

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "service with domain",
			url:  "http://payment-service.example.com/api/pay",
			want: "payment-service",
		},
		{
			name: "service with port",
			url:  "http://inventory:8080/api/check",
			want: "inventory",
		},
		{
			name: "simple service",
			url:  "http://notification/send",
			want: "notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractServiceFromURL(tt.url)
			if result != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, result)
			}
		})
	}
}

func TestExtractEndpointFromURL(t *testing.T) {
	logger := logrus.New()
	detector := NewHTTPDetector(logger)

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "API endpoint",
			url:  "http://payment.com/api/v1/pay",
			want: "/api/v1/pay",
		},
		{
			name: "root endpoint",
			url:  "http://service.com",
			want: "/",
		},
		{
			name: "simple path",
			url:  "http://service.com/users",
			want: "/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractEndpointFromURL(tt.url)
			if result != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, result)
			}
		})
	}
}

func TestGenerateDependencyID(t *testing.T) {
	tests := []struct {
		from     string
		to       string
		endpoint string
		want     string
	}{
		{
			from:     "service-a",
			to:       "service-b",
			endpoint: "/api/test",
			want:     "service-a->service-b/api/test",
		},
	}

	for _, tt := range tests {
		result := generateDependencyID(tt.from, tt.to, tt.endpoint)
		if result != tt.want {
			t.Errorf("Expected %s, got %s", tt.want, result)
		}
	}
}
