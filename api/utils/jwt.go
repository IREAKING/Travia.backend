package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type JwtClams struct {
	Id     pgtype.UUID `json:"id"`
	Email  string      `json:"email"`
	Vaitro string      `json:"vaitro"`
	jwt.RegisteredClaims
}
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
type s map[string]string

func GenerateToken(id pgtype.UUID, email, vaitro, secretkey string) (*TokenPair, error) {
	accessToken, err := generateAccessToken(id, email, vaitro, secretkey)
	if err != nil {
		fmt.Println(map[string]string{
			"message": "lỗi khi tạo accessToken",
			"error":   err.Error(),
		})
		return nil, err
	}
	refreshToken, err := generateRefreshToken(id, email, vaitro, secretkey)
	if err != nil {
		fmt.Println(map[string]string{
			"message": "Lỗi khi tạo refreshToken",
			"error":   err.Error(),
		})
		return nil, err
	}
	token := &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return token, nil
}

func ValidateToken(tokenStr, secretkey string) (*JwtClams, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JwtClams{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("phương thức ký không hợp lệ")
		}
		return []byte(secretkey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JwtClams); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("token không hợp lệ")
}

func generateAccessToken(id pgtype.UUID, email, vaitro, secretkey string) (string, error) {
	jwtclams := JwtClams{
		Id:     id,
		Email:  email,
		Vaitro: vaitro,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "localhost:3000/travia",
			Subject:   id.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // thời gian hết hạn, sau thời gian này jwt không còn giấ trị
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtclams)
	return accessToken.SignedString([]byte(secretkey))
}
func generateRefreshToken(id pgtype.UUID, email, vaitro, secretkey string) (string, error) {
	jwtclams := JwtClams{
		Id:     id,
		Email:  email,
		Vaitro: vaitro,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "localhost:3000/travia",
			Subject:   id.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtclams)
	return accessToken.SignedString([]byte(secretkey))
}
