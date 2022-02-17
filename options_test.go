package kenall_test

import (
	"testing"

	"github.com/osamingo/go-kenall"
)

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	ret := kenall.WithHTTPClient(nil)
	if ret == nil {
		t.Error("a return value should not be nil")
	}
}

func TestWithEndpoint(t *testing.T) {
	t.Parallel()

	ret := kenall.WithEndpoint("")
	if ret == nil {
		t.Error("a return value should not be nil")
	}
}
