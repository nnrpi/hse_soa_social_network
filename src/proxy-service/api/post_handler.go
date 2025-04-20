package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"social-network/post-service/client"
	pb "social-network/post-service/proto/post-service/pb"
	"social-network/user-service/service"
)

type PostHandler struct {
	postClient  *client.PostServiceClient
	userService *service.UserService
}

func NewPostHandler(postClient *client.PostServiceClient, userService *service.UserService) *PostHandler {
	return &PostHandler{
		postClient:  postClient,
		userService: userService,
	}
}

func (h *PostHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/posts", h.SessionAuthMiddleware(h.CreatePost)).Methods("POST")
	router.HandleFunc("/posts/{id}", h.GetPost).Methods("GET")
	router.HandleFunc("/posts/{id}", h.SessionAuthMiddleware(h.UpdatePost)).Methods("PUT")
	router.HandleFunc("/posts/{id}", h.SessionAuthMiddleware(h.DeletePost)).Methods("DELETE")
	router.HandleFunc("/posts", h.SessionAuthMiddleware(h.ListPosts)).Methods("GET")

	router.HandleFunc("/promos", h.SessionAuthMiddleware(h.CreatePromo)).Methods("POST")
	router.HandleFunc("/promos/{id}", h.SessionAuthMiddleware(h.GetPromo)).Methods("GET")
	router.HandleFunc("/promos/{id}", h.SessionAuthMiddleware(h.UpdatePromo)).Methods("PUT")
	router.HandleFunc("/promos/{id}", h.SessionAuthMiddleware(h.DeletePromo)).Methods("DELETE")
	router.HandleFunc("/promos", h.SessionAuthMiddleware(h.ListPromos)).Methods("GET")
}

func (h *PostHandler) SessionAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthorized: No session token", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Error reading cookie", http.StatusBadRequest)
			return
		}

		session, err := h.userService.ValidateSession(cookie.Value)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		user, err := h.userService.GetUserByUsername(session.Username)
		if err != nil {
			http.Error(w, "Error getting user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", user.ID)
		ctx = context.WithValue(ctx, "username", user.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	var reqBody struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		IsPrivate   bool     `json:"is_private"`
		Tags        []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if reqBody.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if reqBody.Description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}

	req := &pb.CreatePostRequest{
		Title:       reqBody.Title,
		Description: reqBody.Description,
		CreatorId:   userID,
		IsPrivate:   reqBody.IsPrivate,
		Tags:        reqBody.Tags,
	}

	post, err := h.postClient.CreatePost(r.Context(), req)
	if err != nil {
		http.Error(w, "Error creating post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapPostToJSON(post))
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var userID int64 = 0
	if cookie, err := r.Cookie("session_token"); err == nil {
		if session, err := h.userService.ValidateSession(cookie.Value); err == nil {
			if user, err := h.userService.GetUserByUsername(session.Username); err == nil {
				userID = user.ID
			}
		}
	}

	req := &pb.GetPostRequest{
		Id:     id,
		UserId: userID,
	}

	post, err := h.postClient.GetPost(r.Context(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				http.Error(w, "Post not found", http.StatusNotFound)
				return
			case codes.PermissionDenied:
				http.Error(w, "You don't have permission to view this post", http.StatusForbidden)
				return
			default:
				http.Error(w, "Error getting post: "+st.Message(), http.StatusInternalServerError)
				return
			}
		}
		http.Error(w, "Error getting post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapPostToJSON(post))
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := r.Context().Value("user_id").(int64)

	var reqBody struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		IsPrivate   bool     `json:"is_private"`
		Tags        []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if reqBody.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if reqBody.Description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}

	req := &pb.UpdatePostRequest{
		Id:          id,
		Title:       reqBody.Title,
		Description: reqBody.Description,
		CreatorId:   userID,
		IsPrivate:   reqBody.IsPrivate,
		Tags:        reqBody.Tags,
	}

	post, err := h.postClient.UpdatePost(r.Context(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.PermissionDenied:
				http.Error(w, "You don't have permission to update this post", http.StatusForbidden)
				return
			default:
				http.Error(w, "Error updating post: "+st.Message(), http.StatusInternalServerError)
				return
			}
		}
		http.Error(w, "Error updating post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapPostToJSON(post))
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := r.Context().Value("user_id").(int64)

	req := &pb.DeletePostRequest{
		Id:        id,
		CreatorId: userID,
	}

	err := h.postClient.DeletePost(r.Context(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.PermissionDenied:
				http.Error(w, "You don't have permission to delete this post", http.StatusForbidden)
				return
			default:
				http.Error(w, "Error deleting post: "+st.Message(), http.StatusInternalServerError)
				return
			}
		}
		http.Error(w, "Error deleting post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	var page int32 = 1
	var pageSize int32 = 10

	if pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err == nil && pageInt > 0 {
			page = int32(pageInt)
		}
	}

	if pageSizeStr != "" {
		pageSizeInt, err := strconv.Atoi(pageSizeStr)
		if err == nil && pageSizeInt > 0 && pageSizeInt <= 100 {
			pageSize = int32(pageSizeInt)
		}
	}

	req := &pb.ListPostsRequest{
		Page:     page,
		PageSize: pageSize,
		UserId:   userID,
	}

	resp, err := h.postClient.ListPosts(r.Context(), req)
	if err != nil {
		http.Error(w, "Error listing posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	posts := make([]map[string]interface{}, 0, len(resp.Posts))
	for _, post := range resp.Posts {
		posts = append(posts, mapPostToJSON(post))
	}

	result := map[string]interface{}{
		"posts":       posts,
		"total_count": resp.TotalCount,
		"page":        resp.Page,
		"page_size":   resp.PageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *PostHandler) CreatePromo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	var reqBody struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Discount    float64 `json:"discount"`
		Code        string  `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if reqBody.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if reqBody.Description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}
	if reqBody.Code == "" {
		http.Error(w, "Code is required", http.StatusBadRequest)
		return
	}
	if reqBody.Discount <= 0 || reqBody.Discount > 100 {
		http.Error(w, "Discount must be between 0 and 100", http.StatusBadRequest)
		return
	}

	req := &pb.CreatePromoRequest{
		Title:       reqBody.Title,
		Description: reqBody.Description,
		CreatorId:   userID,
		Discount:    reqBody.Discount,
		Code:        reqBody.Code,
	}

	promo, err := h.postClient.CreatePromo(r.Context(), req)
	if err != nil {
		http.Error(w, "Error creating promo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapPromoToJSON(promo))
}

func (h *PostHandler) GetPromo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := r.Context().Value("user_id").(int64)

	req := &pb.GetPromoRequest{
		Id:     id,
		UserId: userID,
	}

	promo, err := h.postClient.GetPromo(r.Context(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				http.Error(w, "Promo not found", http.StatusNotFound)
				return
			default:
				http.Error(w, "Error getting promo: "+st.Message(), http.StatusInternalServerError)
				return
			}
		}
		http.Error(w, "Error getting promo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapPromoToJSON(promo))
}

func (h *PostHandler) UpdatePromo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := r.Context().Value("user_id").(int64)

	var reqBody struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Discount    float64 `json:"discount"`
		Code        string  `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if reqBody.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if reqBody.Description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}
	if reqBody.Code == "" {
		http.Error(w, "Code is required", http.StatusBadRequest)
		return
	}
	if reqBody.Discount <= 0 || reqBody.Discount > 100 {
		http.Error(w, "Discount must be between 0 and 100", http.StatusBadRequest)
		return
	}

	req := &pb.UpdatePromoRequest{
		Id:          id,
		Title:       reqBody.Title,
		Description: reqBody.Description,
		CreatorId:   userID,
		Discount:    reqBody.Discount,
		Code:        reqBody.Code,
	}

	promo, err := h.postClient.UpdatePromo(r.Context(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.PermissionDenied:
				http.Error(w, "You don't have permission to update this promo", http.StatusForbidden)
				return
			default:
				http.Error(w, "Error updating promo: "+st.Message(), http.StatusInternalServerError)
				return
			}
		}
		http.Error(w, "Error updating promo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapPromoToJSON(promo))
}

func (h *PostHandler) DeletePromo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := r.Context().Value("user_id").(int64)

	req := &pb.DeletePromoRequest{
		Id:        id,
		CreatorId: userID,
	}

	err := h.postClient.DeletePromo(r.Context(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.PermissionDenied:
				http.Error(w, "You don't have permission to delete this promo", http.StatusForbidden)
				return
			default:
				http.Error(w, "Error deleting promo: "+st.Message(), http.StatusInternalServerError)
				return
			}
		}
		http.Error(w, "Error deleting promo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PostHandler) ListPromos(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	var page int32 = 1
	var pageSize int32 = 10

	if pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err == nil && pageInt > 0 {
			page = int32(pageInt)
		}
	}

	if pageSizeStr != "" {
		pageSizeInt, err := strconv.Atoi(pageSizeStr)
		if err == nil && pageSizeInt > 0 && pageSizeInt <= 100 {
			pageSize = int32(pageSizeInt)
		}
	}

	req := &pb.ListPromosRequest{
		Page:     page,
		PageSize: pageSize,
		UserId:   userID,
	}

	resp, err := h.postClient.ListPromos(r.Context(), req)
	if err != nil {
		http.Error(w, "Error listing promos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	promos := make([]map[string]interface{}, 0, len(resp.Promos))
	for _, promo := range resp.Promos {
		promos = append(promos, mapPromoToJSON(promo))
	}

	result := map[string]interface{}{
		"promos":      promos,
		"total_count": resp.TotalCount,
		"page":        resp.Page,
		"page_size":   resp.PageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func mapPostToJSON(post *pb.Post) map[string]interface{} {
	return map[string]interface{}{
		"id":          post.Id,
		"title":       post.Title,
		"description": post.Description,
		"creator_id":  post.CreatorId,
		"created_at":  post.CreatedAt.AsTime().Format(time.RFC3339),
		"updated_at":  post.UpdatedAt.AsTime().Format(time.RFC3339),
		"is_private":  post.IsPrivate,
		"tags":        post.Tags,
	}
}

func mapPromoToJSON(promo *pb.Promo) map[string]interface{} {
	return map[string]interface{}{
		"id":          promo.Id,
		"title":       promo.Title,
		"description": promo.Description,
		"creator_id":  promo.CreatorId,
		"discount":    promo.Discount,
		"code":        promo.Code,
		"created_at":  promo.CreatedAt.AsTime().Format(time.RFC3339),
		"updated_at":  promo.UpdatedAt.AsTime().Format(time.RFC3339),
	}
}
