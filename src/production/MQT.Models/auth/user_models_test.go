package auth_models

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	lat := 43.65
	lon := -79.38
	u := NewUser("alice", "alice@test.com", "hashed", "user", &lat, &lon)
	if u == nil {
		t.Fatal("NewUser returned nil")
	}
	if u.Username != "alice" || u.Email != "alice@test.com" || u.Password != "hashed" || u.Role != "user" {
		t.Errorf("unexpected user fields: %+v", u)
	}
	if !u.Active {
		t.Error("expected Active=true")
	}
	if u.Latitude == nil || *u.Latitude != lat {
		t.Error("unexpected Latitude")
	}
	if u.Longitude == nil || *u.Longitude != lon {
		t.Error("unexpected Longitude")
	}
	if len(u.Locations) != 0 {
		t.Errorf("expected empty Locations, got %v", u.Locations)
	}
}

func TestNewUser_NilCoords(t *testing.T) {
	u := NewUser("bob", "bob@test.com", "pwd", "admin", nil, nil)
	if u.Latitude != nil || u.Longitude != nil {
		t.Error("expected nil coords")
	}
}

func TestDeviceLocationArray_Value_Empty(t *testing.T) {
	var a DeviceLocationArray
	v, err := a.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if v != "[]" {
		t.Errorf("expected [], got %v", v)
	}
}

func TestDeviceLocationArray_Value_NonEmpty(t *testing.T) {
	a := DeviceLocationArray{
		{DeviceID: "d1", PiID: "p1", Latitude: 1, Longitude: 2},
	}
	v, err := a.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	s, ok := v.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", v)
	}
	if len(s) == 0 || string(s) == "[]" {
		t.Error("expected non-empty JSON")
	}
}

func TestDeviceLocationArray_Scan_Nil(t *testing.T) {
	var a DeviceLocationArray
	err := a.Scan(nil)
	if err != nil {
		t.Fatalf("Scan nil: %v", err)
	}
	if len(a) != 0 {
		t.Errorf("expected empty, got %d", len(a))
	}
}

func TestDeviceLocationArray_Scan_EmptySlice(t *testing.T) {
	var a DeviceLocationArray
	err := a.Scan([]byte("[]"))
	if err != nil {
		t.Fatalf("Scan []: %v", err)
	}
	if len(a) != 0 {
		t.Errorf("expected empty, got %d", len(a))
	}
}

func TestDeviceLocationArray_Scan_String(t *testing.T) {
	var a DeviceLocationArray
	err := a.Scan("[]")
	if err != nil {
		t.Fatalf("Scan string: %v", err)
	}
	if len(a) != 0 {
		t.Errorf("expected empty, got %d", len(a))
	}
}

func TestDeviceLocationArray_Scan_ValidJSON(t *testing.T) {
	json := `[{"device_id":"d1","pi_id":"p1","latitude":1.0,"longitude":2.0}]`
	var a DeviceLocationArray
	err := a.Scan([]byte(json))
	if err != nil {
		t.Fatalf("Scan valid: %v", err)
	}
	if len(a) != 1 {
		t.Fatalf("expected 1 location, got %d", len(a))
	}
	if a[0].DeviceID != "d1" || a[0].PiID != "p1" {
		t.Errorf("unexpected location: %+v", a[0])
	}
}

func TestDeviceLocationArray_Scan_DefaultType(t *testing.T) {
	var a DeviceLocationArray
	err := a.Scan(123)
	if err != nil {
		t.Fatalf("Scan wrong type: %v", err)
	}
	if len(a) != 0 {
		t.Errorf("expected empty for wrong type, got %d", len(a))
	}
}
