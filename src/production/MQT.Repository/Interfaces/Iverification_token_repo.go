package interfaces

import (
	"context"
	"time"

	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
)

type VerificationTokenRepository interface {
	Create(ctx context.Context, token *auth_models.VerificationToken) error
	GetValidByEmailAndPurpose(ctx context.Context, email, purpose string) (*auth_models.VerificationToken, error)
	GetValidByTokenHash(ctx context.Context, tokenHash, purpose string) (*auth_models.VerificationToken, error)
	GetLatestSignupTokenByEmail(ctx context.Context, email string) (*auth_models.VerificationToken, error)
	MarkUsed(ctx context.Context, id string) error
	InvalidateByEmailAndPurpose(ctx context.Context, email, purpose string) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context, before time.Time) error
}
