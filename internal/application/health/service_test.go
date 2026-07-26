package health

import "testing"

func TestServiceStatus(t *testing.T) {
	service := NewService()

	status := service.Status()

	if status.Status != "ok" {
		t.Fatalf("expected status ok, got %q", status.Status)
	}
	if status.Message != "Hello, World!!!" {
		t.Fatalf("expected hello message, got %q", status.Message)
	}
}
