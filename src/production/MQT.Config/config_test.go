package config

import (
	"os"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{"valid", &Config{
			Database: DatabaseConfig{User: "u", Password: "p"},
			Auth:     AuthConfig{PasswordMinLength: 8},
		}, false},
		{"missing user", &Config{
			Database: DatabaseConfig{User: "", Password: "p"},
			Auth:     AuthConfig{PasswordMinLength: 8},
		}, true},
		{"missing password", &Config{
			Database: DatabaseConfig{User: "u", Password: ""},
			Auth:     AuthConfig{PasswordMinLength: 8},
		}, true},
		{"password too short", &Config{
			Database: DatabaseConfig{User: "u", Password: "p"},
			Auth:     AuthConfig{PasswordMinLength: 4},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GetDatabaseDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "db",
			Port:     5433,
			User:     "user",
			Password: "pass",
			DBName:   "mydb",
			SSLMode:  "require",
		},
	}
	dsn := cfg.GetDatabaseDSN()
	if dsn != "host=db port=5433 user=user password=pass dbname=mydb sslmode=require" {
		t.Errorf("unexpected DSN: %s", dsn)
	}
}

func TestConfig_GetMQTTBrokerURL(t *testing.T) {
	t.Run("tcp", func(t *testing.T) {
		cfg := &Config{MQTT: MQTTConfig{BrokerHost: "broker", BrokerPort: 1883, UseTLS: false}}
		url := cfg.GetMQTTBrokerURL()
		if url != "tcp://broker:1883" {
			t.Errorf("got %s", url)
		}
	})
	t.Run("tcps", func(t *testing.T) {
		cfg := &Config{MQTT: MQTTConfig{BrokerHost: "broker", BrokerPort: 8883, UseTLS: true}}
		url := cfg.GetMQTTBrokerURL()
		if url != "tcps://broker:8883" {
			t.Errorf("got %s", url)
		}
	})
}

func TestLoadIngestorConfig(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-secret")
	os.Setenv("API_SERVICE_URL", "http://api:9002")
	defer os.Unsetenv("INTERNAL_API_SECRET")
	defer os.Unsetenv("API_SERVICE_URL")

	cfg, err := LoadIngestorConfig()
	if err != nil {
		t.Fatalf("LoadIngestorConfig: %v", err)
	}
	if cfg.InternalAPISecret != "test-secret" {
		t.Errorf("unexpected secret")
	}
}

func TestLoadEmailConfig(t *testing.T) {
	cfg, err := LoadEmailConfig()
	if err != nil {
		t.Fatalf("LoadEmailConfig: %v", err)
	}
	if cfg.Server.Port == "" {
		t.Error("expected port")
	}
}

func TestLoadApiConfig(t *testing.T) {
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	defer os.Unsetenv("POSTGRES_USER")
	defer os.Unsetenv("POSTGRES_PASSWORD")

	cfg, err := LoadApiConfig()
	if err != nil {
		t.Fatalf("LoadApiConfig: %v", err)
	}
	if cfg.Database.User != "testuser" {
		t.Errorf("unexpected user: %s", cfg.Database.User)
	}
}

func TestLoad(t *testing.T) {
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	defer os.Unsetenv("POSTGRES_USER")
	defer os.Unsetenv("POSTGRES_PASSWORD")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Database.User != "testuser" {
		t.Errorf("unexpected user: %s", cfg.Database.User)
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "val")
	defer os.Unsetenv("TEST_KEY")
	if getEnv("TEST_KEY", "default") != "val" {
		t.Error("getEnv with set key")
	}
	if getEnv("MISSING_KEY", "default") != "default" {
		t.Error("getEnv with missing key")
	}
}

func TestGetInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")
	if getInt("TEST_INT", 0) != 42 {
		t.Error("getInt")
	}
	if getInt("MISSING_INT", 99) != 99 {
		t.Error("getInt default")
	}
}

func TestGetBool(t *testing.T) {
	for _, s := range []string{"1", "true", "TRUE"} {
		os.Setenv("TEST_BOOL", s)
		if getBool("TEST_BOOL", false) != true {
			t.Errorf("getBool %s", s)
		}
	}
	for _, s := range []string{"0", "false", "FALSE"} {
		os.Setenv("TEST_BOOL", s)
		if getBool("TEST_BOOL", true) != false {
			t.Errorf("getBool %s", s)
		}
	}
	os.Unsetenv("TEST_BOOL")
	if getBool("TEST_BOOL", true) != true {
		t.Error("getBool default")
	}
}

func TestGetDuration(t *testing.T) {
	os.Setenv("TEST_DUR", "5s")
	defer os.Unsetenv("TEST_DUR")
	if getDuration("TEST_DUR", time.Second) != 5*time.Second {
		t.Error("getDuration")
	}
}

func TestGetStringSlice(t *testing.T) {
	os.Setenv("TEST_SLICE", "a, b , c")
	defer os.Unsetenv("TEST_SLICE")
	got := getStringSlice("TEST_SLICE", nil)
	if len(got) != 3 {
		t.Errorf("getStringSlice: %v", got)
	}
}

func TestSplitString(t *testing.T) {
	got := splitString("a,b,c", ",")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("splitString: %v", got)
	}
	if len(splitString("", ",")) != 0 {
		t.Error("splitString empty")
	}
}

func TestTrimString(t *testing.T) {
	if trimString("  x  ") != "x" {
		t.Error("trimString")
	}
}
