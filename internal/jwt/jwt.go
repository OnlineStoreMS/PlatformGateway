package jwt

import (
	"errors"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var ErrInvalid = errors.New("invalid token")

type Claims struct {
	UserID      uint64   `json:"uid"`
	CompanyID   uint64   `json:"cid"`
	TenantID    uint64   `json:"tid"`
	Email       string   `json:"email"`
	DisplayName string   `json:"name"`
	Permissions []string `json:"perms"`
	IsPlatform  bool     `json:"platform"`
	jwtlib.RegisteredClaims
}

type Validator struct {
	secret []byte
}

func NewValidator(secret string) *Validator {
	return &Validator{secret: []byte(secret)}
}

func (v *Validator) Parse(tokenStr string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtlib.Token) (interface{}, error) {
		if t.Method != jwtlib.SigningMethodHS256 {
			return nil, ErrInvalid
		}
		return v.secret, nil
	})
	if err != nil {
		return nil, ErrInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.TenantID == 0 {
		return nil, ErrInvalid
	}
	return claims, nil
}
