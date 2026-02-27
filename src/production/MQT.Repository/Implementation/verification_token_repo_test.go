package implementation

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
)

func TestPostgresVerificationTokenRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	token := auth_models.NewVerificationToken("a@b.com", "hash", auth_models.PurposeSignupOTP, time.Now().Add(time.Hour), []byte(`{}`))
	mock.ExpectExec("INSERT INTO verification_tokens").
		WithArgs(token.ID, "a@b.com", "hash", auth_models.PurposeSignupOTP, sqlmock.AnyArg(), sqlmock.AnyArg(), nil, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewPostgresVerificationTokenRepository(db)
	err = repo.Create(context.Background(), token)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresVerificationTokenRepository_GetValidByEmailAndPurpose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	exp := time.Now().Add(time.Hour)
	rows := sqlmock.NewRows([]string{"id", "email", "token_hash", "purpose", "metadata", "expires_at", "used_at", "created_at"}).
		AddRow("t1", "a@b.com", "hash", auth_models.PurposeSignupOTP, []byte("{}"), exp, nil, time.Now())
	mock.ExpectQuery("SELECT .+ FROM verification_tokens").
		WithArgs("a@b.com", auth_models.PurposeSignupOTP, sqlmock.AnyArg()).
		WillReturnRows(rows)

	repo := NewPostgresVerificationTokenRepository(db)
	token, err := repo.GetValidByEmailAndPurpose(context.Background(), "a@b.com", auth_models.PurposeSignupOTP)
	if err != nil {
		t.Fatalf("GetValidByEmailAndPurpose: %v", err)
	}
	if token == nil || token.Email != "a@b.com" {
		t.Errorf("unexpected token: %+v", token)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresVerificationTokenRepository_GetValidByTokenHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	exp := time.Now().Add(time.Hour)
	rows := sqlmock.NewRows([]string{"id", "email", "token_hash", "purpose", "metadata", "expires_at", "used_at", "created_at"}).
		AddRow("t1", "a@b.com", "h123", auth_models.PurposePasswordReset, nil, exp, nil, time.Now())
	mock.ExpectQuery("SELECT .+ FROM verification_tokens").
		WithArgs("h123", auth_models.PurposePasswordReset, sqlmock.AnyArg()).
		WillReturnRows(rows)

	repo := NewPostgresVerificationTokenRepository(db)
	token, err := repo.GetValidByTokenHash(context.Background(), "h123", auth_models.PurposePasswordReset)
	if err != nil {
		t.Fatalf("GetValidByTokenHash: %v", err)
	}
	if token == nil || token.TokenHash != "h123" {
		t.Errorf("unexpected token: %+v", token)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresVerificationTokenRepository_GetLatestSignupTokenByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	exp := time.Now().Add(time.Hour)
	rows := sqlmock.NewRows([]string{"id", "email", "token_hash", "purpose", "metadata", "expires_at", "used_at", "created_at"}).
		AddRow("t1", "a@b.com", "hash", auth_models.PurposeSignupOTP, []byte("{}"), exp, nil, time.Now())
	mock.ExpectQuery("SELECT .+ FROM verification_tokens").
		WithArgs("a@b.com", auth_models.PurposeSignupOTP).
		WillReturnRows(rows)

	repo := NewPostgresVerificationTokenRepository(db)
	token, err := repo.GetLatestSignupTokenByEmail(context.Background(), "a@b.com")
	if err != nil {
		t.Fatalf("GetLatestSignupTokenByEmail: %v", err)
	}
	if token == nil || token.Email != "a@b.com" {
		t.Errorf("unexpected token: %+v", token)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresVerificationTokenRepository_MarkUsed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE verification_tokens SET used_at").
		WithArgs(sqlmock.AnyArg(), "t1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresVerificationTokenRepository(db)
	err = repo.MarkUsed(context.Background(), "t1")
	if err != nil {
		t.Fatalf("MarkUsed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresVerificationTokenRepository_InvalidateByEmailAndPurpose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE verification_tokens SET used_at").
		WithArgs(sqlmock.AnyArg(), "a@b.com", auth_models.PurposeSignupOTP).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresVerificationTokenRepository(db)
	err = repo.InvalidateByEmailAndPurpose(context.Background(), "a@b.com", auth_models.PurposeSignupOTP)
	if err != nil {
		t.Fatalf("InvalidateByEmailAndPurpose: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresVerificationTokenRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM verification_tokens").
		WithArgs("t1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresVerificationTokenRepository(db)
	err = repo.Delete(context.Background(), "t1")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresVerificationTokenRepository_DeleteExpired(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	before := time.Now()
	mock.ExpectExec("DELETE FROM verification_tokens").
		WithArgs(before).
		WillReturnResult(sqlmock.NewResult(0, 5))

	repo := NewPostgresVerificationTokenRepository(db)
	err = repo.DeleteExpired(context.Background(), before)
	if err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
