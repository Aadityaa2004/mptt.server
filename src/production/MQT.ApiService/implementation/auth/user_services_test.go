package auth

import (
	"context"
	"testing"

	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
)

type mockPiRepo struct {
	unassignCalled bool
}

func (m *mockPiRepo) CreateOrUpdatePi(ctx context.Context, pi hardware_models.Pi) error { return nil }
func (m *mockPiRepo) GetPi(ctx context.Context, piID string) (*hardware_models.Pi, error) { return nil, nil }
func (m *mockPiRepo) ListPis(ctx context.Context, userID string, page, pageSize int) (*interfaces.PaginationResult, error) {
	return nil, nil
}
func (m *mockPiRepo) UpdatePi(ctx context.Context, pi hardware_models.Pi) error { return nil }
func (m *mockPiRepo) DeletePi(ctx context.Context, piID string, cascade bool) error { return nil }
func (m *mockPiRepo) UnassignPisByUserID(ctx context.Context, userID string) (int64, error) {
	m.unassignCalled = true
	return 1, nil
}

func TestUserService_GetUserByID(t *testing.T) {
	userRepo := newMockUserRepo()
	user := &auth_models.User{UserID: "u1", Username: "u1", Email: "u1@x.com", Role: "user"}
	userRepo.add(user)
	piRepo := &mockPiRepo{}

	svc := NewUserService(userRepo, piRepo)
	ctx := context.Background()

	got, err := svc.GetUserByID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.UserID != "u1" {
		t.Errorf("UserID = %q", got.UserID)
	}
}

func TestUserService_GetAllUsers(t *testing.T) {
	userRepo := newMockUserRepo()
	piRepo := &mockPiRepo{}
	svc := NewUserService(userRepo, piRepo)
	ctx := context.Background()

	users, err := svc.GetAllUsers(ctx)
	if err != nil {
		t.Fatalf("GetAllUsers: %v", err)
	}
	if users != nil && len(users) != 0 {
		t.Errorf("expected empty or nil, got %d users", len(users))
	}
}

func TestUserService_UpdateUserRole(t *testing.T) {
	userRepo := newMockUserRepo()
	user := &auth_models.User{UserID: "u1", Username: "u1", Role: "user"}
	userRepo.add(user)
	piRepo := &mockPiRepo{}

	svc := NewUserService(userRepo, piRepo)
	ctx := context.Background()

	got, err := svc.UpdateUserRole(ctx, "u1", "admin")
	if err != nil {
		t.Fatalf("UpdateUserRole: %v", err)
	}
	if got.Role != "admin" {
		t.Errorf("Role = %q", got.Role)
	}
	if !piRepo.unassignCalled {
		t.Error("UnassignPisByUserID should be called when promoting to admin")
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	userRepo := newMockUserRepo()
	user := &auth_models.User{UserID: "u1", Username: "u1", Role: "user"}
	userRepo.add(user)
	piRepo := &mockPiRepo{}

	svc := NewUserService(userRepo, piRepo)
	ctx := context.Background()
	user.Username = "updated"
	got, err := svc.UpdateUser(ctx, user)
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if got.Username != "updated" {
		t.Errorf("Username = %q", got.Username)
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	userRepo := newMockUserRepo()
	user := &auth_models.User{UserID: "u1"}
	userRepo.add(user)
	piRepo := &mockPiRepo{}

	svc := NewUserService(userRepo, piRepo)
	ctx := context.Background()
	err := svc.DeleteUser(ctx, "u1")
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
}

func TestUserService_GetUserByDeviceID(t *testing.T) {
	userRepo := newMockUserRepo()
	piRepo := &mockPiRepo{}
	svc := NewUserService(userRepo, piRepo)
	ctx := context.Background()

	_, err := svc.GetUserByDeviceID(ctx, "dev1")
	if err != nil {
		t.Fatalf("GetUserByDeviceID: %v", err)
	}
}

func TestUserService_HashPassword(t *testing.T) {
	svc := NewUserService(nil, nil)
	hash, err := svc.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" || hash == "secret123" {
		t.Error("expected hashed password")
	}
}
