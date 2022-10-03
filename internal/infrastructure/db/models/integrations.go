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

// Integration is an object representing the database table.
type Integration struct {
	ID        string    `boil:"id" json:"id" toml:"id" yaml:"id"`
	Type      string    `boil:"type" json:"type" toml:"type" yaml:"type"`
	Style     string    `boil:"style" json:"style" toml:"style" yaml:"style"`
	Vendor    string    `boil:"vendor" json:"vendor" toml:"vendor" yaml:"vendor"`
	CreatedAt time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	// How often can integration be called in seconds
	RefreshLimitSecs int       `boil:"refresh_limit_secs" json:"refresh_limit_secs" toml:"refresh_limit_secs" yaml:"refresh_limit_secs"`
	Metadata         null.JSON `boil:"metadata" json:"metadata,omitempty" toml:"metadata" yaml:"metadata,omitempty"`

	R *integrationR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L integrationL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var IntegrationColumns = struct {
	ID               string
	Type             string
	Style            string
	Vendor           string
	CreatedAt        string
	UpdatedAt        string
	RefreshLimitSecs string
	Metadata         string
}{
	ID:               "id",
	Type:             "type",
	Style:            "style",
	Vendor:           "vendor",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
	RefreshLimitSecs: "refresh_limit_secs",
	Metadata:         "metadata",
}

var IntegrationTableColumns = struct {
	ID               string
	Type             string
	Style            string
	Vendor           string
	CreatedAt        string
	UpdatedAt        string
	RefreshLimitSecs string
	Metadata         string
}{
	ID:               "integrations.id",
	Type:             "integrations.type",
	Style:            "integrations.style",
	Vendor:           "integrations.vendor",
	CreatedAt:        "integrations.created_at",
	UpdatedAt:        "integrations.updated_at",
	RefreshLimitSecs: "integrations.refresh_limit_secs",
	Metadata:         "integrations.metadata",
}

// Generated where

type whereHelperint struct{ field string }

func (w whereHelperint) EQ(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint) NEQ(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint) LT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint) LTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint) GT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint) GTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint) IN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint) NIN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

var IntegrationWhere = struct {
	ID               whereHelperstring
	Type             whereHelperstring
	Style            whereHelperstring
	Vendor           whereHelperstring
	CreatedAt        whereHelpertime_Time
	UpdatedAt        whereHelpertime_Time
	RefreshLimitSecs whereHelperint
	Metadata         whereHelpernull_JSON
}{
	ID:               whereHelperstring{field: "\"device_definitions_api\".\"integrations\".\"id\""},
	Type:             whereHelperstring{field: "\"device_definitions_api\".\"integrations\".\"type\""},
	Style:            whereHelperstring{field: "\"device_definitions_api\".\"integrations\".\"style\""},
	Vendor:           whereHelperstring{field: "\"device_definitions_api\".\"integrations\".\"vendor\""},
	CreatedAt:        whereHelpertime_Time{field: "\"device_definitions_api\".\"integrations\".\"created_at\""},
	UpdatedAt:        whereHelpertime_Time{field: "\"device_definitions_api\".\"integrations\".\"updated_at\""},
	RefreshLimitSecs: whereHelperint{field: "\"device_definitions_api\".\"integrations\".\"refresh_limit_secs\""},
	Metadata:         whereHelpernull_JSON{field: "\"device_definitions_api\".\"integrations\".\"metadata\""},
}

// IntegrationRels is where relationship names are stored.
var IntegrationRels = struct {
	DeviceIntegrations string
}{
	DeviceIntegrations: "DeviceIntegrations",
}

// integrationR is where relationships are stored.
type integrationR struct {
	DeviceIntegrations DeviceIntegrationSlice `boil:"DeviceIntegrations" json:"DeviceIntegrations" toml:"DeviceIntegrations" yaml:"DeviceIntegrations"`
}

// NewStruct creates a new relationship struct
func (*integrationR) NewStruct() *integrationR {
	return &integrationR{}
}

func (r *integrationR) GetDeviceIntegrations() DeviceIntegrationSlice {
	if r == nil {
		return nil
	}
	return r.DeviceIntegrations
}

// integrationL is where Load methods for each relationship are stored.
type integrationL struct{}

var (
	integrationAllColumns            = []string{"id", "type", "style", "vendor", "created_at", "updated_at", "refresh_limit_secs", "metadata"}
	integrationColumnsWithoutDefault = []string{"id", "type", "style", "vendor"}
	integrationColumnsWithDefault    = []string{"created_at", "updated_at", "refresh_limit_secs", "metadata"}
	integrationPrimaryKeyColumns     = []string{"id"}
	integrationGeneratedColumns      = []string{}
)

type (
	// IntegrationSlice is an alias for a slice of pointers to Integration.
	// This should almost always be used instead of []Integration.
	IntegrationSlice []*Integration
	// IntegrationHook is the signature for custom Integration hook methods
	IntegrationHook func(context.Context, boil.ContextExecutor, *Integration) error

	integrationQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	integrationType                 = reflect.TypeOf(&Integration{})
	integrationMapping              = queries.MakeStructMapping(integrationType)
	integrationPrimaryKeyMapping, _ = queries.BindMapping(integrationType, integrationMapping, integrationPrimaryKeyColumns)
	integrationInsertCacheMut       sync.RWMutex
	integrationInsertCache          = make(map[string]insertCache)
	integrationUpdateCacheMut       sync.RWMutex
	integrationUpdateCache          = make(map[string]updateCache)
	integrationUpsertCacheMut       sync.RWMutex
	integrationUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var integrationAfterSelectHooks []IntegrationHook

var integrationBeforeInsertHooks []IntegrationHook
var integrationAfterInsertHooks []IntegrationHook

var integrationBeforeUpdateHooks []IntegrationHook
var integrationAfterUpdateHooks []IntegrationHook

var integrationBeforeDeleteHooks []IntegrationHook
var integrationAfterDeleteHooks []IntegrationHook

var integrationBeforeUpsertHooks []IntegrationHook
var integrationAfterUpsertHooks []IntegrationHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Integration) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Integration) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Integration) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Integration) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Integration) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Integration) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Integration) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Integration) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Integration) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range integrationAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddIntegrationHook registers your hook function for all future operations.
func AddIntegrationHook(hookPoint boil.HookPoint, integrationHook IntegrationHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		integrationAfterSelectHooks = append(integrationAfterSelectHooks, integrationHook)
	case boil.BeforeInsertHook:
		integrationBeforeInsertHooks = append(integrationBeforeInsertHooks, integrationHook)
	case boil.AfterInsertHook:
		integrationAfterInsertHooks = append(integrationAfterInsertHooks, integrationHook)
	case boil.BeforeUpdateHook:
		integrationBeforeUpdateHooks = append(integrationBeforeUpdateHooks, integrationHook)
	case boil.AfterUpdateHook:
		integrationAfterUpdateHooks = append(integrationAfterUpdateHooks, integrationHook)
	case boil.BeforeDeleteHook:
		integrationBeforeDeleteHooks = append(integrationBeforeDeleteHooks, integrationHook)
	case boil.AfterDeleteHook:
		integrationAfterDeleteHooks = append(integrationAfterDeleteHooks, integrationHook)
	case boil.BeforeUpsertHook:
		integrationBeforeUpsertHooks = append(integrationBeforeUpsertHooks, integrationHook)
	case boil.AfterUpsertHook:
		integrationAfterUpsertHooks = append(integrationAfterUpsertHooks, integrationHook)
	}
}

// One returns a single integration record from the query.
func (q integrationQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Integration, error) {
	o := &Integration{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for integrations")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Integration records from the query.
func (q integrationQuery) All(ctx context.Context, exec boil.ContextExecutor) (IntegrationSlice, error) {
	var o []*Integration

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Integration slice")
	}

	if len(integrationAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Integration records in the query.
func (q integrationQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count integrations rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q integrationQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if integrations exists")
	}

	return count > 0, nil
}

// DeviceIntegrations retrieves all the device_integration's DeviceIntegrations with an executor.
func (o *Integration) DeviceIntegrations(mods ...qm.QueryMod) deviceIntegrationQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"device_definitions_api\".\"device_integrations\".\"integration_id\"=?", o.ID),
	)

	return DeviceIntegrations(queryMods...)
}

// LoadDeviceIntegrations allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (integrationL) LoadDeviceIntegrations(ctx context.Context, e boil.ContextExecutor, singular bool, maybeIntegration interface{}, mods queries.Applicator) error {
	var slice []*Integration
	var object *Integration

	if singular {
		var ok bool
		object, ok = maybeIntegration.(*Integration)
		if !ok {
			object = new(Integration)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeIntegration)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeIntegration))
			}
		}
	} else {
		s, ok := maybeIntegration.(*[]*Integration)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeIntegration)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeIntegration))
			}
		}
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &integrationR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &integrationR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`device_definitions_api.device_integrations`),
		qm.WhereIn(`device_definitions_api.device_integrations.integration_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load device_integrations")
	}

	var resultSlice []*DeviceIntegration
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice device_integrations")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on device_integrations")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for device_integrations")
	}

	if len(deviceIntegrationAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.DeviceIntegrations = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &deviceIntegrationR{}
			}
			foreign.R.Integration = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.IntegrationID {
				local.R.DeviceIntegrations = append(local.R.DeviceIntegrations, foreign)
				if foreign.R == nil {
					foreign.R = &deviceIntegrationR{}
				}
				foreign.R.Integration = local
				break
			}
		}
	}

	return nil
}

// AddDeviceIntegrations adds the given related objects to the existing relationships
// of the integration, optionally inserting them as new records.
// Appends related to o.R.DeviceIntegrations.
// Sets related.R.Integration appropriately.
func (o *Integration) AddDeviceIntegrations(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*DeviceIntegration) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.IntegrationID = o.ID
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"device_definitions_api\".\"device_integrations\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"integration_id"}),
				strmangle.WhereClause("\"", "\"", 2, deviceIntegrationPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.DeviceDefinitionID, rel.IntegrationID, rel.Region}

			if boil.IsDebug(ctx) {
				writer := boil.DebugWriterFrom(ctx)
				fmt.Fprintln(writer, updateQuery)
				fmt.Fprintln(writer, values)
			}
			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.IntegrationID = o.ID
		}
	}

	if o.R == nil {
		o.R = &integrationR{
			DeviceIntegrations: related,
		}
	} else {
		o.R.DeviceIntegrations = append(o.R.DeviceIntegrations, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &deviceIntegrationR{
				Integration: o,
			}
		} else {
			rel.R.Integration = o
		}
	}
	return nil
}

// Integrations retrieves all the records using an executor.
func Integrations(mods ...qm.QueryMod) integrationQuery {
	mods = append(mods, qm.From("\"device_definitions_api\".\"integrations\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"device_definitions_api\".\"integrations\".*"})
	}

	return integrationQuery{q}
}

// FindIntegration retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindIntegration(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*Integration, error) {
	integrationObj := &Integration{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"device_definitions_api\".\"integrations\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, integrationObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from integrations")
	}

	if err = integrationObj.doAfterSelectHooks(ctx, exec); err != nil {
		return integrationObj, err
	}

	return integrationObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Integration) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no integrations provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(integrationColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	integrationInsertCacheMut.RLock()
	cache, cached := integrationInsertCache[key]
	integrationInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			integrationAllColumns,
			integrationColumnsWithDefault,
			integrationColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(integrationType, integrationMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(integrationType, integrationMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"device_definitions_api\".\"integrations\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"device_definitions_api\".\"integrations\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into integrations")
	}

	if !cached {
		integrationInsertCacheMut.Lock()
		integrationInsertCache[key] = cache
		integrationInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Integration.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Integration) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	integrationUpdateCacheMut.RLock()
	cache, cached := integrationUpdateCache[key]
	integrationUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			integrationAllColumns,
			integrationPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update integrations, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"device_definitions_api\".\"integrations\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, integrationPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(integrationType, integrationMapping, append(wl, integrationPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update integrations row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for integrations")
	}

	if !cached {
		integrationUpdateCacheMut.Lock()
		integrationUpdateCache[key] = cache
		integrationUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q integrationQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for integrations")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for integrations")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o IntegrationSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), integrationPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"device_definitions_api\".\"integrations\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, integrationPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in integration slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all integration")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Integration) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no integrations provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(integrationColumnsWithDefault, o)

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

	integrationUpsertCacheMut.RLock()
	cache, cached := integrationUpsertCache[key]
	integrationUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			integrationAllColumns,
			integrationColumnsWithDefault,
			integrationColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			integrationAllColumns,
			integrationPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert integrations, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(integrationPrimaryKeyColumns))
			copy(conflict, integrationPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"device_definitions_api\".\"integrations\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(integrationType, integrationMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(integrationType, integrationMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert integrations")
	}

	if !cached {
		integrationUpsertCacheMut.Lock()
		integrationUpsertCache[key] = cache
		integrationUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Integration record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Integration) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Integration provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), integrationPrimaryKeyMapping)
	sql := "DELETE FROM \"device_definitions_api\".\"integrations\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from integrations")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for integrations")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q integrationQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no integrationQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from integrations")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for integrations")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o IntegrationSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(integrationBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), integrationPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"device_definitions_api\".\"integrations\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, integrationPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from integration slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for integrations")
	}

	if len(integrationAfterDeleteHooks) != 0 {
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
func (o *Integration) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindIntegration(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *IntegrationSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := IntegrationSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), integrationPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"device_definitions_api\".\"integrations\".* FROM \"device_definitions_api\".\"integrations\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, integrationPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in IntegrationSlice")
	}

	*o = slice

	return nil
}

// IntegrationExists checks if the Integration row exists.
func IntegrationExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"device_definitions_api\".\"integrations\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if integrations exists")
	}

	return exists, nil
}
