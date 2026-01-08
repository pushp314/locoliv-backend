package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/locolive/backend/internal/domain"
)

// DiscoveryRepository implements domain.DiscoveryRepository with PostGIS
type DiscoveryRepository struct {
	db *pgxpool.Pool
}

func NewDiscoveryRepository(db *pgxpool.Pool) *DiscoveryRepository {
	return &DiscoveryRepository{db: db}
}

func (r *DiscoveryRepository) GetNearbyUsers(ctx context.Context, excludeUserID uuid.UUID, lat, lng, radiusKm float64, limit int) ([]domain.NearbyUser, error) {
	// Using Haversine formula for distance calculation
	query := `
		SELECT 
			u.id, u.name, p.avatar_url, p.city,
			u.is_verified, u.verification_type,
			COALESCE(uk.total_karma, 0) as total_karma,
			COALESCE(uk.current_tier, 'navasadhak') as current_tier,
			uk.last_calculated_at as last_active_at,
			(6371 * acos(
				cos(radians($1)) * cos(radians(p.location_lat)) *
				cos(radians(p.location_lng) - radians($2)) +
				sin(radians($1)) * sin(radians(p.location_lat))
			)) AS distance_km
		FROM users u
		JOIN profiles p ON u.id = p.user_id
		LEFT JOIN user_karma uk ON u.id = uk.user_id
		WHERE u.id != $3
			AND p.location_lat IS NOT NULL
			AND p.location_lng IS NOT NULL
			AND (6371 * acos(
				cos(radians($1)) * cos(radians(p.location_lat)) *
				cos(radians(p.location_lng) - radians($2)) +
				sin(radians($1)) * sin(radians(p.location_lat))
			)) <= $4
		ORDER BY distance_km ASC
		LIMIT $5
	`

	rows, err := r.db.Query(ctx, query, lat, lng, excludeUserID, radiusKm, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.NearbyUser
	for rows.Next() {
		var u domain.NearbyUser
		if err := rows.Scan(
			&u.ID, &u.Name, &u.AvatarURL, &u.City,
			&u.IsVerified, &u.VerificationType,
			&u.TotalKarma, &u.CurrentTier, &u.LastActiveAt,
			&u.DistanceKm,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *DiscoveryRepository) GetNearbyStories(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]domain.NearbyStory, error) {
	query := `
		SELECT 
			s.id, s.user_id, u.name, p.avatar_url,
			s.media_url, s.media_type, s.caption, s.city,
			s.created_at, COALESCE(s.view_count, 0),
			(6371 * acos(
				cos(radians($1)) * cos(radians(s.location_lat)) *
				cos(radians(s.location_lng) - radians($2)) +
				sin(radians($1)) * sin(radians(s.location_lat))
			)) AS distance_km
		FROM stories s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN profiles p ON u.id = p.user_id
		WHERE s.expires_at > NOW()
			AND s.location_lat IS NOT NULL
			AND s.location_lng IS NOT NULL
			AND (6371 * acos(
				cos(radians($1)) * cos(radians(s.location_lat)) *
				cos(radians(s.location_lng) - radians($2)) +
				sin(radians($1)) * sin(radians(s.location_lat))
			)) <= $3
		ORDER BY s.created_at DESC
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query, lat, lng, radiusKm, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stories []domain.NearbyStory
	for rows.Next() {
		var s domain.NearbyStory
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.UserName, &s.UserAvatar,
			&s.MediaURL, &s.MediaType, &s.Caption, &s.City,
			&s.CreatedAt, &s.ViewCount, &s.DistanceKm,
		); err != nil {
			return nil, err
		}
		stories = append(stories, s)
	}
	return stories, nil
}

func (r *DiscoveryRepository) GetNearbyEvents(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]domain.NearbyEvent, error) {
	query := `
		SELECT 
			e.id, e.title, e.description, e.event_date,
			e.location, e.city,
			(SELECT COUNT(*) FROM event_rsvps WHERE event_id = e.id) as rsvp_count,
			u.name as creator_name,
			(6371 * acos(
				cos(radians($1)) * cos(radians(e.location_lat)) *
				cos(radians(e.location_lng) - radians($2)) +
				sin(radians($1)) * sin(radians(e.location_lat))
			)) AS distance_km
		FROM local_events e
		JOIN users u ON e.creator_id = u.id
		WHERE e.event_date > NOW()
			AND e.location_lat IS NOT NULL
			AND e.location_lng IS NOT NULL
			AND (6371 * acos(
				cos(radians($1)) * cos(radians(e.location_lat)) *
				cos(radians(e.location_lng) - radians($2)) +
				sin(radians($1)) * sin(radians(e.location_lat))
			)) <= $3
		ORDER BY e.event_date ASC
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query, lat, lng, radiusKm, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.NearbyEvent
	for rows.Next() {
		var e domain.NearbyEvent
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.EventDate,
			&e.Location, &e.City, &e.RSVPCount, &e.CreatorName,
			&e.DistanceKm,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *DiscoveryRepository) UpdateUserLocation(ctx context.Context, userID uuid.UUID, lat, lng float64, city string) error {
	query := `
		UPDATE profiles 
		SET location_lat = $1, location_lng = $2, city = COALESCE(NULLIF($3, ''), city), updated_at = NOW()
		WHERE user_id = $4
	`
	_, err := r.db.Exec(ctx, query, lat, lng, city, userID)
	return err
}

func (r *DiscoveryRepository) GetLocalStats(ctx context.Context, lat, lng, radiusKm float64) (*domain.LocalStats, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM profiles p WHERE 
				p.location_lat IS NOT NULL AND
				(6371 * acos(
					cos(radians($1)) * cos(radians(p.location_lat)) *
					cos(radians(p.location_lng) - radians($2)) +
					sin(radians($1)) * sin(radians(p.location_lat))
				)) <= $3
			) as total_users,
			(SELECT COUNT(*) FROM stories s WHERE 
				s.created_at > NOW() - INTERVAL '24 hours' AND
				s.location_lat IS NOT NULL AND
				(6371 * acos(
					cos(radians($1)) * cos(radians(s.location_lat)) *
					cos(radians(s.location_lng) - radians($2)) +
					sin(radians($1)) * sin(radians(s.location_lat))
				)) <= $3
			) as new_stories,
			(SELECT COUNT(*) FROM local_events e WHERE 
				e.event_date > NOW() AND e.event_date < NOW() + INTERVAL '7 days' AND
				e.location_lat IS NOT NULL AND
				(6371 * acos(
					cos(radians($1)) * cos(radians(e.location_lat)) *
					cos(radians(e.location_lng) - radians($2)) +
					sin(radians($1)) * sin(radians(e.location_lat))
				)) <= $3
			) as upcoming_events
	`

	stats := &domain.LocalStats{}
	err := r.db.QueryRow(ctx, query, lat, lng, radiusKm).Scan(
		&stats.TotalUsers, &stats.NewStories, &stats.UpcomingEvents,
	)
	if err != nil {
		return nil, err
	}

	// Active today is approximated
	stats.ActiveToday = stats.TotalUsers / 10
	if stats.ActiveToday < 1 {
		stats.ActiveToday = 1
	}

	return stats, nil
}

func (r *DiscoveryRepository) GetUserProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	query := `
		SELECT 
			u.id, u.name, p.bio, p.avatar_url, p.city,
			u.is_verified, u.verification_type,
			COALESCE(uk.total_karma, 0), COALESCE(uk.current_tier, 'navasadhak'),
			u.created_at,
			(SELECT COUNT(*) FROM stories WHERE user_id = u.id AND expires_at > NOW()) as story_count,
			(SELECT COUNT(*) FROM connections WHERE (requester_id = u.id OR recipient_id = u.id) AND status = 'accepted') as connection_count,
			(SELECT COUNT(*) FROM user_badges WHERE user_id = u.id) as badge_count
		FROM users u
		LEFT JOIN profiles p ON u.id = p.user_id
		LEFT JOIN user_karma uk ON u.id = uk.user_id
		WHERE u.id = $1
	`

	profile := &domain.UserProfile{}
	var tierStr string
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&profile.ID, &profile.Name, &profile.Bio, &profile.AvatarURL, &profile.City,
		&profile.IsVerified, &profile.VerificationType,
		&profile.TotalKarma, &tierStr,
		&profile.JoinedAt,
		&profile.StoryCount, &profile.ConnectionCount, &profile.BadgeCount,
	)
	if err != nil {
		return nil, err
	}

	profile.CurrentTier = tierStr
	profile.TierName = domain.TierDisplayName(domain.KarmaTier(tierStr))

	return profile, nil
}

func (r *DiscoveryRepository) GetConnectionStatus(ctx context.Context, userID, targetID uuid.UUID) (isConnected, isPending bool, err error) {
	query := `
		SELECT status FROM connections 
		WHERE (requester_id = $1 AND recipient_id = $2) OR (requester_id = $2 AND recipient_id = $1)
		LIMIT 1
	`
	var status string
	err = r.db.QueryRow(ctx, query, userID, targetID).Scan(&status)
	if err != nil {
		return false, false, nil // No connection
	}
	return status == "accepted", status == "pending", nil
}
