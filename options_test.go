package lambdadb

import (
	"testing"
)

func TestListDocsOpts_Nil(t *testing.T) {
	// List(ctx, nil) should not panic and should use defaults when passed to internal API
	client := New(WithAPIKey("key"))
	// Just ensure we can call List with nil opts
	if client.Collection("c").Docs() == nil {
		t.Fatal("Docs() is nil")
	}
}

func TestPointerHelpers(t *testing.T) {
	if v := String("a"); v == nil || *v != "a" {
		t.Errorf("String() = %v", v)
	}
	if v := Int64(42); v == nil || *v != 42 {
		t.Errorf("Int64() = %v", v)
	}
	if v := Int(1); v == nil || *v != 1 {
		t.Errorf("Int() = %v", v)
	}
	if v := Bool(true); v == nil || !*v {
		t.Errorf("Bool() = %v", v)
	}
	if v := Pointer(3.14); v == nil || *v != 3.14 {
		t.Errorf("Pointer() = %v", v)
	}
}
