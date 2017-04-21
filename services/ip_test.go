package services

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestGetIP(t *testing.T) {
	ip := "123.14.3.45"
	r := NewRequest()
	r.RemoteAddr = ip + ":80"
	assert.Equal(t, ip, getIp(r))
}

func TestUndefinedIP(t *testing.T) {
	r := NewRequest()
	assert.Equal(t, "undefined", getIp(r))
}

func TestGetIPViaNginxProxy(t *testing.T) {
	ip := "123.14.3.45"
	r := NewRequest()
	r.Header.Set("X-Real-IP", ip)
	r.RemoteAddr = "127.0.0.1:80"
	assert.Equal(t, ip, getIp(r))
}

func TestGetIPViaProxy(t *testing.T) {
	ip := "123.14.3.45"
	r := NewRequest()
	r.Header.Set("X-Forwarded-For", ip)
	r.RemoteAddr = "120.34.3.1:80"
	assert.Equal(t, ip, getIp(r))
}

func TestBadIP(t *testing.T) {
	r := NewRequest()
	r.RemoteAddr = "14.3.45:80"
	assert.Equal(t, "undefined", getIp(r))
}
