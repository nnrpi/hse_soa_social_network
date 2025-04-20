package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "social-network/post-service/proto/post-service/pb"
)

type PostServiceClient struct {
	client pb.PostServiceClient
	conn   *grpc.ClientConn
}

func NewPostServiceClient(address string) (*PostServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewPostServiceClient(conn)
	return &PostServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *PostServiceClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *PostServiceClient) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.Post, error) {
	return c.client.CreatePost(ctx, req)
}

func (c *PostServiceClient) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.Post, error) {
	return c.client.GetPost(ctx, req)
}

func (c *PostServiceClient) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.Post, error) {
	return c.client.UpdatePost(ctx, req)
}

func (c *PostServiceClient) DeletePost(ctx context.Context, req *pb.DeletePostRequest) error {
	_, err := c.client.DeletePost(ctx, req)
	return err
}

func (c *PostServiceClient) ListPosts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	return c.client.ListPosts(ctx, req)
}

func (c *PostServiceClient) CreatePromo(ctx context.Context, req *pb.CreatePromoRequest) (*pb.Promo, error) {
	return c.client.CreatePromo(ctx, req)
}

func (c *PostServiceClient) GetPromo(ctx context.Context, req *pb.GetPromoRequest) (*pb.Promo, error) {
	return c.client.GetPromo(ctx, req)
}

func (c *PostServiceClient) UpdatePromo(ctx context.Context, req *pb.UpdatePromoRequest) (*pb.Promo, error) {
	return c.client.UpdatePromo(ctx, req)
}

func (c *PostServiceClient) DeletePromo(ctx context.Context, req *pb.DeletePromoRequest) error {
	_, err := c.client.DeletePromo(ctx, req)
	return err
}

func (c *PostServiceClient) ListPromos(ctx context.Context, req *pb.ListPromosRequest) (*pb.ListPromosResponse, error) {
	return c.client.ListPromos(ctx, req)
}
