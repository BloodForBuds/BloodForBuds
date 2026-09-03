package kms

import "testing"

func TestNewRequiresAddress(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected an error when the OpenBao address is empty")
	}
}
