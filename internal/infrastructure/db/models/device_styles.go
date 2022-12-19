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

// DeviceStyle is an object representing the database table.
type DeviceStyle struct {
	ID                 string      `boil:"id" json:"id" toml:"id" yaml:"id"`
	DeviceDefinitionID string      `boil:"device_definition_id" json:"device_definition_id" toml:"device_definition_id" yaml:"device_definition_id"`
	Name               string      `boil:"name" json:"name" toml:"name" yaml:"name"`
	ExternalStyleID    string      `boil:"external_style_id" json:"external_style_id" toml:"external_style_id" yaml:"external_style_id"`
	Source             string      `boil:"source" json:"source" toml:"source" yaml:"source"`
	CreatedAt          time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt          time.Time   `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	SubModel           string      `boil:"sub_model" json:"sub_model" toml:"sub_model" yaml:"sub_model"`
	HardwareTemplateID null.String `boil:"hardware_template_id" json:"hardware_template_id,omitempty" toml:"hardware_template_id" yaml:"hardware_template_id,omitempty"`

	R *deviceStyleR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L deviceStyleL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var DeviceStyleColumns = struct {
	ID                 string
	DeviceDefinitionID string
	Name               string
	ExternalStyleID    string
	Source             string
	CreatedAt          string
	UpdatedAt          string
	SubModel           string
	HardwareTemplateID string
}{
	ID:                 "id",
	DeviceDefinitionID: "device_definition_id",
	Name:               "name",
	ExternalStyleID:    "external_style_id",
	Source:             "source",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
	SubModel:           "sub_model",
	HardwareTemplateID: "hardware_template_id",
}

var DeviceStyleTableColumns = struct {
	ID                 string
	DeviceDefinitionID string
	Name               string
	ExternalStyleID    string
	Source             string
	CreatedAt          string
	UpdatedAt          string
	SubModel           string
	HardwareTemplateID string
}{
	ID:                 "device_styles.id",
	DeviceDefinitionID: "device_styles.device_definition_id",
	Name:               "device_styles.name",
	ExternalStyleID:    "device_styles.external_style_id",
	Source:             "device_styles.source",
	CreatedAt:          "device_styles.created_at",
	UpdatedAt:          "device_styles.updated_at",
	SubModel:           "device_styles.sub_model",
	HardwareTemplateID: "device_styles.hardware_template_id",
}

// Generated where

var DeviceStyleWhere = struct {
	ID                 whereHelperstring
	DeviceDefinitionID whereHelperstring
	Name               whereHelperstring
	ExternalStyleID    whereHelperstring
	Source             whereHelperstring
	CreatedAt          whereHelpertime_Time
	UpdatedAt          whereHelpertime_Time
	SubModel           whereHelperstring
	HardwareTemplateID whereHelpernull_String
}{
	ID:                 whereHelperstring{field: "\"device_definitions_api\".\"device_styles\".\"id\""},
	DeviceDefinitionID: whereHelperstring{field: "\"device_definitions_api\".\"device_styles\".\"device_definition_id\""},
	Name:               whereHelperstring{field: "\"device_definitions_api\".\"device_styles\".\"name\""},
	ExternalStyleID:    whereHelperstring{field: "\"device_definitions_api\".\"device_styles\".\"external_style_id\""},
	Source:             whereHelperstring{field: "\"device_definitions_api\".\"device_styles\".\"source\""},
	CreatedAt:          whereHelpertime_Time{field: "\"device_definitions_api\".\"device_styles\".\"created_at\""},
	UpdatedAt:          whereHelpertime_Time{field: "\"device_definitions_api\".\"device_styles\".\"updated_at\""},
	SubModel:           whereHelperstring{field: "\"device_definitions_api\".\"device_styles\".\"sub_model\""},
	HardwareTemplateID: whereHelpernull_String{field: "\"device_definitions_api\".\"device_styles\".\"hardware_template_id\""},
}

// DeviceStyleRels is where relationship names are stored.
var DeviceStyleRels = struct {
	DeviceDefinition string
}{
	DeviceDefinition: "DeviceDefinition",
}

// deviceStyleR is where relationships are stored.
type deviceStyleR struct {
	DeviceDefinition *DeviceDefinition `boil:"DeviceDefinition" json:"DeviceDefinition" toml:"DeviceDefinition" yaml:"DeviceDefinition"`
}

// NewStruct creates a new relationship struct
func (*deviceStyleR) NewStruct() *deviceStyleR {
	return &deviceStyleR{}
}

func (r *deviceStyleR) GetDeviceDefinition() *DeviceDefinition {
	if r == nil {
		return nil
	}
	return r.DeviceDefinition
}

// deviceStyleL is where Load methods for each relationship are stored.
type deviceStyleL struct{}

var (
	deviceStyleAllColumns            = []string{"id", "device_definition_id", "name", "external_style_id", "source", "created_at", "updated_at", "sub_model", "hardware_template_id"}
	deviceStyleColumnsWithoutDefault = []string{"id", "device_definition_id", "name", "external_style_id", "source", "sub_model"}
	deviceStyleColumnsWithDefault    = []string{"created_at", "updated_at", "hardware_template_id"}
	deviceStylePrimaryKeyColumns     = []string{"id"}
	deviceStyleGeneratedColumns      = []string{}
)

type (
	// DeviceStyleSlice is an alias for a slice of pointers to DeviceStyle.
	// This should almost always be used instead of []DeviceStyle.
	DeviceStyleSlice []*DeviceStyle
	// DeviceStyleHook is the signature for custom DeviceStyle hook methods
	DeviceStyleHook func(context.Context, boil.ContextExecutor, *DeviceStyle) error

	deviceStyleQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	deviceStyleType                 = reflect.TypeOf(&DeviceStyle{})
	deviceStyleMapping              = queries.MakeStructMapping(deviceStyleType)
	deviceStylePrimaryKeyMapping, _ = queries.BindMapping(deviceStyleType, deviceStyleMapping, deviceStylePrimaryKeyColumns)
	deviceStyleInsertCacheMut       sync.RWMutex
	deviceStyleInsertCache          = make(map[string]insertCache)
	deviceStyleUpdateCacheMut       sync.RWMutex
	deviceStyleUpdateCache          = make(map[string]updateCache)
	deviceStyleUpsertCacheMut       sync.RWMutex
	deviceStyleUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var deviceStyleAfterSelectHooks []DeviceStyleHook

var deviceStyleBeforeInsertHooks []DeviceStyleHook
var deviceStyleAfterInsertHooks []DeviceStyleHook

var deviceStyleBeforeUpdateHooks []DeviceStyleHook
var deviceStyleAfterUpdateHooks []DeviceStyleHook

var deviceStyleBeforeDeleteHooks []DeviceStyleHook
var deviceStyleAfterDeleteHooks []DeviceStyleHook

var deviceStyleBeforeUpsertHooks []DeviceStyleHook
var deviceStyleAfterUpsertHooks []DeviceStyleHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *DeviceStyle) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceStyleAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *DeviceStyle) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceStyleBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *DeviceStyle) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceStyleAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *DeviceStyle) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceStyleBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *DeviceStyle) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceStyleAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *DeviceStyle) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceStyleBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *DeviceStyle) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceStyleAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *DeviceStyle) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceStyleBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *DeviceStyle) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceStyleAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddDeviceStyleHook registers your hook function for all future operations.
func AddDeviceStyleHook(hookPoint boil.HookPoint, deviceStyleHook DeviceStyleHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		deviceStyleAfterSelectHooks = append(deviceStyleAfterSelectHooks, deviceStyleHook)
	case boil.BeforeInsertHook:
		deviceStyleBeforeInsertHooks = append(deviceStyleBeforeInsertHooks, deviceStyleHook)
	case boil.AfterInsertHook:
		deviceStyleAfterInsertHooks = append(deviceStyleAfterInsertHooks, deviceStyleHook)
	case boil.BeforeUpdateHook:
		deviceStyleBeforeUpdateHooks = append(deviceStyleBeforeUpdateHooks, deviceStyleHook)
	case boil.AfterUpdateHook:
		deviceStyleAfterUpdateHooks = append(deviceStyleAfterUpdateHooks, deviceStyleHook)
	case boil.BeforeDeleteHook:
		deviceStyleBeforeDeleteHooks = append(deviceStyleBeforeDeleteHooks, deviceStyleHook)
	case boil.AfterDeleteHook:
		deviceStyleAfterDeleteHooks = append(deviceStyleAfterDeleteHooks, deviceStyleHook)
	case boil.BeforeUpsertHook:
		deviceStyleBeforeUpsertHooks = append(deviceStyleBeforeUpsertHooks, deviceStyleHook)
	case boil.AfterUpsertHook:
		deviceStyleAfterUpsertHooks = append(deviceStyleAfterUpsertHooks, deviceStyleHook)
	}
}

// One returns a single deviceStyle record from the query.
func (q deviceStyleQuery) One(ctx context.Context, exec boil.ContextExecutor) (*DeviceStyle, error) {
	o := &DeviceStyle{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for device_styles")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all DeviceStyle records from the query.
func (q deviceStyleQuery) All(ctx context.Context, exec boil.ContextExecutor) (DeviceStyleSlice, error) {
	var o []*DeviceStyle

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to DeviceStyle slice")
	}

	if len(deviceStyleAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all DeviceStyle records in the query.
func (q deviceStyleQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count device_styles rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q deviceStyleQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if device_styles exists")
	}

	return count > 0, nil
}

// DeviceDefinition pointed to by the foreign key.
func (o *DeviceStyle) DeviceDefinition(mods ...qm.QueryMod) deviceDefinitionQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.DeviceDefinitionID),
	}

	queryMods = append(queryMods, mods...)

	return DeviceDefinitions(queryMods...)
}

// LoadDeviceDefinition allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (deviceStyleL) LoadDeviceDefinition(ctx context.Context, e boil.ContextExecutor, singular bool, maybeDeviceStyle interface{}, mods queries.Applicator) error {
	var slice []*DeviceStyle
	var object *DeviceStyle

	if singular {
		var ok bool
		object, ok = maybeDeviceStyle.(*DeviceStyle)
		if !ok {
			object = new(DeviceStyle)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeDeviceStyle)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeDeviceStyle))
			}
		}
	} else {
		s, ok := maybeDeviceStyle.(*[]*DeviceStyle)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeDeviceStyle)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeDeviceStyle))
			}
		}
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &deviceStyleR{}
		}
		args = append(args, object.DeviceDefinitionID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &deviceStyleR{}
			}

			for _, a := range args {
				if a == obj.DeviceDefinitionID {
					continue Outer
				}
			}

			args = append(args, obj.DeviceDefinitionID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`device_definitions_api.device_definitions`),
		qm.WhereIn(`device_definitions_api.device_definitions.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load DeviceDefinition")
	}

	var resultSlice []*DeviceDefinition
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice DeviceDefinition")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for device_definitions")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for device_definitions")
	}

	if len(deviceStyleAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.DeviceDefinition = foreign
		if foreign.R == nil {
			foreign.R = &deviceDefinitionR{}
		}
		foreign.R.DeviceStyles = append(foreign.R.DeviceStyles, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.DeviceDefinitionID == foreign.ID {
				local.R.DeviceDefinition = foreign
				if foreign.R == nil {
					foreign.R = &deviceDefinitionR{}
				}
				foreign.R.DeviceStyles = append(foreign.R.DeviceStyles, local)
				break
			}
		}
	}

	return nil
}

// SetDeviceDefinition of the deviceStyle to the related item.
// Sets o.R.DeviceDefinition to related.
// Adds o to related.R.DeviceStyles.
func (o *DeviceStyle) SetDeviceDefinition(ctx context.Context, exec boil.ContextExecutor, insert bool, related *DeviceDefinition) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"device_definitions_api\".\"device_styles\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"device_definition_id"}),
		strmangle.WhereClause("\"", "\"", 2, deviceStylePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.DeviceDefinitionID = related.ID
	if o.R == nil {
		o.R = &deviceStyleR{
			DeviceDefinition: related,
		}
	} else {
		o.R.DeviceDefinition = related
	}

	if related.R == nil {
		related.R = &deviceDefinitionR{
			DeviceStyles: DeviceStyleSlice{o},
		}
	} else {
		related.R.DeviceStyles = append(related.R.DeviceStyles, o)
	}

	return nil
}

// DeviceStyles retrieves all the records using an executor.
func DeviceStyles(mods ...qm.QueryMod) deviceStyleQuery {
	mods = append(mods, qm.From("\"device_definitions_api\".\"device_styles\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"device_definitions_api\".\"device_styles\".*"})
	}

	return deviceStyleQuery{q}
}

// FindDeviceStyle retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindDeviceStyle(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*DeviceStyle, error) {
	deviceStyleObj := &DeviceStyle{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"device_definitions_api\".\"device_styles\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, deviceStyleObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from device_styles")
	}

	if err = deviceStyleObj.doAfterSelectHooks(ctx, exec); err != nil {
		return deviceStyleObj, err
	}

	return deviceStyleObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *DeviceStyle) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no device_styles provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(deviceStyleColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	deviceStyleInsertCacheMut.RLock()
	cache, cached := deviceStyleInsertCache[key]
	deviceStyleInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			deviceStyleAllColumns,
			deviceStyleColumnsWithDefault,
			deviceStyleColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(deviceStyleType, deviceStyleMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(deviceStyleType, deviceStyleMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"device_definitions_api\".\"device_styles\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"device_definitions_api\".\"device_styles\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into device_styles")
	}

	if !cached {
		deviceStyleInsertCacheMut.Lock()
		deviceStyleInsertCache[key] = cache
		deviceStyleInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the DeviceStyle.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *DeviceStyle) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	deviceStyleUpdateCacheMut.RLock()
	cache, cached := deviceStyleUpdateCache[key]
	deviceStyleUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			deviceStyleAllColumns,
			deviceStylePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update device_styles, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"device_definitions_api\".\"device_styles\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, deviceStylePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(deviceStyleType, deviceStyleMapping, append(wl, deviceStylePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update device_styles row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for device_styles")
	}

	if !cached {
		deviceStyleUpdateCacheMut.Lock()
		deviceStyleUpdateCache[key] = cache
		deviceStyleUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q deviceStyleQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for device_styles")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for device_styles")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o DeviceStyleSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceStylePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"device_definitions_api\".\"device_styles\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, deviceStylePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in deviceStyle slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all deviceStyle")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *DeviceStyle) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no device_styles provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(deviceStyleColumnsWithDefault, o)

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

	deviceStyleUpsertCacheMut.RLock()
	cache, cached := deviceStyleUpsertCache[key]
	deviceStyleUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			deviceStyleAllColumns,
			deviceStyleColumnsWithDefault,
			deviceStyleColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			deviceStyleAllColumns,
			deviceStylePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert device_styles, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(deviceStylePrimaryKeyColumns))
			copy(conflict, deviceStylePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"device_definitions_api\".\"device_styles\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(deviceStyleType, deviceStyleMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(deviceStyleType, deviceStyleMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert device_styles")
	}

	if !cached {
		deviceStyleUpsertCacheMut.Lock()
		deviceStyleUpsertCache[key] = cache
		deviceStyleUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single DeviceStyle record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *DeviceStyle) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no DeviceStyle provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), deviceStylePrimaryKeyMapping)
	sql := "DELETE FROM \"device_definitions_api\".\"device_styles\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from device_styles")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for device_styles")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q deviceStyleQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no deviceStyleQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from device_styles")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for device_styles")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o DeviceStyleSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(deviceStyleBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceStylePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"device_definitions_api\".\"device_styles\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, deviceStylePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from deviceStyle slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for device_styles")
	}

	if len(deviceStyleAfterDeleteHooks) != 0 {
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
func (o *DeviceStyle) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindDeviceStyle(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *DeviceStyleSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := DeviceStyleSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceStylePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"device_definitions_api\".\"device_styles\".* FROM \"device_definitions_api\".\"device_styles\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, deviceStylePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in DeviceStyleSlice")
	}

	*o = slice

	return nil
}

// DeviceStyleExists checks if the DeviceStyle row exists.
func DeviceStyleExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"device_definitions_api\".\"device_styles\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if device_styles exists")
	}

	return exists, nil
}
