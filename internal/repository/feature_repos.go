package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/locolive/backend/internal/domain"
)

// NeighborhoodRepository implements domain.NeighborhoodRepository
type NeighborhoodRepository struct {
	db *pgxpool.Pool
}

func NewNeighborhoodRepository(db *pgxpool.Pool) *NeighborhoodRepository {
	return &NeighborhoodRepository{db: db}
}

func (r *NeighborhoodRepository) CreatePost(ctx context.Context, params domain.CreateNeighborhoodPostParams) (*domain.NeighborhoodPost, error) {
	query := `
		INSERT INTO neighborhood_posts (user_id, post_type, title, content, location_lat, location_lng, city, radius_km)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, post_type, title, content, location_lat, location_lng, city, radius_km, is_resolved, upvotes, answer_count, created_at, updated_at
	`
	post := &domain.NeighborhoodPost{}
	err := r.db.QueryRow(ctx, query,
		params.UserID, params.PostType, params.Title, params.Content,
		params.LocationLat, params.LocationLng, params.City, params.RadiusKm,
	).Scan(
		&post.ID, &post.UserID, &post.PostType, &post.Title, &post.Content,
		&post.LocationLat, &post.LocationLng, &post.City, &post.RadiusKm,
		&post.IsResolved, &post.Upvotes, &post.AnswerCount, &post.CreatedAt, &post.CreatedAt,
	)
	return post, err
}

func (r *NeighborhoodRepository) GetPost(ctx context.Context, postID uuid.UUID) (*domain.NeighborhoodPost, error) {
	query := `
		SELECT np.id, np.user_id, np.post_type, np.title, np.content, np.city, np.radius_km,
		       np.is_resolved, np.upvotes, np.answer_count, np.created_at
		FROM neighborhood_posts np WHERE np.id = $1
	`
	post := &domain.NeighborhoodPost{}
	err := r.db.QueryRow(ctx, query, postID).Scan(
		&post.ID, &post.UserID, &post.PostType, &post.Title, &post.Content, &post.City,
		&post.RadiusKm, &post.IsResolved, &post.Upvotes, &post.AnswerCount, &post.CreatedAt,
	)
	return post, err
}

func (r *NeighborhoodRepository) GetPostsByCity(ctx context.Context, city string, postType *string, limit, offset int) ([]domain.NeighborhoodPost, error) {
	query := `
		SELECT np.id, np.user_id, np.post_type, np.title, np.content, np.city, np.radius_km,
		       np.is_resolved, np.upvotes, np.answer_count, np.created_at,
		       u.id, u.name, u.avatar_url
		FROM neighborhood_posts np
		JOIN users u ON np.user_id = u.id
		WHERE ($1 = '' OR np.city = $1)
		  AND ($2::text IS NULL OR np.post_type = $2)
		ORDER BY np.created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, city, postType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []domain.NeighborhoodPost
	for rows.Next() {
		var post domain.NeighborhoodPost
		var user domain.UserResponse
		if err := rows.Scan(
			&post.ID, &post.UserID, &post.PostType, &post.Title, &post.Content, &post.City,
			&post.RadiusKm, &post.IsResolved, &post.Upvotes, &post.AnswerCount, &post.CreatedAt,
			&user.ID, &user.Name, &user.AvatarURL,
		); err != nil {
			return nil, err
		}
		post.User = &user
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *NeighborhoodRepository) GetPostsByLocation(ctx context.Context, lat, lng float64, radiusKm int, limit, offset int) ([]domain.NeighborhoodPost, error) {
	// For now, fallback to city-based query
	return r.GetPostsByCity(ctx, "", nil, limit, offset)
}

func (r *NeighborhoodRepository) UpdatePost(ctx context.Context, postID uuid.UUID, isResolved bool) error {
	query := `UPDATE neighborhood_posts SET is_resolved = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, isResolved, postID)
	return err
}

func (r *NeighborhoodRepository) DeletePost(ctx context.Context, postID uuid.UUID) error {
	query := `DELETE FROM neighborhood_posts WHERE id = $1`
	_, err := r.db.Exec(ctx, query, postID)
	return err
}

func (r *NeighborhoodRepository) CreateAnswer(ctx context.Context, postID, userID uuid.UUID, content string) (*domain.NeighborhoodAnswer, error) {
	query := `
		INSERT INTO neighborhood_answers (post_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, post_id, user_id, content, is_accepted, upvotes, created_at
	`
	answer := &domain.NeighborhoodAnswer{}
	err := r.db.QueryRow(ctx, query, postID, userID, content).Scan(
		&answer.ID, &answer.PostID, &answer.UserID, &answer.Content,
		&answer.IsAccepted, &answer.Upvotes, &answer.CreatedAt,
	)
	if err == nil {
		// Update answer count
		r.db.Exec(ctx, `UPDATE neighborhood_posts SET answer_count = answer_count + 1 WHERE id = $1`, postID)
	}
	return answer, err
}

func (r *NeighborhoodRepository) GetAnswers(ctx context.Context, postID uuid.UUID) ([]domain.NeighborhoodAnswer, error) {
	query := `
		SELECT na.id, na.post_id, na.user_id, na.content, na.is_accepted, na.upvotes, na.created_at,
		       u.id, u.name, u.avatar_url
		FROM neighborhood_answers na
		JOIN users u ON na.user_id = u.id
		WHERE na.post_id = $1
		ORDER BY na.is_accepted DESC, na.upvotes DESC
	`
	rows, err := r.db.Query(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []domain.NeighborhoodAnswer
	for rows.Next() {
		var answer domain.NeighborhoodAnswer
		var user domain.UserResponse
		if err := rows.Scan(
			&answer.ID, &answer.PostID, &answer.UserID, &answer.Content,
			&answer.IsAccepted, &answer.Upvotes, &answer.CreatedAt,
			&user.ID, &user.Name, &user.AvatarURL,
		); err != nil {
			return nil, err
		}
		answer.User = &user
		answers = append(answers, answer)
	}
	return answers, nil
}

func (r *NeighborhoodRepository) AcceptAnswer(ctx context.Context, answerID uuid.UUID) error {
	query := `UPDATE neighborhood_answers SET is_accepted = TRUE WHERE id = $1`
	_, err := r.db.Exec(ctx, query, answerID)
	return err
}

func (r *NeighborhoodRepository) UpvotePost(ctx context.Context, postID, userID uuid.UUID) error {
	// Insert upvote and increment count
	query := `
		INSERT INTO neighborhood_upvotes (user_id, post_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	result, err := r.db.Exec(ctx, query, userID, postID)
	if err != nil {
		return err
	}
	if result.RowsAffected() > 0 {
		r.db.Exec(ctx, `UPDATE neighborhood_posts SET upvotes = upvotes + 1 WHERE id = $1`, postID)
	}
	return nil
}

func (r *NeighborhoodRepository) RemovePostUpvote(ctx context.Context, postID, userID uuid.UUID) error {
	query := `DELETE FROM neighborhood_upvotes WHERE user_id = $1 AND post_id = $2`
	result, err := r.db.Exec(ctx, query, userID, postID)
	if err == nil && result.RowsAffected() > 0 {
		r.db.Exec(ctx, `UPDATE neighborhood_posts SET upvotes = upvotes - 1 WHERE id = $1 AND upvotes > 0`, postID)
	}
	return err
}

func (r *NeighborhoodRepository) UpvoteAnswer(ctx context.Context, answerID, userID uuid.UUID) error {
	query := `
		INSERT INTO neighborhood_upvotes (user_id, answer_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	result, err := r.db.Exec(ctx, query, userID, answerID)
	if err == nil && result.RowsAffected() > 0 {
		r.db.Exec(ctx, `UPDATE neighborhood_answers SET upvotes = upvotes + 1 WHERE id = $1`, answerID)
	}
	return err
}

func (r *NeighborhoodRepository) HasUpvoted(ctx context.Context, postID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM neighborhood_upvotes WHERE user_id = $1 AND post_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, userID, postID).Scan(&exists)
	return exists, err
}

// PrivacyRepository implements domain.PrivacyRepository
type PrivacyRepository struct {
	db *pgxpool.Pool
}

func NewPrivacyRepository(db *pgxpool.Pool) *PrivacyRepository {
	return &PrivacyRepository{db: db}
}

func (r *PrivacyRepository) GetPrivacySettings(ctx context.Context, userID uuid.UUID) (*domain.PrivacySettings, error) {
	query := `
		SELECT user_id, default_story_visibility, show_in_discovery, show_location,
		       who_can_message, incognito_mode
		FROM user_privacy_settings WHERE user_id = $1
	`
	settings := &domain.PrivacySettings{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&settings.UserID, &settings.DefaultStoryVisibility, &settings.ShowInDiscovery,
		&settings.ShowLocation, &settings.WhoCanMessage, &settings.IncognitoMode,
	)
	return settings, err
}

func (r *PrivacyRepository) CreatePrivacySettings(ctx context.Context, userID uuid.UUID) (*domain.PrivacySettings, error) {
	query := `
		INSERT INTO user_privacy_settings (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
		RETURNING user_id, default_story_visibility, show_in_discovery, show_location, who_can_message, incognito_mode
	`
	settings := &domain.PrivacySettings{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&settings.UserID, &settings.DefaultStoryVisibility, &settings.ShowInDiscovery,
		&settings.ShowLocation, &settings.WhoCanMessage, &settings.IncognitoMode,
	)
	if err != nil {
		// Row already exists, fetch it
		return r.GetPrivacySettings(ctx, userID)
	}
	return settings, nil
}

func (r *PrivacyRepository) UpdatePrivacySettings(ctx context.Context, settings *domain.PrivacySettings) error {
	query := `
		UPDATE user_privacy_settings
		SET default_story_visibility = $1, show_in_discovery = $2, show_location = $3,
		    who_can_message = $4, incognito_mode = $5, updated_at = NOW()
		WHERE user_id = $6
	`
	_, err := r.db.Exec(ctx, query,
		settings.DefaultStoryVisibility, settings.ShowInDiscovery, settings.ShowLocation,
		settings.WhoCanMessage, settings.IncognitoMode, settings.UserID,
	)
	return err
}

func (r *PrivacyRepository) BlockUser(ctx context.Context, userID, blockedID uuid.UUID, reason *string) error {
	query := `
		INSERT INTO blocked_users (user_id, blocked_user_id, reason)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, blocked_user_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, userID, blockedID, reason)
	return err
}

func (r *PrivacyRepository) UnblockUser(ctx context.Context, userID, blockedID uuid.UUID) error {
	query := `DELETE FROM blocked_users WHERE user_id = $1 AND blocked_user_id = $2`
	_, err := r.db.Exec(ctx, query, userID, blockedID)
	return err
}

func (r *PrivacyRepository) GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT blocked_user_id FROM blocked_users WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *PrivacyRepository) IsBlocked(ctx context.Context, userID, otherUserID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM blocked_users WHERE user_id = $1 AND blocked_user_id = $2)`
	var blocked bool
	err := r.db.QueryRow(ctx, query, userID, otherUserID).Scan(&blocked)
	return blocked, err
}

func (r *PrivacyRepository) AddCloseFriend(ctx context.Context, userID, friendID uuid.UUID, listType string) error {
	query := `
		INSERT INTO close_friends (user_id, friend_id, list_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, friend_id) DO UPDATE SET list_type = $3
	`
	_, err := r.db.Exec(ctx, query, userID, friendID, listType)
	return err
}

func (r *PrivacyRepository) RemoveCloseFriend(ctx context.Context, userID, friendID uuid.UUID) error {
	query := `DELETE FROM close_friends WHERE user_id = $1 AND friend_id = $2`
	_, err := r.db.Exec(ctx, query, userID, friendID)
	return err
}

func (r *PrivacyRepository) GetCloseFriends(ctx context.Context, userID uuid.UUID, listType string) ([]uuid.UUID, error) {
	query := `SELECT friend_id FROM close_friends WHERE user_id = $1 AND list_type = $2`
	rows, err := r.db.Query(ctx, query, userID, listType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *PrivacyRepository) IsCloseFriend(ctx context.Context, userID, friendID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM close_friends WHERE user_id = $1 AND friend_id = $2)`
	var isFriend bool
	err := r.db.QueryRow(ctx, query, userID, friendID).Scan(&isFriend)
	return isFriend, err
}

// ChallengeRepository implements domain.ChallengeRepository
type ChallengeRepository struct {
	db *pgxpool.Pool
}

func NewChallengeRepository(db *pgxpool.Pool) *ChallengeRepository {
	return &ChallengeRepository{db: db}
}

func (r *ChallengeRepository) GetUserChallenges(ctx context.Context, userID uuid.UUID, date string) ([]domain.Challenge, error) {
	query := `
		SELECT id, user_id, challenge_type, title, title_hindi, description, reward_karma,
		       is_completed, progress, target, assigned_date::text, completed_at, expires_at
		FROM daily_challenges
		WHERE user_id = $1 AND assigned_date = $2::date
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []domain.Challenge
	for rows.Next() {
		var c domain.Challenge
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.Type, &c.Title, &c.TitleHindi, &c.Description,
			&c.RewardKarma, &c.IsCompleted, &c.Progress, &c.Target, &c.AssignedAt,
			&c.CompletedAt, &c.ExpiresAt,
		); err != nil {
			return nil, err
		}
		challenges = append(challenges, c)
	}
	return challenges, nil
}

func (r *ChallengeRepository) CreateChallenge(ctx context.Context, c *domain.Challenge) error {
	query := `
		INSERT INTO daily_challenges (id, user_id, challenge_type, title, title_hindi, description, reward_karma, target, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		c.ID, c.UserID, c.Type, c.Title, c.TitleHindi, c.Description, c.RewardKarma, c.Target, c.ExpiresAt,
	)
	return err
}

func (r *ChallengeRepository) UpdateChallengeProgress(ctx context.Context, challengeID uuid.UUID, progress int) error {
	query := `UPDATE daily_challenges SET progress = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, progress, challengeID)
	return err
}

func (r *ChallengeRepository) CompleteChallenge(ctx context.Context, challengeID uuid.UUID) error {
	query := `UPDATE daily_challenges SET is_completed = TRUE, completed_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, challengeID)
	return err
}

func (r *ChallengeRepository) GetActiveTemplates(ctx context.Context) ([]domain.ChallengeTemplate, error) {
	query := `
		SELECT id, type, title, title_hindi, description, description_hindi, reward_karma, target, difficulty
		FROM challenge_templates WHERE is_active = TRUE
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []domain.ChallengeTemplate
	for rows.Next() {
		var t domain.ChallengeTemplate
		if err := rows.Scan(&t.ID, &t.Type, &t.Title, &t.TitleHindi, &t.Description, &t.DescriptionHindi, &t.RewardKarma, &t.Target, &t.Difficulty); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *ChallengeRepository) GetRandomTemplates(ctx context.Context, count int, excludeTypes []string) ([]domain.ChallengeTemplate, error) {
	query := `
		SELECT id, type, title, title_hindi, description, description_hindi, reward_karma, target, difficulty
		FROM challenge_templates WHERE is_active = TRUE
		ORDER BY RANDOM() LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []domain.ChallengeTemplate
	for rows.Next() {
		var t domain.ChallengeTemplate
		if err := rows.Scan(&t.ID, &t.Type, &t.Title, &t.TitleHindi, &t.Description, &t.DescriptionHindi, &t.RewardKarma, &t.Target, &t.Difficulty); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *ChallengeRepository) GetActiveFestival(ctx context.Context) (*domain.FestivalEvent, error) {
	query := `
		SELECT id, name, name_hindi, description, karma_multiplier, bonus_karma, start_date, end_date, is_active
		FROM festival_events
		WHERE is_active = TRUE AND start_date <= NOW() AND end_date >= NOW()
		LIMIT 1
	`
	festival := &domain.FestivalEvent{}
	err := r.db.QueryRow(ctx, query).Scan(
		&festival.ID, &festival.Name, &festival.NameHindi, &festival.Description,
		&festival.KarmaMultiplier, &festival.BonusKarma, &festival.StartDate, &festival.EndDate, &festival.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return festival, nil
}

func (r *ChallengeRepository) GetUpcomingFestivals(ctx context.Context, limit int) ([]domain.FestivalEvent, error) {
	query := `
		SELECT id, name, name_hindi, description, karma_multiplier, bonus_karma, start_date, end_date
		FROM festival_events
		WHERE is_active = TRUE AND start_date > NOW()
		ORDER BY start_date
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var festivals []domain.FestivalEvent
	for rows.Next() {
		var f domain.FestivalEvent
		if err := rows.Scan(&f.ID, &f.Name, &f.NameHindi, &f.Description, &f.KarmaMultiplier, &f.BonusKarma, &f.StartDate, &f.EndDate); err != nil {
			return nil, err
		}
		festivals = append(festivals, f)
	}
	return festivals, nil
}

func (r *ChallengeRepository) GetUserStreak(ctx context.Context, userID uuid.UUID) (*domain.UserStreak, error) {
	query := `
		SELECT user_id, current_streak, longest_streak, last_challenge_date::text, total_challenges_completed
		FROM user_streaks WHERE user_id = $1
	`
	streak := &domain.UserStreak{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&streak.UserID, &streak.CurrentStreak, &streak.LongestStreak, &streak.LastChallengeDate, &streak.TotalChallengesCompleted,
	)
	return streak, err
}

func (r *ChallengeRepository) UpdateUserStreak(ctx context.Context, userID uuid.UUID, streak int, lastDate string, total int) error {
	query := `
		INSERT INTO user_streaks (user_id, current_streak, longest_streak, last_challenge_date, total_challenges_completed)
		VALUES ($1, $2, $2, $3::date, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			current_streak = $2,
			longest_streak = GREATEST(user_streaks.longest_streak, $2),
			last_challenge_date = $3::date,
			total_challenges_completed = $4,
			updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, userID, streak, lastDate, total)
	return err
}

// EventRepository implements domain.EventRepository
type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) CreateEvent(ctx context.Context, params domain.CreateEventParams) (*domain.LocalEvent, error) {
	query := `
		INSERT INTO local_events (creator_id, title, title_hindi, description, event_type, venue, city, event_date, end_date, max_attendees, is_public)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, creator_id, title, title_hindi, description, event_type, venue, city, event_date, end_date, max_attendees, is_public, karma_reward, created_at
	`
	event := &domain.LocalEvent{}
	err := r.db.QueryRow(ctx, query,
		params.CreatorID, params.Title, params.TitleHindi, params.Description, params.EventType,
		params.Venue, params.City, params.EventDate, params.EndDate, params.MaxAttendees, params.IsPublic,
	).Scan(
		&event.ID, &event.CreatorID, &event.Title, &event.TitleHindi, &event.Description,
		&event.EventType, &event.Venue, &event.City, &event.EventDate, &event.EndDate,
		&event.MaxAttendees, &event.IsPublic, &event.KarmaReward, &event.CreatedAt,
	)
	return event, err
}

func (r *EventRepository) GetEvent(ctx context.Context, eventID uuid.UUID) (*domain.LocalEvent, error) {
	query := `
		SELECT e.id, e.creator_id, e.title, e.title_hindi, e.description, e.event_type, e.venue, e.city,
		       e.event_date, e.end_date, e.max_attendees, e.is_public, e.karma_reward, e.created_at
		FROM local_events e WHERE e.id = $1
	`
	event := &domain.LocalEvent{}
	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&event.ID, &event.CreatorID, &event.Title, &event.TitleHindi, &event.Description,
		&event.EventType, &event.Venue, &event.City, &event.EventDate, &event.EndDate,
		&event.MaxAttendees, &event.IsPublic, &event.KarmaReward, &event.CreatedAt,
	)
	return event, err
}

func (r *EventRepository) GetEventsByCity(ctx context.Context, city string, eventType *string, limit, offset int) ([]domain.LocalEvent, error) {
	query := `
		SELECT e.id, e.creator_id, e.title, e.title_hindi, e.description, e.event_type, e.venue, e.city,
		       e.event_date, e.max_attendees, e.is_public, e.karma_reward, e.created_at,
		       u.id, u.name, u.avatar_url,
		       (SELECT COUNT(*) FROM event_rsvps WHERE event_id = e.id AND status = 'going') as rsvp_count
		FROM local_events e
		JOIN users u ON e.creator_id = u.id
		WHERE e.event_date >= NOW()
		  AND ($1 = '' OR e.city = $1)
		  AND ($2::text IS NULL OR e.event_type = $2)
		  AND e.is_public = TRUE
		ORDER BY e.event_date ASC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, city, eventType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.LocalEvent
	for rows.Next() {
		var event domain.LocalEvent
		var creator domain.UserResponse
		if err := rows.Scan(
			&event.ID, &event.CreatorID, &event.Title, &event.TitleHindi, &event.Description,
			&event.EventType, &event.Venue, &event.City, &event.EventDate, &event.MaxAttendees,
			&event.IsPublic, &event.KarmaReward, &event.CreatedAt,
			&creator.ID, &creator.Name, &creator.AvatarURL,
			&event.RSVPCount,
		); err != nil {
			return nil, err
		}
		event.Creator = &creator
		events = append(events, event)
	}
	return events, nil
}

func (r *EventRepository) GetUpcomingEvents(ctx context.Context, limit int) ([]domain.LocalEvent, error) {
	return r.GetEventsByCity(ctx, "", nil, limit, 0)
}

func (r *EventRepository) GetUserEvents(ctx context.Context, userID uuid.UUID) (created []domain.LocalEvent, attending []domain.LocalEvent, err error) {
	// Created events
	createdQuery := `
		SELECT id, title, title_hindi, event_type, venue, city, event_date, karma_reward
		FROM local_events WHERE creator_id = $1 AND event_date >= NOW()
		ORDER BY event_date
	`
	rows, err := r.db.Query(ctx, createdQuery, userID)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var e domain.LocalEvent
		rows.Scan(&e.ID, &e.Title, &e.TitleHindi, &e.EventType, &e.Venue, &e.City, &e.EventDate, &e.KarmaReward)
		created = append(created, e)
	}
	rows.Close()

	// Attending events
	attendingQuery := `
		SELECT e.id, e.title, e.title_hindi, e.event_type, e.venue, e.city, e.event_date, e.karma_reward
		FROM local_events e
		JOIN event_rsvps r ON e.id = r.event_id
		WHERE r.user_id = $1 AND r.status = 'going' AND e.event_date >= NOW()
		ORDER BY e.event_date
	`
	rows, err = r.db.Query(ctx, attendingQuery, userID)
	if err != nil {
		return created, nil, err
	}
	for rows.Next() {
		var e domain.LocalEvent
		rows.Scan(&e.ID, &e.Title, &e.TitleHindi, &e.EventType, &e.Venue, &e.City, &e.EventDate, &e.KarmaReward)
		attending = append(attending, e)
	}
	rows.Close()

	return created, attending, nil
}

func (r *EventRepository) UpdateEvent(ctx context.Context, eventID uuid.UUID, params domain.CreateEventParams) error {
	query := `
		UPDATE local_events SET title = $1, description = $2, venue = $3, event_date = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, params.Title, params.Description, params.Venue, params.EventDate, eventID)
	return err
}

func (r *EventRepository) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM local_events WHERE id = $1`, eventID)
	return err
}

func (r *EventRepository) RSVPEvent(ctx context.Context, eventID, userID uuid.UUID, status string) error {
	query := `
		INSERT INTO event_rsvps (event_id, user_id, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (event_id, user_id) DO UPDATE SET status = $3
	`
	_, err := r.db.Exec(ctx, query, eventID, userID, status)
	return err
}

func (r *EventRepository) GetEventRSVPs(ctx context.Context, eventID uuid.UUID) ([]domain.EventRSVP, error) {
	query := `SELECT id, event_id, user_id, status, created_at FROM event_rsvps WHERE event_id = $1`
	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rsvps []domain.EventRSVP
	for rows.Next() {
		var rsvp domain.EventRSVP
		if err := rows.Scan(&rsvp.ID, &rsvp.EventID, &rsvp.UserID, &rsvp.Status, &rsvp.CreatedAt); err != nil {
			return nil, err
		}
		rsvps = append(rsvps, rsvp)
	}
	return rsvps, nil
}

func (r *EventRepository) GetUserRSVP(ctx context.Context, eventID, userID uuid.UUID) (*string, error) {
	query := `SELECT status FROM event_rsvps WHERE event_id = $1 AND user_id = $2`
	var status string
	err := r.db.QueryRow(ctx, query, eventID, userID).Scan(&status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *EventRepository) GetRSVPCount(ctx context.Context, eventID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM event_rsvps WHERE event_id = $1 AND status = 'going'`
	var count int
	err := r.db.QueryRow(ctx, query, eventID).Scan(&count)
	return count, err
}

// IcebreakerRepository implements domain.IcebreakerRepository
type IcebreakerRepository struct {
	db *pgxpool.Pool
}

func NewIcebreakerRepository(db *pgxpool.Pool) *IcebreakerRepository {
	return &IcebreakerRepository{db: db}
}

func (r *IcebreakerRepository) GetTodaysMatch(ctx context.Context, userID uuid.UUID, date string) (*domain.DailyMatch, error) {
	query := `
		SELECT dm.id, dm.user_id, dm.matched_user_id, dm.match_date::text, dm.match_reason, dm.is_accepted, dm.accepted_at,
		       u.id, u.name, u.avatar_url
		FROM daily_matches dm
		JOIN users u ON dm.matched_user_id = u.id
		WHERE dm.user_id = $1 AND dm.match_date = $2::date
	`
	match := &domain.DailyMatch{}
	var user domain.UserResponse
	err := r.db.QueryRow(ctx, query, userID, date).Scan(
		&match.ID, &match.UserID, &match.MatchedUserID, &match.MatchDate, &match.MatchReason, &match.IsAccepted, &match.AcceptedAt,
		&user.ID, &user.Name, &user.AvatarURL,
	)
	if err != nil {
		return nil, err
	}
	match.MatchedUser = &user
	return match, nil
}

func (r *IcebreakerRepository) CreateMatch(ctx context.Context, userID, matchedUserID uuid.UUID, reason string) (*domain.DailyMatch, error) {
	query := `
		INSERT INTO daily_matches (user_id, matched_user_id, match_reason)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, matched_user_id, match_date::text, match_reason
	`
	match := &domain.DailyMatch{}
	err := r.db.QueryRow(ctx, query, userID, matchedUserID, reason).Scan(
		&match.ID, &match.UserID, &match.MatchedUserID, &match.MatchDate, &match.MatchReason,
	)
	if err != nil {
		return nil, err
	}

	// Get matched user info
	var user domain.UserResponse
	r.db.QueryRow(ctx, `SELECT id, name, avatar_url FROM users WHERE id = $1`, matchedUserID).Scan(&user.ID, &user.Name, &user.AvatarURL)
	match.MatchedUser = &user

	return match, nil
}

func (r *IcebreakerRepository) RespondToMatch(ctx context.Context, matchID uuid.UUID, accept bool) error {
	query := `UPDATE daily_matches SET is_accepted = $1, accepted_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, accept, matchID)
	return err
}

func (r *IcebreakerRepository) MarkKarmaAwarded(ctx context.Context, matchID uuid.UUID) error {
	query := `UPDATE daily_matches SET karma_awarded = TRUE WHERE id = $1`
	_, err := r.db.Exec(ctx, query, matchID)
	return err
}

func (r *IcebreakerRepository) FindPotentialMatches(ctx context.Context, userID uuid.UUID, city *string, limit int) ([]uuid.UUID, error) {
	query := `
		SELECT u.id FROM users u
		LEFT JOIN profiles p ON u.id = p.user_id
		WHERE u.id != $1
		  AND u.is_active = TRUE
		  AND u.id NOT IN (SELECT matched_user_id FROM daily_matches WHERE user_id = $1 AND match_date >= CURRENT_DATE - INTERVAL '7 days')
		  AND u.id NOT IN (SELECT blocked_user_id FROM blocked_users WHERE user_id = $1)
		  AND ($2::text IS NULL OR p.city = $2)
		ORDER BY RANDOM()
		LIMIT $3
	`
	rows, err := r.db.Query(ctx, query, userID, city, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *IcebreakerRepository) GetUserInterests(ctx context.Context, userID uuid.UUID) ([]domain.UserInterest, error) {
	query := `SELECT id, user_id, interest, category, created_at FROM user_interests WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interests []domain.UserInterest
	for rows.Next() {
		var i domain.UserInterest
		if err := rows.Scan(&i.ID, &i.UserID, &i.Interest, &i.Category, &i.CreatedAt); err != nil {
			return nil, err
		}
		interests = append(interests, i)
	}
	return interests, nil
}

func (r *IcebreakerRepository) AddInterest(ctx context.Context, userID uuid.UUID, interest, category string) error {
	query := `INSERT INTO user_interests (user_id, interest, category) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, query, userID, interest, category)
	return err
}

func (r *IcebreakerRepository) RemoveInterest(ctx context.Context, userID uuid.UUID, interest string) error {
	query := `DELETE FROM user_interests WHERE user_id = $1 AND interest = $2`
	_, err := r.db.Exec(ctx, query, userID, interest)
	return err
}

func (r *IcebreakerRepository) GetPopularInterests(ctx context.Context, limit int) ([]struct {
	Interest string
	Category string
	Count    int
}, error) {
	query := `
		SELECT interest, category, COUNT(*) as count
		FROM user_interests
		GROUP BY interest, category
		ORDER BY count DESC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Interest string
		Category string
		Count    int
	}
	for rows.Next() {
		var item struct {
			Interest string
			Category string
			Count    int
		}
		if err := rows.Scan(&item.Interest, &item.Category, &item.Count); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

// StoryCleanup cleans up expired stories
func (r *PostgresRepository) CleanupExpiredStories(ctx context.Context) (int64, error) {
	query := `DELETE FROM stories WHERE expires_at < NOW()`
	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// GetStoryByID gets a single story
func (r *PostgresRepository) GetStoryByID(ctx context.Context, storyID uuid.UUID) (*domain.Story, error) {
	query := `
		SELECT s.id, s.user_id, s.media_url, s.media_type, s.caption, s.location_lat, s.location_lng,
		       s.expires_at, s.created_at, u.id, u.name, u.avatar_url
		FROM stories s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = $1
	`
	var story domain.Story
	var user domain.UserResponse
	err := r.db.QueryRow(ctx, query, storyID).Scan(
		&story.ID, &story.UserID, &story.MediaURL, &story.MediaType, &story.Caption,
		&story.LocationLat, &story.LocationLng, &story.ExpiresAt, &story.CreatedAt,
		&user.ID, &user.Name, &user.AvatarURL,
	)
	if err != nil {
		return nil, err
	}
	story.User = &user
	return &story, nil
}

// DeleteStory deletes a story
func (r *PostgresRepository) DeleteStory(ctx context.Context, storyID, userID uuid.UUID) error {
	query := `DELETE FROM stories WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(ctx, query, storyID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("story not found or not owned by user")
	}
	return nil
}
