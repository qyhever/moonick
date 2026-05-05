package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInitMergesAppEnvAndLocalConfig(t *testing.T) {
	t.Setenv("MOONICK_ENV", "dev")

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "internal", "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	writeTestFile(t, filepath.Join(configDir, "app.yml"), `
mode: "app"
server:
  port: 6301
logger:
  level: "info"
  filename: "./log/app.log"
database:
  mysql:
    addr: "app-host:3306"
    user: "app-user"
    db_name: "app-db"
auth:
  whitelist:
    - "GET:/health"
`)
	writeTestFile(t, filepath.Join(configDir, "dev.yml"), `
mode: "dev"
server:
  port: 6303
logger:
  level: "debug"
database:
  mysql:
    addr: "dev-host:3306"
`)
	writeTestFile(t, filepath.Join(configDir, "dev.local.yml"), `
database:
  mysql:
    user: "local-user"
auth:
  whitelist:
    - "GET:/local-health"
`)

	prevWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWd)
		viper.Reset()
		GlobalConfig = nil
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	if err := Init(); err != nil {
		t.Fatalf("init config: %v", err)
	}

	if GlobalConfig.Mode != "dev" {
		t.Fatalf("mode = %q, want dev", GlobalConfig.Mode)
	}
	if GlobalConfig.Server.Port != 6303 {
		t.Fatalf("server.port = %d, want 6303", GlobalConfig.Server.Port)
	}
	if GlobalConfig.Logger.Level != "debug" {
		t.Fatalf("logger.level = %q, want debug", GlobalConfig.Logger.Level)
	}
	if GlobalConfig.Logger.Filename != "./log/app.log" {
		t.Fatalf("logger.filename = %q, want ./log/app.log", GlobalConfig.Logger.Filename)
	}
	if GlobalConfig.Database.MySQL.Addr != "dev-host:3306" {
		t.Fatalf("database.mysql.addr = %q, want dev-host:3306", GlobalConfig.Database.MySQL.Addr)
	}
	if GlobalConfig.Database.MySQL.User != "local-user" {
		t.Fatalf("database.mysql.user = %q, want local-user", GlobalConfig.Database.MySQL.User)
	}
	if GlobalConfig.Database.MySQL.DBName != "app-db" {
		t.Fatalf("database.mysql.db_name = %q, want app-db", GlobalConfig.Database.MySQL.DBName)
	}
	if len(GlobalConfig.Auth.Whitelist) != 1 || GlobalConfig.Auth.Whitelist[0] != "GET:/local-health" {
		t.Fatalf("auth.whitelist = %#v, want local override", GlobalConfig.Auth.Whitelist)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
