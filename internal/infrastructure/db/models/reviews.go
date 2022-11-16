// Code generated by SQLBoiler 4.13.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// Review is an object representing the database table.
type Review struct {
	DeviceDefinitionID string      `boil:"device_definition_id" json:"device_definition_id" toml:"device_definition_id" yaml:"device_definition_id"`
	URL                string      `boil:"url" json:"url" toml:"url" yaml:"url"`
	ImageURL           string      `boil:"image_url" json:"image_url" toml:"image_url" yaml:"image_url"`
	Channel            null.String `boil:"channel" json:"channel,omitempty" toml:"channel" yaml:"channel,omitempty"`
	Approved           bool        `boil:"approved" json:"approved" toml:"approved" yaml:"approved"`
	CreatedAt          time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt          time.Time   `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	ID                 string      `boil:"id" json:"id" toml:"id" yaml:"id"`
	Comments           string      `boil:"comments" json:"comments" toml:"comments" yaml:"comments"`
	ApprovedBy         string      `boil:"approved_by" json:"approved_by" toml:"approved_by" yaml:"approved_by"`
	Position           int         `boil:"position" json:"position" toml:"position" yaml:"position"`

	R *reviewR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L reviewL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ReviewColumns = struct {
	DeviceDefinitionID string
	URL                string
	ImageURL           string
	Channel            string
	Approved           string
	CreatedAt          string
	UpdatedAt          string
	ID                 string
	Comments           string
	ApprovedBy         string
	Position           string
}{
	DeviceDefinitionID: "device_definition_id",
	URL:                "url",
	ImageURL:           "image_url",
	Channel:            "channel",
	Approved:           "approved",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
	ID:                 "id",
	Comments:           "comments",
	ApprovedBy:         "approved_by",
	Position:           "position",
}

var ReviewTableColumns = struct {
	DeviceDefinitionID string
	URL                string
	ImageURL           string
	Channel            string
	Approved           string
	CreatedAt          string
	UpdatedAt          string
	ID                 string
	Comments           string
	ApprovedBy         string
	Position           string
}{
	DeviceDefinitionID: "reviews.device_definition_id",
	URL:                "reviews.url",
	ImageURL:           "reviews.image_url",
	Channel:            "reviews.channel",
	Approved:           "reviews.approved",
	CreatedAt:          "reviews.created_at",
	UpdatedAt:          "reviews.updated_at",
	ID:                 "reviews.id",
	Comments:           "reviews.comments",
	ApprovedBy:         "reviews.approved_by",
	Position:           "reviews.position",
}

// Generated where

var ReviewWhere = struct {
	DeviceDefinitionID whereHelperstring
	URL                whereHelperstring
	ImageURL           whereHelperstring
	Channel            whereHelpernull_String
	Approved           whereHelperbool
	CreatedAt          whereHelpertime_Time
	UpdatedAt          whereHelpertime_Time
	ID                 whereHelperstring
	Comments           whereHelperstring
	ApprovedBy         whereHelperstring
	Position           whereHelperint
}{
	DeviceDefinitionID: whereHelperstring{field: "\"device_definitions_api\".\"reviews\".\"device_definition_id\""},
	URL:                whereHelperstring{field: "\"device_definitions_api\".\"reviews\".\"url\""},
	ImageURL:           whereHelperstring{field: "\"device_definitions_api\".\"reviews\".\"image_url\""},
	Channel:            whereHelpernull_String{field: "\"device_definitions_api\".\"reviews\".\"channel\""},
	Approved:           whereHelperbool{field: "\"device_definitions_api\".\"reviews\".\"approved\""},
	CreatedAt:          whereHelpertime_Time{field: "\"device_definitions_api\".\"reviews\".\"created_at\""},
	UpdatedAt:          whereHelpertime_Time{field: "\"device_definitions_api\".\"reviews\".\"updated_at\""},
	ID:                 whereHelperstring{field: "\"device_definitions_api\".\"reviews\".\"id\""},
	Comments:           whereHelperstring{field: "\"device_definitions_api\".\"reviews\".\"comments\""},
	ApprovedBy:         whereHelperstring{field: "\"device_definitions_api\".\"reviews\".\"approved_by\""},
	Position:           whereHelperint{field: "\"device_definitions_api\".\"reviews\".\"position\""},
}

// ReviewRels is where relationship names are stored.
var ReviewRels = struct {
}{}

// reviewR is where relationships are stored.
type reviewR struct {
}

// NewStruct creates a new relationship struct
func (*reviewR) NewStruct() *reviewR {
	return &reviewR{}
}

// reviewL is where Load methods for each relationship are stored.
type reviewL struct{}

var (
	reviewAllColumns            = []string{"device_definition_id", "url", "image_url", "channel", "approved", "created_at", "updated_at", "id", "comments", "approved_by", "position"}
	reviewColumnsWithoutDefault = []string{"device_definition_id", "url", "image_url", "approved", "id", "comments", "approved_by", "position"}
	reviewColumnsWithDefault    = []string{"channel", "created_at", "updated_at"}
	reviewPrimaryKeyColumns     = []string{"id"}
	reviewGeneratedColumns      = []string{}
)

type (
	// ReviewSlice is an alias for a slice of pointers to Review.
	// This should almost always be used instead of []Review.
	ReviewSlice []*Review
	// ReviewHook is the signature for custom Review hook methods
	ReviewHook func(context.Context, boil.ContextExecutor, *Review) error

	reviewQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	reviewType                 = reflect.TypeOf(&Review{})
	reviewMapping              = queries.MakeStructMapping(reviewType)
	reviewPrimaryKeyMapping, _ = queries.BindMapping(reviewType, reviewMapping, reviewPrimaryKeyColumns)
	reviewInsertCacheMut       sync.RWMutex
	reviewInsertCache          = make(map[string]insertCache)
	reviewUpdateCacheMut       sync.RWMutex
	reviewUpdateCache          = make(map[string]updateCache)
	reviewUpsertCacheMut       sync.RWMutex
	reviewUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var reviewAfterSelectHooks []ReviewHook

var reviewBeforeInsertHooks []ReviewHook
var reviewAfterInsertHooks []ReviewHook

var reviewBeforeUpdateHooks []ReviewHook
var reviewAfterUpdateHooks []ReviewHook

var reviewBeforeDeleteHooks []ReviewHook
var reviewAfterDeleteHooks []ReviewHook

var reviewBeforeUpsertHooks []ReviewHook
var reviewAfterUpsertHooks []ReviewHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Review) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range reviewAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Review) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range reviewBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Review) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range reviewAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Review) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range reviewBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Review) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range reviewAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Review) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range reviewBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Review) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range reviewAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Review) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range reviewBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Review) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range reviewAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddReviewHook registers your hook function for all future operations.
func AddReviewHook(hookPoint boil.HookPoint, reviewHook ReviewHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		reviewAfterSelectHooks = append(reviewAfterSelectHooks, reviewHook)
	case boil.BeforeInsertHook:
		reviewBeforeInsertHooks = append(reviewBeforeInsertHooks, reviewHook)
	case boil.AfterInsertHook:
		reviewAfterInsertHooks = append(reviewAfterInsertHooks, reviewHook)
	case boil.BeforeUpdateHook:
		reviewBeforeUpdateHooks = append(reviewBeforeUpdateHooks, reviewHook)
	case boil.AfterUpdateHook:
		reviewAfterUpdateHooks = append(reviewAfterUpdateHooks, reviewHook)
	case boil.BeforeDeleteHook:
		reviewBeforeDeleteHooks = append(reviewBeforeDeleteHooks, reviewHook)
	case boil.AfterDeleteHook:
		reviewAfterDeleteHooks = append(reviewAfterDeleteHooks, reviewHook)
	case boil.BeforeUpsertHook:
		reviewBeforeUpsertHooks = append(reviewBeforeUpsertHooks, reviewHook)
	case boil.AfterUpsertHook:
		reviewAfterUpsertHooks = append(reviewAfterUpsertHooks, reviewHook)
	}
}

// One returns a single review record from the query.
func (q reviewQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Review, error) {
	o := &Review{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for reviews")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Review records from the query.
func (q reviewQuery) All(ctx context.Context, exec boil.ContextExecutor) (ReviewSlice, error) {
	var o []*Review

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Review slice")
	}

	if len(reviewAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Review records in the query.
func (q reviewQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count reviews rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q reviewQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if reviews exists")
	}

	return count > 0, nil
}

// Reviews retrieves all the records using an executor.
func Reviews(mods ...qm.QueryMod) reviewQuery {
	mods = append(mods, qm.From("\"device_definitions_api\".\"reviews\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"device_definitions_api\".\"reviews\".*"})
	}

	return reviewQuery{q}
}

// FindReview retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindReview(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*Review, error) {
	reviewObj := &Review{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"device_definitions_api\".\"reviews\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, reviewObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from reviews")
	}

	if err = reviewObj.doAfterSelectHooks(ctx, exec); err != nil {
		return reviewObj, err
	}

	return reviewObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Review) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no reviews provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = currTime
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(reviewColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	reviewInsertCacheMut.RLock()
	cache, cached := reviewInsertCache[key]
	reviewInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			reviewAllColumns,
			reviewColumnsWithDefault,
			reviewColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(reviewType, reviewMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(reviewType, reviewMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"device_definitions_api\".\"reviews\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"device_definitions_api\".\"reviews\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into reviews")
	}

	if !cached {
		reviewInsertCacheMut.Lock()
		reviewInsertCache[key] = cache
		reviewInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Review.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Review) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	reviewUpdateCacheMut.RLock()
	cache, cached := reviewUpdateCache[key]
	reviewUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			reviewAllColumns,
			reviewPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update reviews, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"device_definitions_api\".\"reviews\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, reviewPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(reviewType, reviewMapping, append(wl, reviewPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update reviews row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for reviews")
	}

	if !cached {
		reviewUpdateCacheMut.Lock()
		reviewUpdateCache[key] = cache
		reviewUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q reviewQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for reviews")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for reviews")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ReviewSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), reviewPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"device_definitions_api\".\"reviews\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, reviewPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in review slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all review")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Review) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no reviews provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
		o.UpdatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(reviewColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	reviewUpsertCacheMut.RLock()
	cache, cached := reviewUpsertCache[key]
	reviewUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			reviewAllColumns,
			reviewColumnsWithDefault,
			reviewColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			reviewAllColumns,
			reviewPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert reviews, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(reviewPrimaryKeyColumns))
			copy(conflict, reviewPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"device_definitions_api\".\"reviews\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(reviewType, reviewMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(reviewType, reviewMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert reviews")
	}

	if !cached {
		reviewUpsertCacheMut.Lock()
		reviewUpsertCache[key] = cache
		reviewUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Review record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Review) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Review provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), reviewPrimaryKeyMapping)
	sql := "DELETE FROM \"device_definitions_api\".\"reviews\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from reviews")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for reviews")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q reviewQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no reviewQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from reviews")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for reviews")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ReviewSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(reviewBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), reviewPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"device_definitions_api\".\"reviews\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, reviewPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from review slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for reviews")
	}

	if len(reviewAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Review) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindReview(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ReviewSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ReviewSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), reviewPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"device_definitions_api\".\"reviews\".* FROM \"device_definitions_api\".\"reviews\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, reviewPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in ReviewSlice")
	}

	*o = slice

	return nil
}

// ReviewExists checks if the Review row exists.
func ReviewExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"device_definitions_api\".\"reviews\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if reviews exists")
	}

	return exists, nil
}
