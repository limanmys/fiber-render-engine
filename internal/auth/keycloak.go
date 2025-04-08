package auth

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/helpers"

	"github.com/Nerzal/gocloak/v13"
)

type Keycloak struct {
	Client *gocloak.GoCloak
	Ctx    context.Context
}

var (
	keycloak Keycloak
)

func GetOauth2Token(user_id string) (*models.Oauth2Token, error) {
	token := &models.Oauth2Token{}
	err := database.Connection().First(&token, "user_id = ?", user_id).Error
	if err != nil {
		return nil, err
	}
	return token, nil
}

func RefreshTokenIfNecessary(user_id string) error {
	token, err := GetOauth2Token(user_id)
	if err != nil {
		return errors.New("an error occured while getting keycloak token")
	}

	if keycloak.Client == nil {
		keycloak.Client = gocloak.NewClient(helpers.Env("KEYCLOAK_BASE_URL", ""))

		client := resty.New()
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		keycloak.Client.SetRestyClient(client)

		keycloak.Ctx = context.Background()
	}

	result, _, err := keycloak.Client.DecodeAccessToken(
		keycloak.Ctx,
		token.AccessToken,
		helpers.Env("KEYCLOAK_REALM", ""),
	)

	if err != nil && !strings.Contains(err.Error(), "expired") {
		return errors.New("an error occured while validating token")
	}

	if err != nil && strings.Contains(err.Error(), "expired") {
		err := RefreshToken(token)
		if err != nil {
			return err
		}

		return nil
	}

	if !result.Valid {
		err := RefreshToken(token)
		if err != nil {
			return err
		}
	}

	return nil
}

func RefreshToken(token *models.Oauth2Token) error {
	jwt, err := keycloak.Client.RefreshToken(
		keycloak.Ctx,
		token.RefreshToken,
		helpers.Env("KEYCLOAK_CLIENT_ID", ""),
		helpers.Env("KEYCLOAK_CLIENT_SECRET", ""),
		helpers.Env("KEYCLOAK_REALM", ""),
	)

	if err != nil {
		return errors.New("an error occured while refreshing token")
	}

	err = database.Connection().
		Model(&token).
		Where("user_id = ?", token.UserID).
		Updates(&models.Oauth2Token{
			AccessToken:      jwt.AccessToken,
			RefreshToken:     jwt.RefreshToken,
			ExpiresIn:        jwt.ExpiresIn,
			RefreshExpiresIn: jwt.RefreshExpiresIn,
		}).Error

	if err != nil {
		return errors.New("an error occured while updating token")
	}

	return nil
}
