package mysql

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"moonick/internal/model/entity"
)

var ErrTripNotFound = errors.New("trip not found")

const tripBaseSelect = `SELECT id, publisher_user_id, trip_type, from_text, to_text,
	TIMESTAMP(departure_date, departure_time) AS departure_at,
	seat_count, price_amount, is_price_negotiable, contact_wechat, contact_phone,
	remark, status, closed_reason, created_at, updated_at
FROM trips`

type TripRepository struct {
	db        *sql.DB
	mu        sync.RWMutex
	nextID    int64
	tripsByID map[int64]entity.Trip
}

func NewTripRepository(dbs ...*sql.DB) *TripRepository {
	if len(dbs) > 0 && dbs[0] != nil {
		return &TripRepository{db: dbs[0]}
	}
	if len(dbs) == 0 {
		if db := GetDB(); db != nil {
			return &TripRepository{db: db}
		}
	}

	return &TripRepository{
		nextID:    2000,
		tripsByID: make(map[int64]entity.Trip),
	}
}

func (r *TripRepository) Create(ctx context.Context, trip entity.Trip) (*entity.Trip, error) {
	if r.db != nil {
		return r.createInDB(ctx, trip)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if trip.ID == 0 {
		r.nextID++
		trip.ID = r.nextID
	} else if trip.ID > r.nextID {
		r.nextID = trip.ID
	}
	if trip.CreatedAt.IsZero() {
		trip.CreatedAt = now
	}
	trip.UpdatedAt = now
	r.tripsByID[trip.ID] = trip
	return cloneTrip(trip), nil
}

func (r *TripRepository) Update(ctx context.Context, trip entity.Trip) (*entity.Trip, error) {
	if r.db != nil {
		return r.updateInDB(ctx, trip)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.tripsByID[trip.ID]
	if !ok {
		return nil, ErrTripNotFound
	}

	trip.CreatedAt = current.CreatedAt
	trip.UpdatedAt = time.Now()
	r.tripsByID[trip.ID] = trip
	return cloneTrip(trip), nil
}

func (r *TripRepository) FindByID(ctx context.Context, id int64) (*entity.Trip, error) {
	if r.db != nil {
		return r.findOne(ctx, tripBaseSelect+` WHERE id = ? AND deleted_at IS NULL`, id)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	trip, ok := r.tripsByID[id]
	if !ok {
		return nil, nil
	}
	return cloneTrip(trip), nil
}

func (r *TripRepository) List(ctx context.Context, filter entity.TripFilter) ([]*entity.Trip, int, error) {
	if r.db != nil {
		return r.listFromDB(ctx, filter)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]entity.Trip, 0, len(r.tripsByID))
	for _, trip := range r.tripsByID {
		if !matchTripFilter(trip, filter) {
			continue
		}
		items = append(items, trip)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	total := len(items)
	start := filter.Offset
	if start > total {
		start = total
	}
	end := total
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}

	result := make([]*entity.Trip, 0, end-start)
	for _, trip := range items[start:end] {
		result = append(result, cloneTrip(trip))
	}
	return result, total, nil
}

func (r *TripRepository) ExpireTripsBefore(ctx context.Context, before time.Time) (int64, error) {
	if r.db != nil {
		beforeDate := before.Format(time.DateOnly)
		beforeTime := before.Format("15:04:05")
		result, err := r.db.ExecContext(
			ctx,
			`UPDATE trips
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE deleted_at IS NULL
  AND status IN (?, ?)
  AND (
	departure_date < ?
	OR (departure_date = ? AND departure_time < ?)
  )`,
			entity.TripStatusExpired,
			entity.TripStatusActive,
			entity.TripStatusFull,
			beforeDate,
			beforeDate,
			beforeTime,
		)
		if err != nil {
			return 0, err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}

		return affected, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var affected int64
	for id, trip := range r.tripsByID {
		if trip.Status != entity.TripStatusActive && trip.Status != entity.TripStatusFull {
			continue
		}
		if trip.DepartureAt.Before(before) {
			trip.Status = entity.TripStatusExpired
			trip.UpdatedAt = time.Now()
			r.tripsByID[id] = trip
			affected++
		}
	}
	return affected, nil
}

func (r *TripRepository) createInDB(ctx context.Context, trip entity.Trip) (*entity.Trip, error) {
	now := time.Now()
	if trip.CreatedAt.IsZero() {
		trip.CreatedAt = now
	}
	if trip.UpdatedAt.IsZero() {
		trip.UpdatedAt = trip.CreatedAt
	}

	result, err := r.db.ExecContext(ctx, `INSERT INTO trips (
	publisher_user_id, trip_type, from_text, to_text, departure_date, departure_time,
	seat_count, price_amount, is_price_negotiable, contact_wechat, contact_phone,
	remark, status, closed_reason, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		trip.UserID,
		trip.TripType,
		trip.FromText,
		trip.ToText,
		trip.DepartureAt.Format(time.DateOnly),
		trip.DepartureAt.Format("15:04:05"),
		trip.SeatCount,
		trip.PriceAmount,
		trip.IsPriceNegotiable,
		trip.ContactWechat,
		trip.ContactPhone,
		trip.Remark,
		trip.Status,
		trip.ClosedReason,
		trip.CreatedAt,
		trip.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *TripRepository) updateInDB(ctx context.Context, trip entity.Trip) (*entity.Trip, error) {
	current, err := r.FindByID(ctx, trip.ID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrTripNotFound
	}

	trip.CreatedAt = current.CreatedAt
	trip.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, `UPDATE trips
SET publisher_user_id = ?, trip_type = ?, from_text = ?, to_text = ?,
	departure_date = ?, departure_time = ?, seat_count = ?, price_amount = ?,
	is_price_negotiable = ?, contact_wechat = ?, contact_phone = ?, remark = ?,
	status = ?, closed_reason = ?, updated_at = ?
WHERE id = ? AND deleted_at IS NULL`,
		trip.UserID,
		trip.TripType,
		trip.FromText,
		trip.ToText,
		trip.DepartureAt.Format(time.DateOnly),
		trip.DepartureAt.Format("15:04:05"),
		trip.SeatCount,
		trip.PriceAmount,
		trip.IsPriceNegotiable,
		trip.ContactWechat,
		trip.ContactPhone,
		trip.Remark,
		trip.Status,
		trip.ClosedReason,
		trip.UpdatedAt,
		trip.ID,
	)
	if err != nil {
		return nil, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrTripNotFound
	}

	return r.FindByID(ctx, trip.ID)
}

func (r *TripRepository) findOne(ctx context.Context, query string, arg any) (*entity.Trip, error) {
	var trip entity.Trip

	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&trip.ID,
		&trip.UserID,
		&trip.TripType,
		&trip.FromText,
		&trip.ToText,
		&trip.DepartureAt,
		&trip.SeatCount,
		&trip.PriceAmount,
		&trip.IsPriceNegotiable,
		&trip.ContactWechat,
		&trip.ContactPhone,
		&trip.Remark,
		&trip.Status,
		&trip.ClosedReason,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *TripRepository) listFromDB(ctx context.Context, filter entity.TripFilter) ([]*entity.Trip, int, error) {
	whereClauses := []string{"deleted_at IS NULL"}
	args := make([]any, 0, 16)
	countArgs := make([]any, 0, 16)
	appendFilter := func(clause string, values ...any) {
		whereClauses = append(whereClauses, clause)
		args = append(args, values...)
		countArgs = append(countArgs, values...)
	}

	if filter.UserID != nil {
		appendFilter("publisher_user_id = ?", *filter.UserID)
	}
	if filter.TripType != "" {
		appendFilter("trip_type = ?", filter.TripType)
	}
	if len(filter.Statuses) > 0 {
		appendFilter("status IN ("+placeholders(len(filter.Statuses))+")", stringsToAny(filter.Statuses)...)
	}
	if len(filter.IDs) > 0 {
		appendFilter("id IN ("+placeholders(len(filter.IDs))+")", int64sToAny(filter.IDs)...)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		appendFilter("(from_text LIKE ? OR to_text LIKE ?)", like, like)
	}

	whereSQL := ` WHERE ` + strings.Join(whereClauses, ` AND `)
	countQuery := `SELECT COUNT(*) FROM trips` + whereSQL
	listQuery := tripBaseSelect + whereSQL + ` ORDER BY created_at DESC, id DESC`
	if filter.Limit > 0 {
		listQuery += ` LIMIT ? OFFSET ?`
		args = append(args, filter.Limit, filter.Offset)
	} else if filter.Offset > 0 {
		listQuery += ` LIMIT 18446744073709551615 OFFSET ?`
		args = append(args, filter.Offset)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*entity.Trip, 0)
	for rows.Next() {
		var trip entity.Trip
		if err := rows.Scan(
			&trip.ID,
			&trip.UserID,
			&trip.TripType,
			&trip.FromText,
			&trip.ToText,
			&trip.DepartureAt,
			&trip.SeatCount,
			&trip.PriceAmount,
			&trip.IsPriceNegotiable,
			&trip.ContactWechat,
			&trip.ContactPhone,
			&trip.Remark,
			&trip.Status,
			&trip.ClosedReason,
			&trip.CreatedAt,
			&trip.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, &trip)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func matchTripFilter(trip entity.Trip, filter entity.TripFilter) bool {
	if filter.UserID != nil && trip.UserID != *filter.UserID {
		return false
	}
	if filter.TripType != "" && trip.TripType != filter.TripType {
		return false
	}
	if len(filter.Statuses) > 0 && !containsString(filter.Statuses, trip.Status) {
		return false
	}
	if len(filter.IDs) > 0 && !containsInt64(filter.IDs, trip.ID) {
		return false
	}
	if filter.Keyword != "" {
		keyword := strings.ToLower(strings.TrimSpace(filter.Keyword))
		haystack := strings.ToLower(trip.FromText + " " + trip.ToText)
		if !strings.Contains(haystack, keyword) {
			return false
		}
	}
	return true
}

func cloneTrip(trip entity.Trip) *entity.Trip {
	copied := trip
	return &copied
}

func placeholders(size int) string {
	parts := make([]string, size)
	for i := range parts {
		parts[i] = "?"
	}
	return strings.Join(parts, ", ")
}

func stringsToAny(values []string) []any {
	result := make([]any, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func int64sToAny(values []int64) []any {
	result := make([]any, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func containsInt64(items []int64, target int64) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
