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

// IntegrationFeature is an object representing the database table.
type IntegrationFeature struct {
	FeatureKey      string       `boil:"feature_key" json:"feature_key" toml:"feature_key" yaml:"feature_key"`
	ElasticProperty string       `boil:"elastic_property" json:"elastic_property" toml:"elastic_property" yaml:"elastic_property"`
	DisplayName     string       `boil:"display_name" json:"display_name" toml:"display_name" yaml:"display_name"`
	CSSIcon         null.String  `boil:"css_icon" json:"css_icon,omitempty" toml:"css_icon" yaml:"css_icon,omitempty"`
	CreatedAt       time.Time    `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt       time.Time    `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	FeatureWeight   null.Float64 `boil:"feature_weight" json:"feature_weight,omitempty" toml:"feature_weight" yaml:"feature_weight,omitempty"`

	R *integrationFeatureR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L integrationFeatureL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var IntegrationFeatureColumns = struct {
	FeatureKey      string
	ElasticProperty string
	DisplayName     string
	CSSIcon         string
	CreatedAt       string
	UpdatedAt       string
	FeatureWeight   string
}{
	FeatureKey:      "feature_key",
	ElasticProperty: "elastic_property",
	DisplayName:     "display_name",
	CSSIcon:         "css_icon",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	FeatureWeight:   "feature_weight",
}

var IntegrationFeatureTableColumns = struct {
	FeatureKey      string
	ElasticProperty string
	DisplayName     string
	CSSIcon         string
	CreatedAt       string
	UpdatedAt       string
	FeatureWeight   string
}{
	FeatureKey:      "integration_features.feature_key",
	ElasticProperty: "integration_features.elastic_property",
	DisplayName:     "integration_features.display_name",
	CSSIcon:         "integration_features.css_icon",
	CreatedAt:       "integration_features.created_at",
	UpdatedAt:       "integration_features.updated_at",
	FeatureWeight:   "integration_features.feature_weight",
}

// Generated where

type whereHelpernull_Float64 struct{ field string }

func (w whereHelpernull_Float64) EQ(x null.Float64) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Float64) NEQ(x null.Float64) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Float64) LT(x null.Float64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Float64) LTE(x null.Float64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Float64) GT(x null.Float64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Float64) GTE(x null.Float64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}
func (w whereHelpernull_Float64) IN(slice []float64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelpernull_Float64) NIN(slice []float64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

func (w whereHelpernull_Float64) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Float64) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

var IntegrationFeatureWhere = struct {
	FeatureKey      whereHelperstring
	ElasticProperty whereHelperstring
	DisplayName     whereHelperstring
	CSSIcon         whereHelpernull_String
	CreatedAt       whereHelpertime_Time
	UpdatedAt       whereHelpertime_Time
	FeatureWeight   whereHelpernull_Float64
}{
	FeatureKey:      whereHelperstring{field: "\"device_definitions_api\".\"integration_features\".\"feature_key\""},
	ElasticProperty: whereHelperstring{field: "\"device_definitions_api\".\"integration_features\".\"elastic_property\""},
	DisplayName:     whereHelperstring{field: "\"device_definitions_api\".\"integration_features\".\"display_name\""},
	CSSIcon:         whereHelpernull_String{field: "\"device_definitions_api\".\"integration_features\".\"css_icon\""},
	CreatedAt:       whereHelpertime_Time{field: "\"device_definitions_api\".\"integration_features\".\"created_at\""},
	UpdatedAt:       whereHelpertime_Time{field: "\"device_definitions_api\".\"integration_features\".\"updated_at\""},
	FeatureWeight:   whereHelpernull_Float64{field: "\"device_definitions_api\".\"integration_features\".\"feature_weight\""},
}

// IntegrationFeatureRels is where relationship names are stored.
var IntegrationFeatureRels = struct {
}{}

// integrationFeatureR is where relationships are stored.
type integrationFeatureR struct {
}

// NewStruct creates a new relationship struct
func (*integrationFeatureR) NewStruct() *integrationFeatureR {
	return &integrationFeatureR{}
}

// integrationFeatureL is where Load methods for each relationship are stored.
type integrationFeatureL struct{}

var (
	integrationFeatureAllColumns            = []string{"feature_key", "elastic_property", "display_name", "css_icon", "created_at", "updated_at", "feature_weight"}
	integrationFeatureColumnsWithoutDefault = []string{"feature_key", "elastic_property", "display_name"}
	integrationFeatureColumnsWithDefault    = []string{"css_icon", "created_at", "updated_at", "feature_weight"}
	integrationFeaturePrimaryKeyColumns     = []string{"feature_key"}
	integrationFeatureGeneratedColumns      = []string{}
)

type (
	// IntegrationFeatureSlice is an alias for a slice of pointers to IntegrationFeature.
	// This should almost always be used instead of []IntegrationFeature.
	IntegrationFeatureSlice []*IntegrationFeature
	// IntegrationFeatureHook is the signature for custom IntegrationFeature hook methods
	IntegrationFeatureHook func(context.Context, boil.ContextExecutor, *IntegrationFeature) error

	integrationFeatureQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	integrationFeatureType                 = reflect.TypeOf(&IntegrationFeature{})
	integrationFeatureMapping              = queries.MakeStructMapping(integrationFeatureType)
	integrationFeaturePrimaryKeyMapping, _ = queries.BindMapping(integrationFeatureType, integrationFeatureMapping, integrationFeaturePrimaryKeyColumns)
	integrationFeatureInsertCacheMut       sync.RWMutex
	integrationFeatureInsertCache          = make(map[string]insertCache)
	integrationFeatureUpdateCacheMut       sync.RWMutex
	integrationFeatureUpdateCache          = make(map[string]updateCache)
	integrationFeatureUpsertCacheMut       sync.RWMutex
	integrationFeatureUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var integrationFeatureAfterSelectHooks []IntegrationFeatureHook

var integrationFeatureBeforeInsertHooks []IntegrationFeatureHook
var integrationFeatureAfterInsertHooks []IntegrationFeatureHook

var integrationFeatureBeforeUpdateHooks []IntegrationFeatureHook
var integrationFeatureAfterUpdateHooks []IntegrationFeatureHook

var integrationFeatureBeforeDeleteHooks []IntegrationFeatureHook
var integrationFeatureAfterDeleteHooks []IntegrationFeatureHook

var integrationFeatureBeforeUpsertHooks []IntegrationFeatureHook
var integrationFeatureAfterUpsertHooks []IntegrationFeatureHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *IntegrationFeature) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationFeatureAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *IntegrationFeature) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationFeatureBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *IntegrationFeature) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationFeatureAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *IntegrationFeature) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationFeatureBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *IntegrationFeature) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationFeatureAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *IntegrationFeature) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationFeatureBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *IntegrationFeature) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationFeatureAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *IntegrationFeature) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationFeatureBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *IntegrationFeature) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationFeatureAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddIntegrationFeatureHook registers your hook function for all future operations.
func AddIntegrationFeatureHook(hookPoint boil.HookPoint, integrationFeatureHook IntegrationFeatureHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		integrationFeatureAfterSelectHooks = append(integrationFeatureAfterSelectHooks, integrationFeatureHook)
	case boil.BeforeInsertHook:
		integrationFeatureBeforeInsertHooks = append(integrationFeatureBeforeInsertHooks, integrationFeatureHook)
	case boil.AfterInsertHook:
		integrationFeatureAfterInsertHooks = append(integrationFeatureAfterInsertHooks, integrationFeatureHook)
	case boil.BeforeUpdateHook:
		integrationFeatureBeforeUpdateHooks = append(integrationFeatureBeforeUpdateHooks, integrationFeatureHook)
	case boil.AfterUpdateHook:
		integrationFeatureAfterUpdateHooks = append(integrationFeatureAfterUpdateHooks, integrationFeatureHook)
	case boil.BeforeDeleteHook:
		integrationFeatureBeforeDeleteHooks = append(integrationFeatureBeforeDeleteHooks, integrationFeatureHook)
	case boil.AfterDeleteHook:
		integrationFeatureAfterDeleteHooks = append(integrationFeatureAfterDeleteHooks, integrationFeatureHook)
	case boil.BeforeUpsertHook:
		integrationFeatureBeforeUpsertHooks = append(integrationFeatureBeforeUpsertHooks, integrationFeatureHook)
	case boil.AfterUpsertHook:
		integrationFeatureAfterUpsertHooks = append(integrationFeatureAfterUpsertHooks, integrationFeatureHook)
	}
}

// One returns a single integrationFeature record from the query.
func (q integrationFeatureQuery) One(ctx context.Context, exec boil.ContextExecutor) (*IntegrationFeature, error) {
	o := &IntegrationFeature{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for integration_features")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all IntegrationFeature records from the query.
func (q integrationFeatureQuery) All(ctx context.Context, exec boil.ContextExecutor) (IntegrationFeatureSlice, error) {
	var o []*IntegrationFeature

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to IntegrationFeature slice")
	}

	if len(integrationFeatureAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all IntegrationFeature records in the query.
func (q integrationFeatureQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count integration_features rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q integrationFeatureQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if integration_features exists")
	}

	return count > 0, nil
}

// IntegrationFeatures retrieves all the records using an executor.
func IntegrationFeatures(mods ...qm.QueryMod) integrationFeatureQuery {
	mods = append(mods, qm.From("\"device_definitions_api\".\"integration_features\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"device_definitions_api\".\"integration_features\".*"})
	}

	return integrationFeatureQuery{q}
}

// FindIntegrationFeature retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindIntegrationFeature(ctx context.Context, exec boil.ContextExecutor, featureKey string, selectCols ...string) (*IntegrationFeature, error) {
	integrationFeatureObj := &IntegrationFeature{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"device_definitions_api\".\"integration_features\" where \"feature_key\"=$1", sel,
	)

	q := queries.Raw(query, featureKey)

	err := q.Bind(ctx, exec, integrationFeatureObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from integration_features")
	}

	if err = integrationFeatureObj.doAfterSelectHooks(ctx, exec); err != nil {
		return integrationFeatureObj, err
	}

	return integrationFeatureObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *IntegrationFeature) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no integration_features provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(integrationFeatureColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	integrationFeatureInsertCacheMut.RLock()
	cache, cached := integrationFeatureInsertCache[key]
	integrationFeatureInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			integrationFeatureAllColumns,
			integrationFeatureColumnsWithDefault,
			integrationFeatureColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(integrationFeatureType, integrationFeatureMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(integrationFeatureType, integrationFeatureMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"device_definitions_api\".\"integration_features\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"device_definitions_api\".\"integration_features\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into integration_features")
	}

	if !cached {
		integrationFeatureInsertCacheMut.Lock()
		integrationFeatureInsertCache[key] = cache
		integrationFeatureInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the IntegrationFeature.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *IntegrationFeature) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	integrationFeatureUpdateCacheMut.RLock()
	cache, cached := integrationFeatureUpdateCache[key]
	integrationFeatureUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			integrationFeatureAllColumns,
			integrationFeaturePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update integration_features, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"device_definitions_api\".\"integration_features\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, integrationFeaturePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(integrationFeatureType, integrationFeatureMapping, append(wl, integrationFeaturePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update integration_features row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for integration_features")
	}

	if !cached {
		integrationFeatureUpdateCacheMut.Lock()
		integrationFeatureUpdateCache[key] = cache
		integrationFeatureUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q integrationFeatureQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for integration_features")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for integration_features")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o IntegrationFeatureSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), integrationFeaturePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"device_definitions_api\".\"integration_features\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, integrationFeaturePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in integrationFeature slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all integrationFeature")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *IntegrationFeature) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no integration_features provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(integrationFeatureColumnsWithDefault, o)

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

	integrationFeatureUpsertCacheMut.RLock()
	cache, cached := integrationFeatureUpsertCache[key]
	integrationFeatureUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			integrationFeatureAllColumns,
			integrationFeatureColumnsWithDefault,
			integrationFeatureColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			integrationFeatureAllColumns,
			integrationFeaturePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert integration_features, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(integrationFeaturePrimaryKeyColumns))
			copy(conflict, integrationFeaturePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"device_definitions_api\".\"integration_features\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(integrationFeatureType, integrationFeatureMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(integrationFeatureType, integrationFeatureMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert integration_features")
	}

	if !cached {
		integrationFeatureUpsertCacheMut.Lock()
		integrationFeatureUpsertCache[key] = cache
		integrationFeatureUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single IntegrationFeature record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *IntegrationFeature) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no IntegrationFeature provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), integrationFeaturePrimaryKeyMapping)
	sql := "DELETE FROM \"device_definitions_api\".\"integration_features\" WHERE \"feature_key\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from integration_features")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for integration_features")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q integrationFeatureQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no integrationFeatureQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from integration_features")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for integration_features")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o IntegrationFeatureSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(integrationFeatureBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), integrationFeaturePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"device_definitions_api\".\"integration_features\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, integrationFeaturePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from integrationFeature slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for integration_features")
	}

	if len(integrationFeatureAfterDeleteHooks) != 0 {
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
func (o *IntegrationFeature) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindIntegrationFeature(ctx, exec, o.FeatureKey)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *IntegrationFeatureSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := IntegrationFeatureSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), integrationFeaturePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"device_definitions_api\".\"integration_features\".* FROM \"device_definitions_api\".\"integration_features\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, integrationFeaturePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in IntegrationFeatureSlice")
	}

	*o = slice

	return nil
}

// IntegrationFeatureExists checks if the IntegrationFeature row exists.
func IntegrationFeatureExists(ctx context.Context, exec boil.ContextExecutor, featureKey string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"device_definitions_api\".\"integration_features\" where \"feature_key\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, featureKey)
	}
	row := exec.QueryRowContext(ctx, sql, featureKey)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if integration_features exists")
	}

	return exists, nil
}
