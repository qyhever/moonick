package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"moonick/internal/model/entity"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	repositoryTestDriverOnce sync.Once
	repositoryTestDBSeq      atomic.Int64
	repositoryTestDBMu       sync.Mutex
	repositoryTestDBStates   = make(map[string]*repositoryTestState)
)

func newRepositoryTestDB(t *testing.T) *sql.DB {
	t.Helper()

	repositoryTestDriverOnce.Do(func() {
		sql.Register("moonick-repository-test", repositoryTestDriver{})
	})

	dsn := fmt.Sprintf("repo-test-%d", repositoryTestDBSeq.Add(1))
	state := &repositoryTestState{
		nextUserID:     1000,
		nextTripID:     0,
		usersByID:      make(map[int64]entity.User),
		userIDsByEmail: make(map[string]int64),
		tripsByID:      make(map[int64]entity.Trip),
		favoritesByKey: make(map[string]entity.Favorite),
		adminsByID:     make(map[int64]entity.Admin),
		adminIDsByName: make(map[string]int64),
	}

	repositoryTestDBMu.Lock()
	repositoryTestDBStates[dsn] = state
	repositoryTestDBMu.Unlock()

	db, err := sql.Open("moonick-repository-test", dsn)
	if err != nil {
		t.Fatalf("open repository test db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
		repositoryTestDBMu.Lock()
		delete(repositoryTestDBStates, dsn)
		repositoryTestDBMu.Unlock()
	})

	return db
}

type repositoryTestDriver struct{}

func (repositoryTestDriver) Open(name string) (driver.Conn, error) {
	repositoryTestDBMu.Lock()
	defer repositoryTestDBMu.Unlock()

	state := repositoryTestDBStates[name]
	if state == nil {
		return nil, fmt.Errorf("repository test db state not found: %s", name)
	}

	return &repositoryTestConn{state: state}, nil
}

type repositoryTestConn struct {
	state *repositoryTestState
}

func (c *repositoryTestConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *repositoryTestConn) Close() error                        { return nil }
func (c *repositoryTestConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }

func (c *repositoryTestConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return c.state.exec(query, args)
}

func (c *repositoryTestConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.state.query(query, args)
}

func (c *repositoryTestConn) CheckNamedValue(*driver.NamedValue) error { return nil }

type repositoryTestState struct {
	mu             sync.Mutex
	nextUserID     int64
	nextTripID     int64
	usersByID      map[int64]entity.User
	userIDsByEmail map[string]int64
	tripsByID      map[int64]entity.Trip
	favoritesByKey map[string]entity.Favorite
	adminsByID     map[int64]entity.Admin
	adminIDsByName map[string]int64
}

func (s *repositoryTestState) exec(query string, args []driver.NamedValue) (driver.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sqlText := normalizeRepositorySQL(query)
	switch {
	case strings.Contains(sqlText, "insert into users"):
		return s.execInsertUser(args)
	case strings.Contains(sqlText, "update users set nickname"):
		return s.execUpdateUserNickname(args)
	case strings.Contains(sqlText, "update users set default_wechat"):
		return s.execUpdateUserContact(args)
	case strings.Contains(sqlText, "update users set avatar_url"):
		return s.execUpdateUserAvatar(args)
	case strings.Contains(sqlText, "insert into admins") && !strings.Contains(sqlText, "on duplicate key update"):
		return s.execInsertAdmin(args)
	case strings.Contains(sqlText, "insert into admins"):
		return s.execUpsertAdmin(args)
	case strings.Contains(sqlText, "insert into trips"):
		return s.execInsertTrip(args)
	case strings.Contains(sqlText, "update trips set publisher_user_id"):
		return s.execUpdateTrip(args)
	case strings.Contains(sqlText, "update trips set status = ?"):
		return s.execExpireTrips(args)
	case strings.Contains(sqlText, "insert into trip_favorites"):
		return s.execInsertFavorite(args)
	case strings.Contains(sqlText, "delete from trip_favorites"):
		return s.execDeleteFavorite(args)
	default:
		return nil, fmt.Errorf("unsupported exec query: %s", query)
	}
}

func (s *repositoryTestState) query(query string, args []driver.NamedValue) (driver.Rows, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sqlText := normalizeRepositorySQL(query)
	switch {
	case strings.Contains(sqlText, "from users where email = ?"):
		return s.queryUserByEmail(args)
	case strings.Contains(sqlText, "select 1 from users where id = ?"):
		return s.queryUserExistsByID(args)
	case strings.Contains(sqlText, "from users where id = ?"):
		return s.queryUserByID(args)
	case strings.Contains(sqlText, "select count(*) from users where"):
		return s.queryUserCount(args, true)
	case strings.Contains(sqlText, "select count(*) from users"):
		return s.queryUserCount(args, false)
	case strings.Contains(sqlText, "from users") && strings.Contains(sqlText, "order by id desc"):
		return s.queryUserList(args, strings.Contains(sqlText, "where"))
	case strings.Contains(sqlText, "from admins where username = ?"):
		return s.queryAdminByUsername(args)
	case strings.Contains(sqlText, "from admins where id = ?"):
		return s.queryAdminByID(args)
	case strings.Contains(sqlText, "from trips where id = ? and deleted_at is null"):
		return s.queryTripByID(args)
	case strings.Contains(sqlText, "select count(*) from trips where"):
		return s.queryTripCount(sqlText, args)
	case strings.Contains(sqlText, "from trips where deleted_at is null") && strings.Contains(sqlText, "order by created_at desc, id desc"):
		return s.queryTripList(sqlText, args)
	case strings.Contains(sqlText, "select 1 from trip_favorites where user_id = ? and trip_id = ?"):
		return s.queryFavoriteExists(args)
	case strings.Contains(sqlText, "select count(*) from trip_favorites where user_id = ?"):
		return s.queryFavoriteCountByUser(args)
	case strings.Contains(sqlText, "select count(*) from trip_favorites"):
		return s.queryFavoriteCount()
	case strings.Contains(sqlText, "from trip_favorites where user_id = ?") && strings.Contains(sqlText, "order by created_at desc, trip_id desc"):
		return s.queryFavoriteList(args)
	default:
		return nil, fmt.Errorf("unsupported query: %s", query)
	}
}

func (s *repositoryTestState) execInsertUser(args []driver.NamedValue) (driver.Result, error) {
	email := stringArg(args[0])
	if _, exists := s.userIDsByEmail[email]; exists {
		return nil, &mysqlDriver.MySQLError{Number: 1062, Message: "duplicate email"}
	}

	s.nextUserID++
	now := time.Now()
	user := entity.User{
		ID:            s.nextUserID,
		Email:         email,
		Phone:         stringArg(args[1]),
		PasswordHash:  stringArg(args[2]),
		Nickname:      stringArg(args[3]),
		AvatarURL:     stringArg(args[4]),
		Status:        stringArg(args[5]),
		DefaultPhone:  stringArg(args[6]),
		DefaultWechat: stringArg(args[7]),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.usersByID[user.ID] = user
	s.userIDsByEmail[user.Email] = user.ID
	return repositoryTestResult{lastInsertID: user.ID, rowsAffected: 1}, nil
}

func (s *repositoryTestState) execUpdateUserNickname(args []driver.NamedValue) (driver.Result, error) {
	id := int64Arg(args[1])
	user, ok := s.usersByID[id]
	if !ok {
		return repositoryTestResult{rowsAffected: 0}, nil
	}
	if user.Nickname == stringArg(args[0]) {
		return repositoryTestResult{rowsAffected: 0}, nil
	}
	user.Nickname = stringArg(args[0])
	user.UpdatedAt = time.Now()
	s.usersByID[id] = user
	return repositoryTestResult{rowsAffected: 1}, nil
}

func (s *repositoryTestState) execUpdateUserContact(args []driver.NamedValue) (driver.Result, error) {
	id := int64Arg(args[2])
	user, ok := s.usersByID[id]
	if !ok {
		return repositoryTestResult{rowsAffected: 0}, nil
	}
	if user.DefaultWechat == stringArg(args[0]) && user.DefaultPhone == stringArg(args[1]) {
		return repositoryTestResult{rowsAffected: 0}, nil
	}
	user.DefaultWechat = stringArg(args[0])
	user.DefaultPhone = stringArg(args[1])
	user.UpdatedAt = time.Now()
	s.usersByID[id] = user
	return repositoryTestResult{rowsAffected: 1}, nil
}

func (s *repositoryTestState) execUpdateUserAvatar(args []driver.NamedValue) (driver.Result, error) {
	id := int64Arg(args[1])
	user, ok := s.usersByID[id]
	if !ok {
		return repositoryTestResult{rowsAffected: 0}, nil
	}
	if user.AvatarURL == stringArg(args[0]) {
		return repositoryTestResult{rowsAffected: 0}, nil
	}
	user.AvatarURL = stringArg(args[0])
	user.UpdatedAt = time.Now()
	s.usersByID[id] = user
	return repositoryTestResult{rowsAffected: 1}, nil
}

func (s *repositoryTestState) execUpsertAdmin(args []driver.NamedValue) (driver.Result, error) {
	id := int64Arg(args[0])
	now := time.Now()
	admin := entity.Admin{
		ID:           id,
		Username:     stringArg(args[1]),
		PasswordHash: stringArg(args[2]),
		Name:         stringArg(args[3]),
		Status:       stringArg(args[4]),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if existing, ok := s.adminsByID[id]; ok {
		admin.CreatedAt = existing.CreatedAt
		if existing.Username != admin.Username {
			delete(s.adminIDsByName, existing.Username)
		}
	}
	s.adminsByID[id] = admin
	s.adminIDsByName[admin.Username] = admin.ID
	return repositoryTestResult{rowsAffected: 1}, nil
}

func (s *repositoryTestState) execInsertAdmin(args []driver.NamedValue) (driver.Result, error) {
	username := stringArg(args[1])
	if _, exists := s.adminIDsByName[username]; exists {
		return nil, &mysqlDriver.MySQLError{Number: 1062, Message: "duplicate admin username"}
	}

	now := time.Now()
	admin := entity.Admin{
		ID:           int64Arg(args[0]),
		Username:     username,
		PasswordHash: stringArg(args[2]),
		Name:         stringArg(args[3]),
		Status:       stringArg(args[4]),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.adminsByID[admin.ID] = admin
	s.adminIDsByName[admin.Username] = admin.ID
	return repositoryTestResult{lastInsertID: admin.ID, rowsAffected: 1}, nil
}

func (s *repositoryTestState) execInsertTrip(args []driver.NamedValue) (driver.Result, error) {
	s.nextTripID++
	id := s.nextTripID
	trip := entity.Trip{
		ID:                id,
		UserID:            int64Arg(args[0]),
		TripType:          stringArg(args[1]),
		FromText:          stringArg(args[2]),
		ToText:            stringArg(args[3]),
		DepartureAt:       mustParseDateTime(stringArg(args[4]), stringArg(args[5])),
		SeatCount:         int(int64Arg(args[6])),
		PriceAmount:       float64Arg(args[7]),
		IsPriceNegotiable: boolArg(args[8]),
		ContactWechat:     stringArg(args[9]),
		ContactPhone:      stringArg(args[10]),
		Remark:            stringArg(args[11]),
		Status:            stringArg(args[12]),
		ClosedReason:      stringArg(args[13]),
		CreatedAt:         timeArg(args[14]),
		UpdatedAt:         timeArg(args[15]),
	}
	s.tripsByID[id] = trip
	return repositoryTestResult{lastInsertID: id, rowsAffected: 1}, nil
}

func (s *repositoryTestState) execUpdateTrip(args []driver.NamedValue) (driver.Result, error) {
	id := int64Arg(args[15])
	current, ok := s.tripsByID[id]
	if !ok {
		return repositoryTestResult{rowsAffected: 0}, nil
	}
	current.UserID = int64Arg(args[0])
	current.TripType = stringArg(args[1])
	current.FromText = stringArg(args[2])
	current.ToText = stringArg(args[3])
	current.DepartureAt = mustParseDateTime(stringArg(args[4]), stringArg(args[5]))
	current.SeatCount = int(int64Arg(args[6]))
	current.PriceAmount = float64Arg(args[7])
	current.IsPriceNegotiable = boolArg(args[8])
	current.ContactWechat = stringArg(args[9])
	current.ContactPhone = stringArg(args[10])
	current.Remark = stringArg(args[11])
	current.Status = stringArg(args[12])
	current.ClosedReason = stringArg(args[13])
	current.UpdatedAt = timeArg(args[14])
	s.tripsByID[id] = current
	return repositoryTestResult{rowsAffected: 1}, nil
}

func (s *repositoryTestState) execExpireTrips(args []driver.NamedValue) (driver.Result, error) {
	beforeDate := stringArg(args[3])
	sameDate := stringArg(args[4])
	beforeTime := stringArg(args[5])
	affected := int64(0)
	for id, trip := range s.tripsByID {
		if trip.Status != stringArg(args[1]) && trip.Status != stringArg(args[2]) {
			continue
		}
		tripDate := trip.DepartureAt.Format(time.DateOnly)
		tripTime := trip.DepartureAt.Format("15:04:05")
		if tripDate < beforeDate || (tripDate == sameDate && tripTime < beforeTime) {
			trip.Status = stringArg(args[0])
			trip.UpdatedAt = time.Now()
			s.tripsByID[id] = trip
			affected++
		}
	}
	return repositoryTestResult{rowsAffected: affected}, nil
}

func (s *repositoryTestState) execInsertFavorite(args []driver.NamedValue) (driver.Result, error) {
	key := favoriteKey(int64Arg(args[0]), int64Arg(args[1]))
	if existing, ok := s.favoritesByKey[key]; ok {
		return repositoryTestResult{rowsAffected: 0, lastInsertID: existing.TripID}, nil
	}
	favorite := entity.Favorite{
		UserID:    int64Arg(args[0]),
		TripID:    int64Arg(args[1]),
		CreatedAt: timeArg(args[2]),
	}
	s.favoritesByKey[key] = favorite
	return repositoryTestResult{rowsAffected: 1}, nil
}

func (s *repositoryTestState) execDeleteFavorite(args []driver.NamedValue) (driver.Result, error) {
	key := favoriteKey(int64Arg(args[0]), int64Arg(args[1]))
	if _, ok := s.favoritesByKey[key]; !ok {
		return repositoryTestResult{rowsAffected: 0}, nil
	}
	delete(s.favoritesByKey, key)
	return repositoryTestResult{rowsAffected: 1}, nil
}

func (s *repositoryTestState) queryUserByEmail(args []driver.NamedValue) (driver.Rows, error) {
	id, ok := s.userIDsByEmail[stringArg(args[0])]
	if !ok {
		return newRepositoryRows(userColumns, nil), nil
	}
	return newRepositoryRows(userColumns, [][]driver.Value{userRow(s.usersByID[id])}), nil
}

func (s *repositoryTestState) queryUserByID(args []driver.NamedValue) (driver.Rows, error) {
	user, ok := s.usersByID[int64Arg(args[0])]
	if !ok {
		return newRepositoryRows(userColumns, nil), nil
	}
	return newRepositoryRows(userColumns, [][]driver.Value{userRow(user)}), nil
}

func (s *repositoryTestState) queryUserExistsByID(args []driver.NamedValue) (driver.Rows, error) {
	_, ok := s.usersByID[int64Arg(args[0])]
	if !ok {
		return newRepositoryRows([]string{"exists"}, nil), nil
	}
	return newRepositoryRows([]string{"exists"}, [][]driver.Value{{int64(1)}}), nil
}

func (s *repositoryTestState) queryUserCount(args []driver.NamedValue, filtered bool) (driver.Rows, error) {
	total := 0
	keyword := ""
	if filtered {
		keyword = likeKeyword(args[0])
	}
	for _, user := range s.usersByID {
		if keyword != "" && !matchUserKeyword(user, keyword) {
			continue
		}
		total++
	}
	return newRepositoryRows([]string{"count"}, [][]driver.Value{{int64(total)}}), nil
}

func (s *repositoryTestState) queryUserList(args []driver.NamedValue, filtered bool) (driver.Rows, error) {
	keyword := ""
	limitArg := 0
	offsetArg := 0
	if filtered {
		keyword = likeKeyword(args[0])
		switch len(args) {
		case 5:
			limitArg = int(int64Arg(args[3]))
			offsetArg = int(int64Arg(args[4]))
		case 4:
			limitArg = -1
			offsetArg = int(int64Arg(args[3]))
		}
	} else {
		switch len(args) {
		case 2:
			limitArg = int(int64Arg(args[0]))
			offsetArg = int(int64Arg(args[1]))
		case 1:
			limitArg = -1
			offsetArg = int(int64Arg(args[0]))
		}
	}

	items := make([]entity.User, 0, len(s.usersByID))
	for _, user := range s.usersByID {
		if keyword != "" && !matchUserKeyword(user, keyword) {
			continue
		}
		items = append(items, user)
	}

	sortUsersDesc(items)
	if offsetArg > len(items) {
		offsetArg = len(items)
	}
	if limitArg < 0 {
		limitArg = 0
	}
	end := len(items)
	if limitArg > 0 && offsetArg+limitArg < end {
		end = offsetArg + limitArg
	}

	values := make([][]driver.Value, 0, end-offsetArg)
	for _, user := range items[offsetArg:end] {
		values = append(values, userRow(user))
	}
	return newRepositoryRows(userColumns, values), nil
}

func (s *repositoryTestState) queryAdminByUsername(args []driver.NamedValue) (driver.Rows, error) {
	id, ok := s.adminIDsByName[stringArg(args[0])]
	if !ok {
		return newRepositoryRows(adminColumns, nil), nil
	}
	return newRepositoryRows(adminColumns, [][]driver.Value{adminRow(s.adminsByID[id])}), nil
}

func (s *repositoryTestState) queryAdminByID(args []driver.NamedValue) (driver.Rows, error) {
	admin, ok := s.adminsByID[int64Arg(args[0])]
	if !ok {
		return newRepositoryRows(adminColumns, nil), nil
	}
	return newRepositoryRows(adminColumns, [][]driver.Value{adminRow(admin)}), nil
}

func (s *repositoryTestState) queryTripByID(args []driver.NamedValue) (driver.Rows, error) {
	trip, ok := s.tripsByID[int64Arg(args[0])]
	if !ok {
		return newRepositoryRows(tripColumns, nil), nil
	}
	return newRepositoryRows(tripColumns, [][]driver.Value{tripRow(trip)}), nil
}

func (s *repositoryTestState) queryTripCount(sqlText string, args []driver.NamedValue) (driver.Rows, error) {
	filter, _, _ := parseTripQueryOptions(sqlText, args)
	total := 0
	for _, trip := range s.tripsByID {
		if matchTripFilter(trip, filter) {
			total++
		}
	}
	return newRepositoryRows([]string{"count"}, [][]driver.Value{{int64(total)}}), nil
}

func (s *repositoryTestState) queryTripList(sqlText string, args []driver.NamedValue) (driver.Rows, error) {
	filter, offset, limit := parseTripQueryOptions(sqlText, args)

	items := make([]entity.Trip, 0, len(s.tripsByID))
	for _, trip := range s.tripsByID {
		if matchTripFilter(trip, filter) {
			items = append(items, trip)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	if offset > len(items) {
		offset = len(items)
	}
	end := len(items)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}

	values := make([][]driver.Value, 0, end-offset)
	for _, trip := range items[offset:end] {
		values = append(values, tripRow(trip))
	}
	return newRepositoryRows(tripColumns, values), nil
}

func (s *repositoryTestState) queryFavoriteExists(args []driver.NamedValue) (driver.Rows, error) {
	_, ok := s.favoritesByKey[favoriteKey(int64Arg(args[0]), int64Arg(args[1]))]
	if !ok {
		return newRepositoryRows([]string{"exists"}, nil), nil
	}
	return newRepositoryRows([]string{"exists"}, [][]driver.Value{{int64(1)}}), nil
}

func (s *repositoryTestState) queryFavoriteCount() (driver.Rows, error) {
	return newRepositoryRows([]string{"count"}, [][]driver.Value{{int64(len(s.favoritesByKey))}}), nil
}

func (s *repositoryTestState) queryFavoriteCountByUser(args []driver.NamedValue) (driver.Rows, error) {
	total := 0
	for _, favorite := range s.favoritesByKey {
		if favorite.UserID == int64Arg(args[0]) {
			total++
		}
	}
	return newRepositoryRows([]string{"count"}, [][]driver.Value{{int64(total)}}), nil
}

func (s *repositoryTestState) queryFavoriteList(args []driver.NamedValue) (driver.Rows, error) {
	userID := int64Arg(args[0])
	offset := 0
	limit := 0
	if len(args) == 3 {
		offset = int(int64Arg(args[2]))
		limit = int(int64Arg(args[1]))
	} else if len(args) == 2 {
		offset = int(int64Arg(args[1]))
	}

	items := make([]entity.Favorite, 0, len(s.favoritesByKey))
	for _, favorite := range s.favoritesByKey {
		if favorite.UserID == userID {
			items = append(items, favorite)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].TripID > items[j].TripID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	if offset > len(items) {
		offset = len(items)
	}
	end := len(items)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}

	values := make([][]driver.Value, 0, end-offset)
	for _, favorite := range items[offset:end] {
		values = append(values, favoriteRow(favorite))
	}
	return newRepositoryRows(favoriteColumns, values), nil
}

type repositoryTestResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r repositoryTestResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r repositoryTestResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type repositoryRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func newRepositoryRows(columns []string, values [][]driver.Value) *repositoryRows {
	return &repositoryRows{columns: columns, values: values}
}

func (r *repositoryRows) Columns() []string { return r.columns }
func (r *repositoryRows) Close() error      { return nil }

func (r *repositoryRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

var userColumns = []string{
	"id", "email", "phone", "password_hash", "nickname", "avatar_url", "status", "default_phone", "default_wechat", "created_at", "updated_at",
}

var adminColumns = []string{
	"id", "username", "password_hash", "display_name", "status", "created_at", "updated_at",
}

var tripColumns = []string{
	"id", "publisher_user_id", "trip_type", "from_text", "to_text", "departure_at",
	"seat_count", "price_amount", "is_price_negotiable", "contact_wechat", "contact_phone",
	"remark", "status", "closed_reason", "created_at", "updated_at",
}

var favoriteColumns = []string{
	"user_id", "trip_id", "created_at",
}

func userRow(user entity.User) []driver.Value {
	return []driver.Value{
		user.ID,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.Nickname,
		user.AvatarURL,
		user.Status,
		user.DefaultPhone,
		user.DefaultWechat,
		user.CreatedAt,
		user.UpdatedAt,
	}
}

func adminRow(admin entity.Admin) []driver.Value {
	return []driver.Value{
		admin.ID,
		admin.Username,
		admin.PasswordHash,
		admin.Name,
		admin.Status,
		admin.CreatedAt,
		admin.UpdatedAt,
	}
}

func tripRow(trip entity.Trip) []driver.Value {
	return []driver.Value{
		trip.ID,
		trip.UserID,
		trip.TripType,
		trip.FromText,
		trip.ToText,
		trip.DepartureAt,
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
	}
}

func favoriteRow(favorite entity.Favorite) []driver.Value {
	return []driver.Value{
		favorite.UserID,
		favorite.TripID,
		favorite.CreatedAt,
	}
}

func normalizeRepositorySQL(query string) string {
	return strings.Join(strings.Fields(strings.ToLower(query)), " ")
}

func stringArg(arg driver.NamedValue) string {
	value, _ := arg.Value.(string)
	return value
}

func int64Arg(arg driver.NamedValue) int64 {
	switch value := arg.Value.(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case int32:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return 0
	}
}

func float64Arg(arg driver.NamedValue) float64 {
	switch value := arg.Value.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int64:
		return float64(value)
	case int:
		return float64(value)
	default:
		return 0
	}
}

func boolArg(arg driver.NamedValue) bool {
	switch value := arg.Value.(type) {
	case bool:
		return value
	case int64:
		return value != 0
	default:
		return false
	}
}

func timeArg(arg driver.NamedValue) time.Time {
	value, _ := arg.Value.(time.Time)
	return value
}

func likeKeyword(arg driver.NamedValue) string {
	return strings.ToLower(strings.Trim(stringArg(arg), "%"))
}

func matchUserKeyword(user entity.User, keyword string) bool {
	haystack := strings.ToLower(user.Email + " " + user.Phone + " " + user.Nickname)
	return strings.Contains(haystack, keyword)
}

func sortUsersDesc(items []entity.User) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].ID > items[i].ID {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func mustParseDateTime(dateText, timeText string) time.Time {
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", dateText+" "+normalizeTimeText(timeText), time.Local)
	if err != nil {
		panic(err)
	}
	return parsed
}

func normalizeTimeText(value string) string {
	if strings.Count(value, ":") == 1 {
		return value + ":00"
	}
	return value
}

func parseTripQueryOptions(sqlText string, args []driver.NamedValue) (entity.TripFilter, int, int) {
	filter := entity.TripFilter{}
	index := 0

	if strings.Contains(sqlText, "publisher_user_id = ?") {
		userID := int64Arg(args[index])
		filter.UserID = &userID
		index++
	}
	if strings.Contains(sqlText, "trip_type = ?") {
		filter.TripType = stringArg(args[index])
		index++
	}
	if count := placeholderCountInClause(sqlText, "status in ("); count > 0 {
		filter.Statuses = make([]string, 0, count)
		for i := 0; i < count; i++ {
			filter.Statuses = append(filter.Statuses, stringArg(args[index]))
			index++
		}
	}
	if count := placeholderCountInClause(sqlText, "id in ("); count > 0 {
		filter.IDs = make([]int64, 0, count)
		for i := 0; i < count; i++ {
			filter.IDs = append(filter.IDs, int64Arg(args[index]))
			index++
		}
	}
	if strings.Contains(sqlText, "(from_text like ? or to_text like ?)") {
		filter.Keyword = likeKeyword(args[index])
		index += 2
	}

	limit := 0
	offset := 0
	if strings.Contains(sqlText, " limit ") {
		if strings.Contains(sqlText, "18446744073709551615") {
			offset = int(int64Arg(args[index]))
		} else {
			limit = int(int64Arg(args[index]))
			offset = int(int64Arg(args[index+1]))
		}
	}

	return filter, offset, limit
}

func placeholderCountInClause(sqlText, marker string) int {
	start := strings.Index(sqlText, marker)
	if start < 0 {
		return 0
	}
	start += len(marker)
	end := strings.Index(sqlText[start:], ")")
	if end < 0 {
		return 0
	}
	return strings.Count(sqlText[start:start+end], "?")
}
