package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// NeighborhoodPostType defines types of community posts
type NeighborhoodPostType string

const (
	PostTypeQuestion       NeighborhoodPostType = "question"
	PostTypeRecommendation NeighborhoodPostType = "recommendation"
	PostTypeHelp           NeighborhoodPostType = "help"
	PostTypeEvent          NeighborhoodPostType = "event"
)

// NeighborhoodPost represents a community question/post
type NeighborhoodPost struct {
	ID          uuid.UUID            `json:"id"`
	UserID      uuid.UUID            `json:"user_id"`
	PostType    NeighborhoodPostType `json:"post_type"`
	Title       string               `json:"title"`
	Content     *string              `json:"content,omitempty"`
	LocationLat *float64             `json:"location_lat,omitempty"`
	LocationLng *float64             `json:"location_lng,omitempty"`
	City        *string              `json:"city,omitempty"`
	RadiusKm    int                  `json:"radius_km"`
	IsResolved  bool                 `json:"is_resolved"`
	Upvotes     int                  `json:"upvotes"`
	AnswerCount int                  `json:"answer_count"`
	CreatedAt   time.Time            `json:"created_at"`
	User        *UserResponse        `json:"user,omitempty"`
	Answers     []NeighborhoodAnswer `json:"answers,omitempty"`
	HasUpvoted  bool                 `json:"has_upvoted,omitempty"` // For current user
}

// NeighborhoodAnswer represents an answer to a post
type NeighborhoodAnswer struct {
	ID         uuid.UUID     `json:"id"`
	PostID     uuid.UUID     `json:"post_id"`
	UserID     uuid.UUID     `json:"user_id"`
	Content    string        `json:"content"`
	IsAccepted bool          `json:"is_accepted"`
	Upvotes    int           `json:"upvotes"`
	CreatedAt  time.Time     `json:"created_at"`
	User       *UserResponse `json:"user,omitempty"`
	HasUpvoted bool          `json:"has_upvoted,omitempty"`
}

// CreateNeighborhoodPostParams for creating a post
type CreateNeighborhoodPostParams struct {
	UserID      uuid.UUID
	PostType    NeighborhoodPostType
	Title       string
	Content     *string
	LocationLat *float64
	LocationLng *float64
	City        *string
	RadiusKm    int
}

// NeighborhoodRepository interface
type NeighborhoodRepository interface {
	CreatePost(ctx context.Context, params CreateNeighborhoodPostParams) (*NeighborhoodPost, error)
	GetPost(ctx context.Context, postID uuid.UUID) (*NeighborhoodPost, error)
	GetPostsByCity(ctx context.Context, city string, postType *string, limit, offset int) ([]NeighborhoodPost, error)
	GetPostsByLocation(ctx context.Context, lat, lng float64, radiusKm int, limit, offset int) ([]NeighborhoodPost, error)
	UpdatePost(ctx context.Context, postID uuid.UUID, isResolved bool) error
	DeletePost(ctx context.Context, postID uuid.UUID) error

	// Answers
	CreateAnswer(ctx context.Context, postID, userID uuid.UUID, content string) (*NeighborhoodAnswer, error)
	GetAnswers(ctx context.Context, postID uuid.UUID) ([]NeighborhoodAnswer, error)
	AcceptAnswer(ctx context.Context, answerID uuid.UUID) error

	// Upvotes
	UpvotePost(ctx context.Context, postID, userID uuid.UUID) error
	RemovePostUpvote(ctx context.Context, postID, userID uuid.UUID) error
	UpvoteAnswer(ctx context.Context, answerID, userID uuid.UUID) error
	HasUpvoted(ctx context.Context, postID, userID uuid.UUID) (bool, error)
}

// NeighborhoodService handles community Q&A
type NeighborhoodService struct {
	repo         NeighborhoodRepository
	karmaService *KarmaService
}

// NewNeighborhoodService creates a new neighborhood service
func NewNeighborhoodService(repo NeighborhoodRepository, karmaService *KarmaService) *NeighborhoodService {
	return &NeighborhoodService{
		repo:         repo,
		karmaService: karmaService,
	}
}

// CreatePost creates a new neighborhood post
func (s *NeighborhoodService) CreatePost(ctx context.Context, params CreateNeighborhoodPostParams) (*NeighborhoodPost, error) {
	post, err := s.repo.CreatePost(ctx, params)
	if err != nil {
		return nil, err
	}

	// Award karma for asking/helping
	refType := "neighborhood"
	_ = s.karmaService.EarnKarma(ctx, params.UserID, "neighborhood_post", &refType, &post.ID)

	return post, nil
}

// GetNearbyPosts returns posts near a location
func (s *NeighborhoodService) GetNearbyPosts(ctx context.Context, city string, postType *string, limit, offset int) ([]NeighborhoodPost, error) {
	return s.repo.GetPostsByCity(ctx, city, postType, limit, offset)
}

// AnswerPost adds an answer to a post
func (s *NeighborhoodService) AnswerPost(ctx context.Context, postID, userID uuid.UUID, content string) (*NeighborhoodAnswer, error) {
	answer, err := s.repo.CreateAnswer(ctx, postID, userID, content)
	if err != nil {
		return nil, err
	}

	// Award karma for helping
	refType := "neighborhood_answer"
	_ = s.karmaService.EarnKarma(ctx, userID, "neighborhood_answer", &refType, &answer.ID)

	return answer, nil
}

// AcceptAnswer marks an answer as accepted (by post owner)
func (s *NeighborhoodService) AcceptAnswer(ctx context.Context, userID, answerID uuid.UUID) error {
	// TODO: Verify user owns the post
	if err := s.repo.AcceptAnswer(ctx, answerID); err != nil {
		return err
	}

	// Award bonus karma to answer author
	// Would need to fetch answer to get author ID
	return nil
}

// UpvotePost upvotes a post
func (s *NeighborhoodService) UpvotePost(ctx context.Context, postID, userID uuid.UUID) error {
	return s.repo.UpvotePost(ctx, postID, userID)
}
