package analyzer

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewGRPCDetector(t *testing.T) {
	logger := logrus.New()
	detector := NewGRPCDetector(logger)

	if detector == nil {
		t.Fatal("NewGRPCDetector returned nil")
	}

	if detector.logger != logger {
		t.Error("Logger not set correctly")
	}

	if detector.dependencies == nil {
		t.Error("Dependencies slice not initialized")
	}
}

func TestExtractServiceFromClientName(t *testing.T) {
	logger := logrus.New()
	detector := NewGRPCDetector(logger)

	tests := []struct {
		name       string
		clientName string
		want       string
	}{
		{
			name:       "client with suffix",
			clientName: "paymentClient",
			want:       "payment",
		},
		{
			name:       "stub with suffix",
			clientName: "inventoryStub",
			want:       "inventory",
		},
		{
			name:       "new prefix",
			clientName: "newUserClient",
			want:       "user",
		},
		{
			name:       "New capitalized prefix",
			clientName: "NewOrderClient",
			want:       "order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractServiceFromClientName(tt.clientName)
			if result != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, result)
			}
		})
	}
}
