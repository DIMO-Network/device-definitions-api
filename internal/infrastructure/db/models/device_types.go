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

// DeviceType is an object representing the database table.
type DeviceType struct {
	ID         string    `boil:"id" json:"id" toml:"id" yaml:"id"`
	CreatedAt  time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt  time.Time `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	Name       string    `boil:"name" json:"name" toml:"name" yaml:"name"`
	Properties null.JSON `boil:"properties" json:"properties,omitempty" toml:"properties" yaml:"properties,omitempty"`

	R *deviceTypeR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L deviceTypeL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var DeviceTypeColumns = struct {
	ID         string
	CreatedAt  string
	UpdatedAt  string
	Name       string
	Properties string
}{
	ID:         "id",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	Name:       "name",
	Properties: "properties",
}

var DeviceTypeTableColumns = struct {
	ID         string
	CreatedAt  string
	UpdatedAt  string
	Name       string
	Properties string
}{
	ID:         "device_types.id",
	CreatedAt:  "device_types.created_at",
	UpdatedAt:  "device_types.updated_at",
	Name:       "device_types.name",
	Properties: "device_types.properties",
}

// Generated where

var DeviceTypeWhere = struct {
	ID         whereHelperstring
	CreatedAt  whereHelpertime_Time
	UpdatedAt  whereHelpertime_Time
	Name       whereHelperstring
	Properties whereHelpernull_JSON
}{
	ID:         whereHelperstring{field: "\"device_definitions_api\".\"device_types\".\"id\""},
	CreatedAt:  whereHelpertime_Time{field: "\"device_definitions_api\".\"device_types\".\"created_at\""},
	UpdatedAt:  whereHelpertime_Time{field: "\"device_definitions_api\".\"device_types\".\"updated_at\""},
	Name:       whereHelperstring{field: "\"device_definitions_api\".\"device_types\".\"name\""},
	Properties: whereHelpernull_JSON{field: "\"device_definitions_api\".\"device_types\".\"properties\""},
}

// DeviceTypeRels is where relationship names are stored.
var DeviceTypeRels = struct {
	DeviceDefinitions string
}{
	DeviceDefinitions: "DeviceDefinitions",
}

// deviceTypeR is where relationships are stored.
type deviceTypeR struct {
	DeviceDefinitions DeviceDefinitionSlice `boil:"DeviceDefinitions" json:"DeviceDefinitions" toml:"DeviceDefinitions" yaml:"DeviceDefinitions"`
}

// NewStruct creates a new relationship struct
func (*deviceTypeR) NewStruct() *deviceTypeR {
	return &deviceTypeR{}
}

func (r *deviceTypeR) GetDeviceDefinitions() DeviceDefinitionSlice {
	if r == nil {
		return nil
	}
	return r.DeviceDefinitions
}

// deviceTypeL is where Load methods for each relationship are stored.
type deviceTypeL struct{}

var (
	deviceTypeAllColumns            = []string{"id", "created_at", "updated_at", "name", "properties"}
	deviceTypeColumnsWithoutDefault = []string{"id", "name"}
	deviceTypeColumnsWithDefault    = []string{"created_at", "updated_at", "properties"}
	deviceTypePrimaryKeyColumns     = []string{"id"}
	deviceTypeGeneratedColumns      = []string{}
)

type (
	// DeviceTypeSlice is an alias for a slice of pointers to DeviceType.
	// This should almost always be used instead of []DeviceType.
	DeviceTypeSlice []*DeviceType
	// DeviceTypeHook is the signature for custom DeviceType hook methods
	DeviceTypeHook func(context.Context, boil.ContextExecutor, *DeviceType) error

	deviceTypeQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	deviceTypeType                 = reflect.TypeOf(&DeviceType{})
	deviceTypeMapping              = queries.MakeStructMapping(deviceTypeType)
	deviceTypePrimaryKeyMapping, _ = queries.BindMapping(deviceTypeType, deviceTypeMapping, deviceTypePrimaryKeyColumns)
	deviceTypeInsertCacheMut       sync.RWMutex
	deviceTypeInsertCache          = make(map[string]insertCache)
	deviceTypeUpdateCacheMut       sync.RWMutex
	deviceTypeUpdateCache          = make(map[string]updateCache)
	deviceTypeUpsertCacheMut       sync.RWMutex
	deviceTypeUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var deviceTypeAfterSelectHooks []DeviceTypeHook

var deviceTypeBeforeInsertHooks []DeviceTypeHook
var deviceTypeAfterInsertHooks []DeviceTypeHook

var deviceTypeBeforeUpdateHooks []DeviceTypeHook
var deviceTypeAfterUpdateHooks []DeviceTypeHook

var deviceTypeBeforeDeleteHooks []DeviceTypeHook
var deviceTypeAfterDeleteHooks []DeviceTypeHook

var deviceTypeBeforeUpsertHooks []DeviceTypeHook
var deviceTypeAfterUpsertHooks []DeviceTypeHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *DeviceType) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceTypeAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *DeviceType) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceTypeBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *DeviceType) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceTypeAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *DeviceType) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceTypeBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *DeviceType) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceTypeAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *DeviceType) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceTypeBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *DeviceType) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceTypeAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *DeviceType) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceTypeBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *DeviceType) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceTypeAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddDeviceTypeHook registers your hook function for all future operations.
func AddDeviceTypeHook(hookPoint boil.HookPoint, deviceTypeHook DeviceTypeHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		deviceTypeAfterSelectHooks = append(deviceTypeAfterSelectHooks, deviceTypeHook)
	case boil.BeforeInsertHook:
		deviceTypeBeforeInsertHooks = append(deviceTypeBeforeInsertHooks, deviceTypeHook)
	case boil.AfterInsertHook:
		deviceTypeAfterInsertHooks = append(deviceTypeAfterInsertHooks, deviceTypeHook)
	case boil.BeforeUpdateHook:
		deviceTypeBeforeUpdateHooks = append(deviceTypeBeforeUpdateHooks, deviceTypeHook)
	case boil.AfterUpdateHook:
		deviceTypeAfterUpdateHooks = append(deviceTypeAfterUpdateHooks, deviceTypeHook)
	case boil.BeforeDeleteHook:
		deviceTypeBeforeDeleteHooks = append(deviceTypeBeforeDeleteHooks, deviceTypeHook)
	case boil.AfterDeleteHook:
		deviceTypeAfterDeleteHooks = append(deviceTypeAfterDeleteHooks, deviceTypeHook)
	case boil.BeforeUpsertHook:
		deviceTypeBeforeUpsertHooks = append(deviceTypeBeforeUpsertHooks, deviceTypeHook)
	case boil.AfterUpsertHook:
		deviceTypeAfterUpsertHooks = append(deviceTypeAfterUpsertHooks, deviceTypeHook)
	}
}

// One returns a single deviceType record from the query.
func (q deviceTypeQuery) One(ctx context.Context, exec boil.ContextExecutor) (*DeviceType, error) {
	o := &DeviceType{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for device_types")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all DeviceType records from the query.
func (q deviceTypeQuery) All(ctx context.Context, exec boil.ContextExecutor) (DeviceTypeSlice, error) {
	var o []*DeviceType

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to DeviceType slice")
	}

	if len(deviceTypeAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all DeviceType records in the query.
func (q deviceTypeQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count device_types rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q deviceTypeQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if device_types exists")
	}

	return count > 0, nil
}

// DeviceDefinitions retrieves all the device_definition's DeviceDefinitions with an executor.
func (o *DeviceType) DeviceDefinitions(mods ...qm.QueryMod) deviceDefinitionQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"device_definitions_api\".\"device_definitions\".\"device_type_id\"=?", o.ID),
	)

	return DeviceDefinitions(queryMods...)
}

// LoadDeviceDefinitions allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (deviceTypeL) LoadDeviceDefinitions(ctx context.Context, e boil.ContextExecutor, singular bool, maybeDeviceType interface{}, mods queries.Applicator) error {
	var slice []*DeviceType
	var object *DeviceType

	if singular {
		var ok bool
		object, ok = maybeDeviceType.(*DeviceType)
		if !ok {
			object = new(DeviceType)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeDeviceType)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeDeviceType))
			}
		}
	} else {
		s, ok := maybeDeviceType.(*[]*DeviceType)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeDeviceType)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeDeviceType))
			}
		}
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &deviceTypeR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &deviceTypeR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.ID) {
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
		qm.From(`device_definitions_api.device_definitions`),
		qm.WhereIn(`device_definitions_api.device_definitions.device_type_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load device_definitions")
	}

	var resultSlice []*DeviceDefinition
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice device_definitions")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on device_definitions")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for device_definitions")
	}

	if len(deviceDefinitionAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.DeviceDefinitions = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &deviceDefinitionR{}
			}
			foreign.R.DeviceType = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if queries.Equal(local.ID, foreign.DeviceTypeID) {
				local.R.DeviceDefinitions = append(local.R.DeviceDefinitions, foreign)
				if foreign.R == nil {
					foreign.R = &deviceDefinitionR{}
				}
				foreign.R.DeviceType = local
				break
			}
		}
	}

	return nil
}

// AddDeviceDefinitions adds the given related objects to the existing relationships
// of the device_type, optionally inserting them as new records.
// Appends related to o.R.DeviceDefinitions.
// Sets related.R.DeviceType appropriately.
func (o *DeviceType) AddDeviceDefinitions(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*DeviceDefinition) error {
	var err error
	for _, rel := range related {
		if insert {
			queries.Assign(&rel.DeviceTypeID, o.ID)
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"device_definitions_api\".\"device_definitions\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"device_type_id"}),
				strmangle.WhereClause("\"", "\"", 2, deviceDefinitionPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.IsDebug(ctx) {
				writer := boil.DebugWriterFrom(ctx)
				fmt.Fprintln(writer, updateQuery)
				fmt.Fprintln(writer, values)
			}
			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			queries.Assign(&rel.DeviceTypeID, o.ID)
		}
	}

	if o.R == nil {
		o.R = &deviceTypeR{
			DeviceDefinitions: related,
		}
	} else {
		o.R.DeviceDefinitions = append(o.R.DeviceDefinitions, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &deviceDefinitionR{
				DeviceType: o,
			}
		} else {
			rel.R.DeviceType = o
		}
	}
	return nil
}

// SetDeviceDefinitions removes all previously related items of the
// device_type replacing them completely with the passed
// in related items, optionally inserting them as new records.
// Sets o.R.DeviceType's DeviceDefinitions accordingly.
// Replaces o.R.DeviceDefinitions with related.
// Sets related.R.DeviceType's DeviceDefinitions accordingly.
func (o *DeviceType) SetDeviceDefinitions(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*DeviceDefinition) error {
	query := "update \"device_definitions_api\".\"device_definitions\" set \"device_type_id\" = null where \"device_type_id\" = $1"
	values := []interface{}{o.ID}
	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, query)
		fmt.Fprintln(writer, values)
	}
	_, err := exec.ExecContext(ctx, query, values...)
	if err != nil {
		return errors.Wrap(err, "failed to remove relationships before set")
	}

	if o.R != nil {
		for _, rel := range o.R.DeviceDefinitions {
			queries.SetScanner(&rel.DeviceTypeID, nil)
			if rel.R == nil {
				continue
			}

			rel.R.DeviceType = nil
		}
		o.R.DeviceDefinitions = nil
	}

	return o.AddDeviceDefinitions(ctx, exec, insert, related...)
}

// RemoveDeviceDefinitions relationships from objects passed in.
// Removes related items from R.DeviceDefinitions (uses pointer comparison, removal does not keep order)
// Sets related.R.DeviceType.
func (o *DeviceType) RemoveDeviceDefinitions(ctx context.Context, exec boil.ContextExecutor, related ...*DeviceDefinition) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	for _, rel := range related {
		queries.SetScanner(&rel.DeviceTypeID, nil)
		if rel.R != nil {
			rel.R.DeviceType = nil
		}
		if _, err = rel.Update(ctx, exec, boil.Whitelist("device_type_id")); err != nil {
			return err
		}
	}
	if o.R == nil {
		return nil
	}

	for _, rel := range related {
		for i, ri := range o.R.DeviceDefinitions {
			if rel != ri {
				continue
			}

			ln := len(o.R.DeviceDefinitions)
			if ln > 1 && i < ln-1 {
				o.R.DeviceDefinitions[i] = o.R.DeviceDefinitions[ln-1]
			}
			o.R.DeviceDefinitions = o.R.DeviceDefinitions[:ln-1]
			break
		}
	}

	return nil
}

// DeviceTypes retrieves all the records using an executor.
func DeviceTypes(mods ...qm.QueryMod) deviceTypeQuery {
	mods = append(mods, qm.From("\"device_definitions_api\".\"device_types\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"device_definitions_api\".\"device_types\".*"})
	}

	return deviceTypeQuery{q}
}

// FindDeviceType retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindDeviceType(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*DeviceType, error) {
	deviceTypeObj := &DeviceType{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"device_definitions_api\".\"device_types\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, deviceTypeObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from device_types")
	}

	if err = deviceTypeObj.doAfterSelectHooks(ctx, exec); err != nil {
		return deviceTypeObj, err
	}

	return deviceTypeObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *DeviceType) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no device_types provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(deviceTypeColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	deviceTypeInsertCacheMut.RLock()
	cache, cached := deviceTypeInsertCache[key]
	deviceTypeInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			deviceTypeAllColumns,
			deviceTypeColumnsWithDefault,
			deviceTypeColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(deviceTypeType, deviceTypeMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(deviceTypeType, deviceTypeMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"device_definitions_api\".\"device_types\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"device_definitions_api\".\"device_types\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into device_types")
	}

	if !cached {
		deviceTypeInsertCacheMut.Lock()
		deviceTypeInsertCache[key] = cache
		deviceTypeInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the DeviceType.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *DeviceType) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	deviceTypeUpdateCacheMut.RLock()
	cache, cached := deviceTypeUpdateCache[key]
	deviceTypeUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			deviceTypeAllColumns,
			deviceTypePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update device_types, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"device_definitions_api\".\"device_types\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, deviceTypePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(deviceTypeType, deviceTypeMapping, append(wl, deviceTypePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update device_types row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for device_types")
	}

	if !cached {
		deviceTypeUpdateCacheMut.Lock()
		deviceTypeUpdateCache[key] = cache
		deviceTypeUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q deviceTypeQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for device_types")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for device_types")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o DeviceTypeSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceTypePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"device_definitions_api\".\"device_types\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, deviceTypePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in deviceType slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all deviceType")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *DeviceType) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no device_types provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(deviceTypeColumnsWithDefault, o)

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

	deviceTypeUpsertCacheMut.RLock()
	cache, cached := deviceTypeUpsertCache[key]
	deviceTypeUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			deviceTypeAllColumns,
			deviceTypeColumnsWithDefault,
			deviceTypeColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			deviceTypeAllColumns,
			deviceTypePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert device_types, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(deviceTypePrimaryKeyColumns))
			copy(conflict, deviceTypePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"device_definitions_api\".\"device_types\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(deviceTypeType, deviceTypeMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(deviceTypeType, deviceTypeMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert device_types")
	}

	if !cached {
		deviceTypeUpsertCacheMut.Lock()
		deviceTypeUpsertCache[key] = cache
		deviceTypeUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single DeviceType record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *DeviceType) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no DeviceType provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), deviceTypePrimaryKeyMapping)
	sql := "DELETE FROM \"device_definitions_api\".\"device_types\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from device_types")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for device_types")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q deviceTypeQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no deviceTypeQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from device_types")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for device_types")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o DeviceTypeSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(deviceTypeBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceTypePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"device_definitions_api\".\"device_types\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, deviceTypePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from deviceType slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for device_types")
	}

	if len(deviceTypeAfterDeleteHooks) != 0 {
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
func (o *DeviceType) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindDeviceType(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *DeviceTypeSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := DeviceTypeSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceTypePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"device_definitions_api\".\"device_types\".* FROM \"device_definitions_api\".\"device_types\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, deviceTypePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in DeviceTypeSlice")
	}

	*o = slice

	return nil
}

// DeviceTypeExists checks if the DeviceType row exists.
func DeviceTypeExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"device_definitions_api\".\"device_types\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if device_types exists")
	}

	return exists, nil
}
