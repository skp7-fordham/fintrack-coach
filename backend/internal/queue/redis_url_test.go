package queue

import (
	"testing"
)

func TestParseRedisURLPlain(t *testing.T) {
	opts, err := ParseRedisURL("redis://localhost:6379/0")
	if err != nil {
		t.Fatalf("ParseRedisURL: %v", err)
	}
	if opts.TLSConfig != nil {
		t.Fatal("redis:// must not enable TLS")
	}
	if opts.Addr != "localhost:6379" {
		t.Fatalf("addr = %q", opts.Addr)
	}
}

func TestParseRedisURLTLS(t *testing.T) {
	opts, err := ParseRedisURL("rediss://default:secret@notable-osprey-12345.upstash.io:6379")
	if err != nil {
		t.Fatalf("ParseRedisURL: %v", err)
	}
	if opts.TLSConfig == nil {
		t.Fatal("rediss:// must enable TLS")
	}
	if opts.Addr != "notable-osprey-12345.upstash.io:6379" {
		t.Fatalf("addr = %q", opts.Addr)
	}
	if opts.Password != "secret" {
		t.Fatal("password not parsed")
	}
}
