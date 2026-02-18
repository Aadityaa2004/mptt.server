package implementation

import (
	"context"
	"database/sql"
	"time"

	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
)

type PostgresVerificationTokenRepository struct {
	db *sql.DB
}

func NewPostgresVerificationTokenRepository(db *sql.DB) *PostgresVerificationTokenRepository {
	return &PostgresVerificationTokenRepository{db: db}
}

func (r *PostgresVerificationTokenRepository) Create(ctx context.Context, token *auth_models.VerificationToken) error {
	query := `
		INSERT INTO verification_tokens (id, email, token_hash, purpose, metadata, expires_at, used_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	var metadata interface{}
	if len(token.Metadata) > 0 {
		metadata = token.Metadata
	}
	_, err := r.db.ExecContext(ctx, query, token.ID, token.Email, token.TokenHash, token.Purpose,
		metadata, token.ExpiresAt, token.UsedAt, token.CreatedAt)
	return err
}

func (r *PostgresVerificationTokenRepository) GetValidByEmailAndPurpose(ctx context.Context, email, purpose string) (*auth_models.VerificationToken, error) {
	query := `
		SELECT id, email, token_hash, purpose, metadata, expires_at, used_at, created_at
		FROM verification_tokens
		WHERE email = $1 AND purpose = $2 AND used_at IS NULL AND expires_at > $3
		ORDER BY created_at DESC
		LIMIT 1
	`
	var token auth_models.VerificationToken
	var usedAt sql.NullTime
	var metadata []byte
	err := r.db.QueryRowContext(ctx, query, email, purpose, time.Now()).Scan(
		&token.ID, &token.Email, &token.TokenHash, &token.Purpose,
		&metadata, &token.ExpiresAt, &usedAt, &token.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	token.Metadata = metadata
	if usedAt.Valid {
		token.UsedAt = &usedAt.Time
	}
	return &token, nil
}

func (r *PostgresVerificationTokenRepository) GetValidByTokenHash(ctx context.Context, tokenHash, purpose string) (*auth_models.VerificationToken, error) {
	query := `
		SELECT id, email, token_hash, purpose, metadata, expires_at, used_at, created_at
		FROM verification_tokens
		WHERE token_hash = $1 AND purpose = $2 AND used_at IS NULL AND expires_at > $3
		ORDER BY created_at DESC
		LIMIT 1
	`
	var token auth_models.VerificationToken
	var usedAt sql.NullTime
	var metadata []byte
	err := r.db.QueryRowContext(ctx, query, tokenHash, purpose, time.Now()).Scan(
		&token.ID, &token.Email, &token.TokenHash, &token.Purpose,
		&metadata, &token.ExpiresAt, &usedAt, &token.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	token.Metadata = metadata
	if usedAt.Valid {
		token.UsedAt = &usedAt.Time
	}
	return &token, nil
}

func (r *PostgresVerificationTokenRepository) GetLatestSignupTokenByEmail(ctx context.Context, email string) (*auth_models.VerificationToken, error) {
	query := `
		SELECT id, email, token_hash, purpose, metadata, expires_at, used_at, created_at
		FROM verification_tokens
		WHERE email = $1 AND purpose = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	var token auth_models.VerificationToken
	var usedAt sql.NullTime
	var metadata []byte
	err := r.db.QueryRowContext(ctx, query, email, auth_models.PurposeSignupOTP).Scan(
		&token.ID, &token.Email, &token.TokenHash, &token.Purpose,
		&metadata, &token.ExpiresAt, &usedAt, &token.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	token.Metadata = metadata
	if usedAt.Valid {
		token.UsedAt = &usedAt.Time
	}
	return &token, nil
}

func (r *PostgresVerificationTokenRepository) MarkUsed(ctx context.Context, id string) error {
	query := `UPDATE verification_tokens SET used_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *PostgresVerificationTokenRepository) InvalidateByEmailAndPurpose(ctx context.Context, email, purpose string) error {
	query := `UPDATE verification_tokens SET used_at = $1 WHERE email = $2 AND purpose = $3 AND used_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, time.Now(), email, purpose)
	return err
}

func (r *PostgresVerificationTokenRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM verification_tokens WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresVerificationTokenRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	query := `DELETE FROM verification_tokens WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, before)
	return err
}
