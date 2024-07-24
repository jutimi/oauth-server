package client_grpc

import (
	context "context"
	"log"
	"oauth-server/config"
	"oauth-server/package/errors"

	"github.com/jutimi/grpc-service/common"
	"github.com/jutimi/grpc-service/workspace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type workspaceClient struct {
	conn         *grpc.ClientConn
	wsClient     workspace.WorkspaceRouteClient
	userWSClient workspace.UserWorkspaceRouteClient
}

type WorkspaceClient interface {
	GetWorkspaceById(ctx context.Context, data *common.GetByIdParams) (*workspace.WorkspaceResponse, error)
	GetUserWSById(ctx context.Context, data *common.GetByIdParams) (*workspace.UserWorkspaceResponse, error)
	GetUserWSByFilter(ctx context.Context, data *workspace.GetUserWorkspaceByFilterParams) (*workspace.UserWorkspaceResponse, error)
}

func NewWsClient() WorkspaceClient {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conf := config.GetConfiguration().GRPC

	// Connect to Workspace grpc server
	conn, err := grpc.NewClient(conf.WorkspaceGRPC, opts...)
	if err != nil {
		log.Fatalf("Error connect to Workspace grpc server: %s", err.Error())
	}
	wsClient := workspace.NewWorkspaceRouteClient(conn)
	userWSClient := workspace.NewUserWorkspaceRouteClient(conn)

	return &workspaceClient{
		conn:         conn,
		wsClient:     wsClient,
		userWSClient: userWSClient,
	}
}

func (c *workspaceClient) GetWorkspaceById(
	ctx context.Context,
	data *common.GetByIdParams,
) (*workspace.WorkspaceResponse, error) {
	resp, err := c.wsClient.GetWorkspaceById(ctx, data)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}
	if resp.Error != nil {
		return nil, errors.NewCustomError(int(resp.Error.ErrorCode), resp.Error.ErrorMessage)
	}
	if !resp.Success {
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}

	return resp, nil
}

func (c *workspaceClient) GetUserWSById(
	ctx context.Context,
	data *common.GetByIdParams,
) (*workspace.UserWorkspaceResponse, error) {
	resp, err := c.userWSClient.GetUserWorkspaceById(ctx, data)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}
	if resp.Error != nil {
		return nil, errors.NewCustomError(int(resp.Error.ErrorCode), resp.Error.ErrorMessage)
	}
	if !resp.Success {
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}

	return resp, nil
}

func (c *workspaceClient) GetUserWSByFilter(
	ctx context.Context,
	data *workspace.GetUserWorkspaceByFilterParams,
) (*workspace.UserWorkspaceResponse, error) {
	resp, err := c.userWSClient.GetUserWorkspaceByFilter(ctx, data)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}
	if resp.Error != nil {
		return nil, errors.NewCustomError(int(resp.Error.ErrorCode), resp.Error.ErrorMessage)
	}
	if !resp.Success {
		return nil, errors.New(errors.ErrCodeInternalServerError)
	}

	return resp, nil
}

func (c *workspaceClient) CloseConn() {
	c.conn.Close()
}