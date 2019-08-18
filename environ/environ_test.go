package environ_test

import (
	"os"
	"testing"

	"github.com/mimir-news/mimir-go/environ"
)

func TestGet(t *testing.T) {
	name := "ENV_TEST_GET_KEY"
	expectedValue := "some-value-for-get"

	err := os.Setenv(name, expectedValue)
	if err != nil {
		t.Error("os.Setenv returned unexpected error:", err)
	}

	value := environ.Get(name, "default-value")
	if value != expectedValue {
		t.Errorf("env.Get test failed. Expected: [%s] Got: [%s]", expectedValue, value)
	}

	value = environ.Get("ENV_TEST_GET_SOME_OTHER_VALUE", "default-value")
	if value != "default-value" {
		t.Errorf("env.Get test failed. Expected: [default-value] Got: [%s]", value)
	}
}

func TestMustGet(t *testing.T) {
	name := "ENV_TEST_GET_KEY"
	expectedValue := "some-value-for-get"

	err := os.Setenv(name, expectedValue)
	if err != nil {
		t.Error("os.Setenv returned unexpected error:", err)
	}

	value := environ.MustGet(name)
	if value != expectedValue {
		t.Errorf("env.MustGet test failed. Expected: [%s] Got: [%s]", expectedValue, value)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("env.MustGet test failed. Did not panic for missing value")
		}
	}()
	// Should throw an error.
	environ.MustGet("ENV_TEST_GET_SOME_OTHER_VALUE")
}

