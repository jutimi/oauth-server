package service

import (
	"oauth-server/app/helper"
	postgres_repository "oauth-server/app/repository/postgres"
	client_grpc "oauth-server/grpc/client"
)

type ServiceCollections struct {
	UserSvc  UserService
	OAuthSvc OAuthService
}

func RegisterServices(
	helpers helper.HelperCollections,
	clientGRPC client_grpc.ClientGRPCCollection,
	postgresRepo postgres_repository.PostgresRepositoryCollections,
) ServiceCollections {
	userSvc := NewUserService(helpers, postgresRepo)
	oauthSvc := NewOAuthService(helpers, postgresRepo)

	return ServiceCollections{
		UserSvc:  userSvc,
		OAuthSvc: oauthSvc,
	}
}
