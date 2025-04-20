package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"social-network/post-service/models"
	pb "social-network/post-service/proto/post-service/pb"
	"social-network/post-service/repository"
)

type PostServiceServer struct {
	pb.UnimplementedPostServiceServer
	postRepo  *repository.PostRepository
	promoRepo *repository.PromoRepository
}

func NewPostServiceServer(postRepo *repository.PostRepository, promoRepo *repository.PromoRepository) *PostServiceServer {
	return &PostServiceServer{
		postRepo:  postRepo,
		promoRepo: promoRepo,
	}
}

func (s *PostServiceServer) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.Post, error) {
	post := &models.Post{
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   req.CreatorId,
		IsPrivate:   req.IsPrivate,
		Tags:        req.Tags,
	}

	if err := s.postRepo.CreatePost(post); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create post: %v", err)
	}

	return &pb.Post{
		Id:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		CreatorId:   post.CreatorID,
		CreatedAt:   timestamppb.New(post.CreatedAt),
		UpdatedAt:   timestamppb.New(post.UpdatedAt),
		IsPrivate:   post.IsPrivate,
		Tags:        post.Tags,
	}, nil
}

func (s *PostServiceServer) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.Post, error) {
	post, err := s.postRepo.GetPostByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Post not found: %v", err)
	}

	if post.IsPrivate && post.CreatorID != req.UserId {
		return nil, status.Errorf(codes.PermissionDenied, "This post is private")
	}

	return &pb.Post{
		Id:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		CreatorId:   post.CreatorID,
		CreatedAt:   timestamppb.New(post.CreatedAt),
		UpdatedAt:   timestamppb.New(post.UpdatedAt),
		IsPrivate:   post.IsPrivate,
		Tags:        post.Tags,
	}, nil
}

func (s *PostServiceServer) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.Post, error) {
	post := &models.Post{
		ID:          req.Id,
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   req.CreatorId,
		IsPrivate:   req.IsPrivate,
		Tags:        req.Tags,
	}

	if err := s.postRepo.UpdatePost(post); err != nil {
		if err.Error() == "post not found or you don't have permission to update it" {
			return nil, status.Errorf(codes.PermissionDenied, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "Failed to update post: %v", err)
	}

	updatedPost, err := s.postRepo.GetPostByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get updated post: %v", err)
	}

	return &pb.Post{
		Id:          updatedPost.ID,
		Title:       updatedPost.Title,
		Description: updatedPost.Description,
		CreatorId:   updatedPost.CreatorID,
		CreatedAt:   timestamppb.New(updatedPost.CreatedAt),
		UpdatedAt:   timestamppb.New(updatedPost.UpdatedAt),
		IsPrivate:   updatedPost.IsPrivate,
		Tags:        updatedPost.Tags,
	}, nil
}

func (s *PostServiceServer) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*emptypb.Empty, error) {
	if err := s.postRepo.DeletePost(req.Id, req.CreatorId); err != nil {
		if err.Error() == "post not found or you don't have permission to delete it" {
			return nil, status.Errorf(codes.PermissionDenied, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "Failed to delete post: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostServiceServer) ListPosts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	posts, total, err := s.postRepo.ListPosts(req.Page, req.PageSize, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list posts: %v", err)
	}

	var pbPosts []*pb.Post
	for _, post := range posts {
		pbPosts = append(pbPosts, &pb.Post{
			Id:          post.ID,
			Title:       post.Title,
			Description: post.Description,
			CreatorId:   post.CreatorID,
			CreatedAt:   timestamppb.New(post.CreatedAt),
			UpdatedAt:   timestamppb.New(post.UpdatedAt),
			IsPrivate:   post.IsPrivate,
			Tags:        post.Tags,
		})
	}

	return &pb.ListPostsResponse{
		Posts:      pbPosts,
		TotalCount: int32(total),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

func (s *PostServiceServer) CreatePromo(ctx context.Context, req *pb.CreatePromoRequest) (*pb.Promo, error) {
	promo := &models.Promo{
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   req.CreatorId,
		Discount:    req.Discount,
		Code:        req.Code,
	}

	if err := s.promoRepo.CreatePromo(promo); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create promo: %v", err)
	}

	return &pb.Promo{
		Id:          promo.ID,
		Title:       promo.Title,
		Description: promo.Description,
		CreatorId:   promo.CreatorID,
		Discount:    promo.Discount,
		Code:        promo.Code,
		CreatedAt:   timestamppb.New(promo.CreatedAt),
		UpdatedAt:   timestamppb.New(promo.UpdatedAt),
	}, nil
}

func (s *PostServiceServer) GetPromo(ctx context.Context, req *pb.GetPromoRequest) (*pb.Promo, error) {
	promo, err := s.promoRepo.GetPromoByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Promo not found: %v", err)
	}

	return &pb.Promo{
		Id:          promo.ID,
		Title:       promo.Title,
		Description: promo.Description,
		CreatorId:   promo.CreatorID,
		Discount:    promo.Discount,
		Code:        promo.Code,
		CreatedAt:   timestamppb.New(promo.CreatedAt),
		UpdatedAt:   timestamppb.New(promo.UpdatedAt),
	}, nil
}

func (s *PostServiceServer) UpdatePromo(ctx context.Context, req *pb.UpdatePromoRequest) (*pb.Promo, error) {
	promo := &models.Promo{
		ID:          req.Id,
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   req.CreatorId,
		Discount:    req.Discount,
		Code:        req.Code,
	}

	if err := s.promoRepo.UpdatePromo(promo); err != nil {
		if err.Error() == "promo not found or you don't have permission to update it" {
			return nil, status.Errorf(codes.PermissionDenied, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "Failed to update promo: %v", err)
	}

	updatedPromo, err := s.promoRepo.GetPromoByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get updated promo: %v", err)
	}

	return &pb.Promo{
		Id:          updatedPromo.ID,
		Title:       updatedPromo.Title,
		Description: updatedPromo.Description,
		CreatorId:   updatedPromo.CreatorID,
		Discount:    updatedPromo.Discount,
		Code:        updatedPromo.Code,
		CreatedAt:   timestamppb.New(updatedPromo.CreatedAt),
		UpdatedAt:   timestamppb.New(updatedPromo.UpdatedAt),
	}, nil
}

func (s *PostServiceServer) DeletePromo(ctx context.Context, req *pb.DeletePromoRequest) (*emptypb.Empty, error) {
	if err := s.promoRepo.DeletePromo(req.Id, req.CreatorId); err != nil {
		if err.Error() == "promo not found or you don't have permission to delete it" {
			return nil, status.Errorf(codes.PermissionDenied, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "Failed to delete promo: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PostServiceServer) ListPromos(ctx context.Context, req *pb.ListPromosRequest) (*pb.ListPromosResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	promos, total, err := s.promoRepo.ListPromos(req.Page, req.PageSize, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list promos: %v", err)
	}

	var pbPromos []*pb.Promo
	for _, promo := range promos {
		pbPromos = append(pbPromos, &pb.Promo{
			Id:          promo.ID,
			Title:       promo.Title,
			Description: promo.Description,
			CreatorId:   promo.CreatorID,
			Discount:    promo.Discount,
			Code:        promo.Code,
			CreatedAt:   timestamppb.New(promo.CreatedAt),
			UpdatedAt:   timestamppb.New(promo.UpdatedAt),
		})
	}

	return &pb.ListPromosResponse{
		Promos:     pbPromos,
		TotalCount: int32(total),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}
