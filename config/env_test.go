package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	if err := LoadConfig(); err != nil {
		t.Fatal(err)
	}
	t.Logf("config: %+v", GetConfig())
}

func TestMain(m *testing.M) {
	if err := LoadConfig(); err != nil {
		return
	}
	os.Exit(m.Run())
}

func TestInitializeOSS(t *testing.T) {
	if err := InitializeOSS(); err != nil {
		t.Fatal(err)
	}
	t.Logf("oss: %+v", GetOSS())
}

func TestInitializeDB(t *testing.T) {
	if err := InitializeDB(); err != nil {
		t.Fatal(err)
	}
	t.Logf("db: %+v", GetDB())
}

func TestInitializeCache(t *testing.T) {
	if err := InitializeCache(); err != nil {
		t.Fatal(err)
	}
	t.Logf("cache: %+v", GetCache())

}
