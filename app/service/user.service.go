package service

import (
	"context"
	"oauth-server/app/entity"
	"oauth-server/app/helper"
	"oauth-server/app/model"
	"oauth-server/app/repository"
	postgres_repository "oauth-server/app/repository/postgres"
	"oauth-server/package/database"
	"oauth-server/package/errors"
	"oauth-server/utils"
	"time"

	"gorm.io/gorm"
)

type userService struct {
	helpers      helper.HelperCollections
	postgresRepo postgres_repository.PostgresRepositoryCollections
}

func NewUserService(
	helpers helper.HelperCollections,
	postgresRepo postgres_repository.PostgresRepositoryCollections,
) UserService {

	return &userService{
		helpers:      helpers,
		postgresRepo: postgresRepo,
	}
}

func (s *userService) Login(ctx context.Context, data *model.LoginRequest) (*model.LoginResponse, error) {
	var userOAuth *entity.Oauth

	// Check user exit
	user, err := s.postgresRepo.PostgresUserRepo.FindUserByFilter(ctx, nil, &repository.FindUserByFilter{
		PhoneNumber: &data.PhoneNumber,
		Email:       &data.Email,
	})
	if err != nil {
		return nil, errors.New(errors.ErrCodeUserNotFound)
	}
	if err := utils.CheckPasswordHash(data.Password, user.Password); err != nil {
		return nil, errors.New(errors.ErrCodeIncorrectPassword)
	}

	// Generate token
	accessToken, err := s.helpers.OauthHelper.GenerateAccessToken(*user)
	if err != nil || accessToken == "" {
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}
	refreshToken, err := s.helpers.OauthHelper.GenerateRefreshToken(*user)
	if err != nil || refreshToken == "" {
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}

	// Create User OAuth
	tx := database.BeginPostgresTransaction()
	userOAuth, err = s.postgresRepo.PostgresOAuthRepo.FindOAuthByFilter(ctx, tx, &repository.FindOAuthByFilter{
		UserID: &user.ID,
	})
	if err == gorm.ErrRecordNotFound {
		userOAuth = entity.NewOAuth()
		userOAuth.UserID = user.ID
		userOAuth.Status = entity.OAuthStatusActive
	}

	userOAuth.Token = refreshToken
	userOAuth.ExpireAt = time.Now().Add(utils.USER_REFRESH_TOKEN_IAT * time.Second).Unix()
	userOAuth.LoginAt = time.Now().Unix()
	if err := s.postgresRepo.PostgresOAuthRepo.UpdateOAuth(ctx, tx, userOAuth); err != nil {
		tx.WithContext(ctx).Rollback()
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}
	tx.WithContext(ctx).Commit()

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *userService) Register(ctx context.Context, data *model.RegisterRequest) (*model.RegisterResponse, error) {
	// Check user exited
	existedUser, err := s.postgresRepo.PostgresUserRepo.FindUsersByFilter(ctx, nil, &repository.FindUserByFilter{
		PhoneNumber: &data.PhoneNumber,
		Email:       &data.Email,
	})
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}
	if len(existedUser) > 0 {
		return nil, errors.New(errors.ErrCodeUserExisted)
	}

	// Create user
	tx := database.BeginPostgresTransaction()
	user := entity.NewUser()
	user.PhoneNumber = &data.PhoneNumber
	user.Email = &data.Email
	user.Password = data.Password
	if err := s.postgresRepo.PostgresUserRepo.CreateUser(ctx, tx, user); err != nil {
		tx.WithContext(ctx).Rollback()
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}
	tx.WithContext(ctx).Commit()

	return &model.RegisterResponse{}, nil
}

func (s *userService) Logout(ctx context.Context, data *model.LogoutRequest) (*model.LogoutResponse, error) {
	user := ctx.Value(utils.USER_CONTEXT_KEY).(entity.User)

	// Find User OAuth
	userOAuth, err := s.postgresRepo.PostgresOAuthRepo.FindOAuthByFilter(ctx, nil, &repository.FindOAuthByFilter{
		UserID: &user.ID,
	})
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}

	// Deactivate User OAuth
	tx := database.BeginPostgresTransaction()
	userOAuth.Status = entity.OAuthStatusInactive
	if err := s.postgresRepo.PostgresOAuthRepo.UpdateOAuth(ctx, tx, userOAuth); err != nil {
		tx.WithContext(ctx).Rollback()
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}
	tx.WithContext(ctx).Commit()

	return &model.LogoutResponse{}, nil
}
