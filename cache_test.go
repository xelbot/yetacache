package yetacache

import (
	"testing"
	"time"
)

func TestHasMethod(t *testing.T) {
	c := New[int, int](time.Millisecond, time.Second)
	c.Set(1, 1, DefaultTTL)

	if c.Has(2) {
		t.Error("Found value that shouldn't exist")
	}

	if !c.Has(1) {
		t.Error("Not found value that should exist")
	}

	time.Sleep(1010 * time.Microsecond)

	if c.Has(1) {
		t.Error("Found value that should have expired")
	}

	c.StopCleanup()
}

func TestGetMethod(t *testing.T) {
	c := New[string, string](time.Millisecond, time.Second)
	c.Set("123", "abc", DefaultTTL)

	if _, found := c.Get("321"); found {
		t.Error("Found value that shouldn't exist")
	}

	if val, found := c.Get("123"); !found {
		t.Error("Not found value that should exist")
	} else if val != "abc" {
		t.Error("Found incorrect value: ", val)
	}

	time.Sleep(1010 * time.Microsecond)

	if _, found := c.Get("123"); found {
		t.Error("Found value that should have expired")
	}

	c.StopCleanup()
}

func TestDeleteMethod(t *testing.T) {
	c := New[string, int](time.Millisecond, time.Second)
	c.Set("abc", 36, DefaultTTL)
	c.Set("def", 72, DefaultTTL)

	c.Delete("abc")

	if c.Has("abc") {
		t.Error("Found value that shouldn't exist")
	}
	if !c.Has("def") {
		t.Error("Not found value that should exist")
	}

	c.StopCleanup()
}

func TestClearMethod(t *testing.T) {
	c := New[string, int](time.Millisecond, time.Second)
	c.Set("abc", 36, DefaultTTL)
	c.Set("def", 72, DefaultTTL)

	c.Clear()

	if c.Has("abc") || c.Has("def") {
		t.Error("Found value that shouldn't exist")
	}

	c.StopCleanup()
}
