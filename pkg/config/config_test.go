package config

import "testing"

func TestGet(t *testing.T) {
	t.Setenv("FOO", "bar")
	if got := Get("FOO", "fallback"); got != "bar" {
		t.Errorf("Get(FOO) = %q; want bar", got)
	}
	if got := Get("MISSING_KEY", "fallback"); got != "fallback" {
		t.Errorf("Get(MISSING_KEY) = %q; want fallback", got)
	}
}

func TestGetInt(t *testing.T) {
	t.Setenv("PORT", "8081")
	if got := GetInt("PORT", 3000); got != 8081 {
		t.Errorf("GetInt(PORT) = %d; want 8081", got)
	}
	if got := GetInt("MISSING_INT", 3000); got != 3000 {
		t.Errorf("GetInt(MISSING_INT) = %d; want 3000 (fallback)", got)
	}
	t.Setenv("BAD_INT", "not-a-number")
	if got := GetInt("BAD_INT", 3000); got != 3000 {
		t.Errorf("GetInt(BAD_INT) = %d; want 3000 (fallback on parse error)", got)
	}
}

func TestMustGetPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet(missing) did not panic")
		}
	}()
	MustGet("DEFINITELY_MISSING_KEY_12345")
}
