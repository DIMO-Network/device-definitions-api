//nolint:tagliatelle
package queries

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/DIMO-Network/shared/pkg/logfields"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/pkg/db"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"
	"github.com/DIMO-Network/shared/pkg/vin"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

type DecodeVINQueryHandler struct {
	dbs                            func() *db.ReaderWriter
	vinDecodingService             services.VINDecodingService
	logger                         *zerolog.Logger
	vinRepository                  repositories.VINRepository
	fuelAPIService                 gateways.FuelAPIService
	powerTrainTypeService          services.PowerTrainTypeService
	deviceDefinitionCatalogService gateways.DeviceDefinitionCatalogService
	identity                       gateways.IdentityAPI
}

type DecodeVINQuery struct {
	VIN        string `json:"vin"`
	KnownModel string `json:"knownModel"`
	KnownYear  int32  `json:"knownYear"`
	Country    string `json:"country"`
}

func (*DecodeVINQuery) Key() string { return "DecodeVINQuery" }

func NewDecodeVINQueryHandler(dbs func() *db.ReaderWriter, vinDecodingService services.VINDecodingService,
	vinRepository repositories.VINRepository,
	logger *zerolog.Logger,
	fuelAPIService gateways.FuelAPIService,
	powerTrainTypeService services.PowerTrainTypeService,
	deviceDefinitionCatalogService gateways.DeviceDefinitionCatalogService,
	identity gateways.IdentityAPI) DecodeVINQueryHandler {
	return DecodeVINQueryHandler{
		dbs:                            dbs,
		vinDecodingService:             vinDecodingService,
		logger:                         logger,
		vinRepository:                  vinRepository,
		fuelAPIService:                 fuelAPIService,
		powerTrainTypeService:          powerTrainTypeService,
		deviceDefinitionCatalogService: deviceDefinitionCatalogService,
		identity:                       identity,
	}
}

func (dc DecodeVINQueryHandler) Handle(ctx context.Context, query *DecodeVINQuery) (*p_grpc.DecodeVinResponse, error) {
	if query == nil {
		return nil, &exceptions.ValidationError{Err: errors.New("query is nil")}
	}
	if len(query.VIN) < 10 || len(query.VIN) > 17 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid VIN %s", query.VIN)}
	}
	resp := &p_grpc.DecodeVinResponse{}
	vinObj := vin.VIN(query.VIN)

	if !vinObj.IsValidJapanChassis() && !vinObj.IsValidVIN() {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid VIN %s", query.VIN)}
	}

	resp.Year = int32(vinObj.Year())
	wmi := vinObj.Wmi()

	localLog := dc.logger.With().
		Str(logfields.VIN, vinObj.String()).
		Str(logfields.FunctionName, query.Key()).
		Str("vinYear", fmt.Sprintf("%d", resp.Year)).
		Str("knownModel", query.KnownModel).
		Str("knownYear", strconv.Itoa(int(query.KnownYear))).
		Str(logfields.CountryCode, query.Country).
		Logger()

	const (
		VinRequests = "VIN_All_Request"
		VinSuccess  = "VIN_Success_Request"
		VinExists   = "VIN_Exists_Request"
		VinErrors   = "VIN_Error_Request"
	)

	metrics.Success.With(prometheus.Labels{"method": VinRequests}).Inc()
	// The transaction opens and closes inside this call. Everything below it
	// is slow -- hydrateResponseFromVinNumber alone makes a catalog request
	// and a reader query -- and none of it may run while a writer connection
	// and a serializable snapshot are held.
	vinDecodeNumber, err := dc.readCachedVinNumber(ctx, vinObj.String())
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		return nil, err
	}
	// if database vin_number match found, just return it here
	cached, errCached := dc.hydrateResponseFromVinNumber(ctx, vinDecodeNumber)
	if errCached != nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		return nil, errCached
	}
	if cached != nil {
		metrics.Success.With(prometheus.Labels{"method": VinExists}).Inc()
		return cached, nil
	}
	// check if vin has failed in the past, and if it has just fail now
	vinAlreadyFailed, _ := models.FailedVinDecodes(models.FailedVinDecodeWhere.Vin.EQ(vinObj.String())).Exists(ctx, dc.dbs().Writer)
	if vinAlreadyFailed {
		return nil, fmt.Errorf("vin %s failed decoding already", vinObj.String())
	}

	localLog.Info().Msgf("Start Decode VIN ")

	_, err = models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(common.DefaultDeviceType)).One(ctx, dc.dbs().Reader)
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		return nil, errors.Wrap(err, "failed to get device_type")
	}
	// future: see if we can self decode model based on data we have before calling external decode WMI and VDS. Only thing is we won't get the style.

	if resp.Year == 0 || resp.Year > int32(time.Now().Year()+1) {
		localLog.Info().Msgf("encountered vinObj with non-standard year digit")
	}
	// check if this is a Tesla VIN, if not just follow regular path
	vinInfo := &coremodels.VINDecodingInfoData{}
	vinExtra := &coremodels.VINDecodingVendorExtra{}
	dbWMI, err := models.Wmis(models.WmiWhere.Wmi.EQ(wmi)).One(ctx, dc.dbs().Reader)
	if err == nil && dbWMI != nil {
		if dbWMI.ManufacturerName == "Tesla" {
			vinInfo, vinExtra, err = dc.vinDecodingService.GetVIN(ctx, vinObj.String(), coremodels.TeslaProvider, query.Country)
			resp.Manufacturer = "Tesla"
		}
	}
	// not a tesla, regular decode path
	if vinInfo == nil || vinInfo.Model == "" {
		vinInfo, vinExtra, err = dc.vinDecodingService.GetVIN(ctx, vinObj.String(), coremodels.AllProviders, query.Country) // this will try drivly first unless of japan
	}
	// GetVIN overwrites the initialiser above, and reports no vendor extra at
	// all on the paths that fail before any vendor is tried: the invalid-VIN
	// guard, and the 0SC test-VIN branch, which now reads a template from the
	// catalog and hands back that read's error. The failure branch below reads
	// six fields off this pointer, so a nil here panicked the decode instead of
	// recording the failure -- and a decode that records nothing repeats the
	// whole vendor fan-out, and the same panic, on every retry forever.
	if vinExtra == nil {
		vinExtra = &coremodels.VINDecodingVendorExtra{}
	}

	// if no luck decoding VIN, try buildingVinInfo from known data passed in, typically smartcar or software connections
	if err != nil {
		if len(query.KnownModel) > 0 && query.KnownYear > 0 {
			// note if this is successful, err gets set to nil
			// todo: the knownModel should correspond with the Make
			vinInfo, err = dc.vinInfoFromKnown(ctx, vinObj, query.KnownModel, query.KnownYear)
		}
	}

	if vinInfo != nil {
		localLog = localLog.With().Str("decode_source", string(vinInfo.Source)).Logger()
	}

	if err != nil || vinInfo == nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		if err == nil {
			err = errors.New("failed to decode, vinInfo is nil")
		}
		localLog.Err(err).Msgf("failed to decode vinObj from provider, country: %s", query.Country)

		failedVinDecode := models.FailedVinDecode{
			Vin:              vinObj.String(),
			VendorsTried:     vinExtra.VendorsTried,
			VincarioData:     null.JSONFrom(vinExtra.VincarioRaw),
			DrivlyData:       null.JSONFrom(vinExtra.DrivlyRaw),
			AutoisoData:      null.JSONFrom(vinExtra.AutoIsoRaw),
			DatgroupData:     null.JSONFrom(vinExtra.DATGroupRaw),
			Vin17Data:        null.JSONFrom(vinExtra.Japan17VINRaw),
			CountryCode:      null.StringFrom(query.Country),
			ManufacturerName: null.StringFrom(resp.Manufacturer),
		}
		errFailedVin := failedVinDecode.Insert(ctx, dc.dbs().Writer, boil.Infer())
		if errFailedVin != nil {
			localLog.Err(errFailedVin).Msgf("failed to save failed vin decode to database")
		}

		return nil, err
	}
	// WMI's may be re-used by multiple OEM's of same parent OEM, but just create it if needed
	if dbWMI == nil {
		_, err = dc.vinRepository.GetOrCreateWMI(ctx, wmi, vinInfo.Make)
		if err != nil {
			// just log, Japan chasis numbers won't really work with this anyways
			dc.logger.Error().Err(err).Msgf("failed to get or create wmi for vinObj %s", vinObj.String())
		}
	}
	resp.Manufacturer = vinInfo.Make
	resp.Source = string(vinInfo.Source)
	resp.Year = vinInfo.Year
	resp.Model = vinInfo.Model

	tid, errID := definitionIDForDecode(vinInfo.Make, vinInfo.Model, int16(vinInfo.Year))
	if errID != nil {
		// Nothing below this can succeed: the catalog cannot hold a template
		// at this id, so the read 404s and the create 422s -- identically, on
		// every retry, for every VIN of this model-year. Stop here, with the
		// same answer the low-confidence gate gives, so the client falls back
		// to the manual make/model/year picker instead of being told the
		// service is broken.
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		localLog.Warn().Err(errID).Str("decode_source", string(vinInfo.Source)).
			Msg("decoded model yields an id definitions-worker can never accept; returning not found so the client opens the manual picker")
		return nil, errID
	}
	resp.DefinitionId = tid

	tblDef, _, errTbl := dc.deviceDefinitionCatalogService.GetTemplateByID(ctx, tid)
	if errTbl != nil && !errors.Is(errTbl, gateways.ErrTemplateNotFound) {
		// A catalog outage (5xx, timeout, decode failure) is not the same as
		// the definition not existing. Falling through here would read
		// tblDef as nil and run Create() below -- writing a duplicate
		// definition for a vehicle that may already exist, on the VIN-decode
		// hot path, at decode volume, during the worst possible moment for
		// it. Abort the decode instead of continuing with a nil template.
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		return nil, errors.Wrapf(errTbl, "failed to get definition from catalog for vinObj: %s, id: %s", vinObj.String(), tid)
	}
	if errTbl != nil {
		// Genuinely not found (ErrTemplateNotFound): fall through and create it below.
		dc.logger.Warn().Err(errTbl).Msgf("failed to get definition from catalog for vinObj: %s, id: %s", vinObj.String(), tid)
	} else if tblDef == nil {
		dc.logger.Warn().Msgf("failed to get definition from catalog for vinObj: %s, id: %s", vinObj.String(), tid)
	} else {
		dc.logger.Info().Str(logfields.VIN, vinObj.String()).Msgf("found definition from catalog %s: %+v", tid, tblDef)
	}

	// add images if we don't have any for this definition_id
	images, _ := models.Images(models.ImageWhere.DefinitionID.EQ(resp.DefinitionId)).All(ctx, dc.dbs().Reader)
	localLog.Debug().Msgf("Current Images : %d", len(images))

	if len(images) == 0 {
		err = dc.associateImagesToDeviceDefinition(ctx, resp.DefinitionId, vinInfo.Make, vinInfo.Model, int(resp.Year), 2, 2)
		if err != nil {
			localLog.Err(err).Send()
		}

		err = dc.associateImagesToDeviceDefinition(ctx, resp.DefinitionId, vinInfo.Make, vinInfo.Model, int(resp.Year), 2, 6)
		if err != nil {
			localLog.Err(err).Send()
		}
	}

	// figure out powertrain for the style write further down. This is the
	// old heuristic derivation, kept alive because processDeviceStyle stamps
	// its result onto device_styles -- the table the extraction pipeline
	// reads to build templates in the first place. Changing what gets
	// written there is a separate decision; pt is passed to
	// processDeviceStyle explicitly below so this stays true regardless of
	// what resp.Powertrain ends up holding for the response.
	pt := dc.powerTrainTypeService.ResolvePowerTrainFromVinInfo(vinInfo.StyleName, vinInfo.FuelType)
	if pt == "" {
		// try a different way
		pt, _ = dc.powerTrainTypeService.ResolvePowerTrainType(stringutils.SlugString(resp.Manufacturer), stringutils.SlugString(resp.Model), null.JSON{}, null.JSON{})
	}
	if pt != "" {
		resp.Powertrain = pt
	}

	// Not in the catalog yet. A low-confidence decoder is not trusted to name a
	// new template: the Japanese chassis decoders have returned body-style codes
	// such as "4D" as the model, which became garbage definitions like
	// toyota_4d_2017. Answer not found instead, so the client falls back to the
	// manual make/model/year picker.
	if tblDef == nil && isLowConfidenceSource(vinInfo.Source) {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		localLog.Warn().Str("decode_source", string(vinInfo.Source)).Msg("low-confidence decode and catalog miss; returning not found so the client opens the manual picker")
		return nil, &exceptions.NotFoundError{Err: fmt.Errorf("device definition %s is not in the catalog and decode source %s is low-confidence; manual selection required", tid, vinInfo.Source)}
	}

	// Every other source creates it. Create is create-only: if another writer
	// stored the template after the read above, the decode continues with their
	// template instead of overwriting it.
	if tblDef == nil {
		// if any images were added above, they will be in the database
		latestImages, _ := models.Images(models.ImageWhere.DefinitionID.EQ(resp.DefinitionId)).All(ctx, dc.dbs().Reader)
		// todo load up some metadata from what was decoded. Powertrain too
		md := resolveMetadataFromInfo(resp.Powertrain, vinInfo)

		// The id comes back in DefinitionId; NewTrxHash stays empty because
		// definitions are no longer written on-chain and there is no
		// transaction. Putting the slug here would hand a "0x..." consumer a
		// value that is not a hash.
		tblDef, err = createOrAdoptTemplate(ctx, dc.deviceDefinitionCatalogService, resp.Manufacturer, coremodels.DeviceDefinitionTablelandModel{
			ID:         tid,
			KSUID:      ksuid.New().String(),
			Model:      resp.Model,
			Year:       int(resp.Year),
			DeviceType: common.DefaultDeviceType,
			ImageURI:   common.GetDefaultImageURL(latestImages),
			Metadata:   md,
		})
		if err != nil {
			metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
			return nil, errors.Wrap(err, "error creating new device definition from decoded vinObj")
		}
	}

	if tblDef != nil {
		resp.DefinitionId = tblDef.ID

		// Narrow the template to the trim this VIN decoded to. ManufacturerCode
		// only ever arrives via drivly (VINDecodingInfoData.ManufacturerCode,
		// populated in vin_decoding_service.go's buildFromDrivly from
		// DrivlyVINResponse.ManufacturerCode); every other provider leaves it
		// empty, so a manufacturerCode-keyed selector simply can't match for
		// those decodes -- not an error, just a signal that isn't there.
		resolved := services.MatchTrim(tblDef, services.MatchSignals{
			ManufacturerCode: vinInfo.ManufacturerCode,
			StyleName:        vinInfo.StyleName,
			VIN:              vinObj.String(),
		})
		resp.Trim = resolved.Trim
		resp.TemplateVersion = int32(resolved.TemplateVersion)
		resp.MatchQuality = string(resolved.Quality)
		resp.MatchCandidates = resolved.Candidates
		// MatchBy and HardwareTemplateId are resolved by MatchTrim and were
		// previously dropped here. match.by is how a consumer -- and our own
		// dashboards -- can tell whether trim matching is firing on
		// manufacturerCode or on styleName, and hardwareTemplateId is the
		// template-default-with-trim-override resolution the contract
		// requires be settled here "so callers never reimplement the
		// fallback". Computing both and emitting neither left the only
		// evidence of that work inside the matcher's own unit tests.
		resp.MatchBy = resolved.MatchedBy
		resp.HardwareTemplateId = resolved.HardwareTemplateID

		observeTrimMatch(resolved, resp.Source)

		// The response's powertrain comes from the resolved template/trim
		// attributes now, not from the pt heuristic above -- that heuristic
		// is what produced the ICE/hybrid-attributes mismatch this migration
		// exists to fix. If the template carries no powertrain_type at all
		// (template or matched trim), the response reports none rather than
		// falling back to a guess.
		resp.Powertrain = ""
		if v, ok := resolved.Attributes[common.PowerTrainType].(string); ok {
			resp.Powertrain = v
		}
	}

	// match style - only process style if name is longer than 1
	if len(vinInfo.StyleName) < 2 {
		localLog.Warn().Msgf("decoded style name too short: %s must have a minimum of 2 characters.", vinInfo.StyleName)
	} else {
		var styleErr error
		// pt, not resp.Powertrain: processDeviceStyle writes to device_styles,
		// the extraction pipeline's input, and that write path is unchanged
		// by this migration -- see the comment on pt's declaration above.
		resp.DeviceStyleId, styleErr = dc.processDeviceStyle(ctx, vinInfo, tid, pt)
		if styleErr != nil {
			dc.logger.Error().Err(styleErr).Msgf("error processing device style for vinObj: %s. continuing", vinObj.String())
		}
	}

	// insert vin_numbers
	errVinNumber := dc.saveVinDecodeNumber(ctx, vinObj, vinInfo, resp)
	if errVinNumber != nil {
		return nil, errors.Wrap(errVinNumber, "error saving vin_number")
	}

	localLog.Info().Str("device_definition_id", resp.DefinitionId).
		Str("style_id", resp.DeviceStyleId).
		// How the trim resolved, on the line that already records a
		// successful decode: without these, "how often does a decode fail to
		// narrow, and on which provider" is unanswerable from logs.
		Str("match_quality", resp.MatchQuality).
		Str("trim", resp.Trim).
		Strs("match_by", resp.MatchBy).
		Str("manufacturer_code", vinInfo.ManufacturerCode).
		Str("wmi", wmi).
		Str("vds", vinObj.VDS()).
		Str("vis", vinObj.VIS()).
		Str("check_digit", vinObj.CheckDigit()).Msgf("decoded vin ok with: %s", vinInfo.Source)

	metrics.Success.With(prometheus.Labels{"method": VinSuccess}).Inc()

	return resp, nil
}

func resolveMetadataFromInfo(powertrain string, _ *coremodels.VINDecodingInfoData) *coremodels.DeviceDefinitionMetadata {
	md := coremodels.DeviceDefinitionMetadata{DeviceAttributes: make([]coremodels.DeviceTypeAttribute, 0)}
	if powertrain != "" {
		md.DeviceAttributes = append(md.DeviceAttributes, coremodels.DeviceTypeAttribute{
			Name:  common.PowerTrainType,
			Value: powertrain,
		})
	}

	return &md
}

// readCachedVinNumber reads the vin_numbers row for a VIN, or nil when this
// VIN has not been decoded before.
//
// The transaction holds nothing but this one read, and it is closed before the
// function returns -- on every path, including a failed query and a failed
// begin. Handle used to open it and keep it open across
// hydrateResponseFromVinNumber, which was pure in-memory when that was
// written. It now issues a catalog GET through a client with a 30 second
// timeout and a device_styles query against the reader, and returns early on a
// cache hit, so the path the code itself calls "the one most decodes take"
// held a writer connection and a SERIALIZABLE snapshot across a full CDN round
// trip -- and took a reader connection while holding it. A two second catalog
// stall pinned every writer connection for two seconds per decode; a thirty
// second stall exhausted the pool while unrelated writes queued behind it.
//
// The isolation level is unchanged: what this row is read at is a separate
// decision from how long the read is held.
func (dc DecodeVINQueryHandler) readCachedVinNumber(ctx context.Context, vinStr string) (*models.VinNumber, error) {
	tx, err := dc.dbs().Writer.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, errors.Wrap(err, "error when beginning transaction")
	}
	// Nothing is written here, so Rollback is how this transaction ends, and
	// deferring it closes the error paths too rather than only the happy one.
	defer tx.Rollback() //nolint:errcheck
	vn, err := models.VinNumbers(models.VinNumberWhere.Vin.EQ(vinStr)).One(ctx, tx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "error when querying for existing VIN number")
	}
	return vn, nil
}

// definitionIDForDecode builds the template id for a decoded make/model/year,
// and refuses one definitions-worker can never hold.
//
// common.DeviceDefinitionSlug repairs an id by dropping characters outside the
// worker's class, and does not re-check what it produced. A model written in a
// script with no character in that class -- ハイエース, Нива, reachable through
// vinInfoFromKnown's KnownModel, whose source is not low-confidence -- comes
// back as toyota__2020, whose empty middle segment ID_RE refuses. Writing that
// id would 422 the create and fail every retry of every VIN of that
// model-year, so the failure belongs here, before the id is used for anything.
//
// The error is *exceptions.NotFoundError, the same type the low-confidence
// gate returns: the remedy is the same one, a manual make/model/year pick, and
// both the HTTP and gRPC layers already translate it (404, codes.NotFound)
// rather than reporting the decoder as broken. It wraps
// common.ErrUnmintableDefinitionID so a caller can tell this apart from a
// vehicle that merely has no template yet.
func definitionIDForDecode(makeName, modelName string, year int16) (string, error) {
	id, err := common.DeviceDefinitionSlug(stringutils.SlugString(makeName), stringutils.SlugString(modelName), year)
	if err != nil {
		return "", &exceptions.NotFoundError{
			Err: fmt.Errorf("no device definition id can be built for %d %s %s: %w", year, makeName, modelName, err),
		}
	}
	return id, nil
}

// createOrAdoptTemplate creates the template for the first decode of a
// make/model/year and always returns the stored template, whoever wrote it.
//
// Both outcomes have to hand one back. A decode that returned nothing on the
// path where its own create landed skipped trim matching entirely: the first
// decode of a definition answered trim "", template_version 0 and
// match_quality "", while the next decode of the same VIN read the stored Base
// trim and answered "Base" with quality "exact". Two calls a second apart
// disagreeing about one VIN is the defect hydrateResponseFromVinNumber's doc
// comment forbids, and the trim-match counter under-reported every new
// definition on top of it. Create answers with the document the worker stored,
// so this costs no extra request.
//
// Create is create-only. When another writer -- a Console curator saving trims,
// or a concurrent first decode -- stored the template between the decode's
// catalog read and this write, Create returns ErrTemplateExists instead of
// replacing their version with a single Base trim. That is success by someone
// else: their template is read back with a cache-busting read, since the CDN
// may still be serving the 404 the decode just saw, and returned so the decode
// narrows it like any template it found.
func createOrAdoptTemplate(ctx context.Context, catalog gateways.DeviceDefinitionCatalogService, manufacturer string, dd coremodels.DeviceDefinitionTablelandModel) (*coremodels.Template, error) {
	created, err := catalog.Create(ctx, manufacturer, dd)
	switch {
	case err == nil && created != nil:
		return created, nil
	case err == nil:
		// The write landed but its response was not readable as this
		// template. Rare, and not a reason to answer with no match data:
		// read the document back instead.
		stored, _, errRead := catalog.GetTemplateByIDFresh(ctx, dd.ID)
		if errRead != nil {
			return nil, errors.Wrapf(errRead, "template %s was created but could not be read back", dd.ID)
		}
		return stored, nil
	case !errors.Is(err, gateways.ErrTemplateExists):
		return nil, err
	}
	stored, _, err := catalog.GetTemplateByIDFresh(ctx, dd.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "template %s was created by another writer but could not be read back", dd.ID)
	}
	return stored, nil
}

// hydrateResponseFromVinNumber pass in a vin_number database object and converts to vin decode response.
//
// This is the cached path, and it is the one most decodes take: every VIN that
// has been decoded once is answered from vin_numbers thereafter. It therefore
// has to give the SAME answer a fresh decode of the same VIN would -- same
// trim, same match quality, same powertrain. Leaving the match fields unset
// here would emit an undocumented fourth match_quality ("") on the majority of
// production traffic, and sourcing powertrain from template-level attributes
// alone would silently miss it on exactly the multi-trim templates this
// migration exists for (a template only carries powertrain_type at the top
// level when every trim agrees on it), falling through to a make/model
// heuristic -- the blended answer we are replacing.
//
// It returns an error only when the catalog could not be read at all. A
// template that is genuinely absent is answered, as before, with the fields
// the matcher never got to compute left empty.
func (dc DecodeVINQueryHandler) hydrateResponseFromVinNumber(ctx context.Context, vn *models.VinNumber) (*p_grpc.DecodeVinResponse, error) {
	if vn == nil {
		return nil, nil
	}

	resp := &p_grpc.DecodeVinResponse{
		Manufacturer:  vn.ManufacturerName,
		Year:          int32(vn.Year),
		DeviceStyleId: vn.StyleID.String,
		Source:        vn.DecodeProvider.String,
		DefinitionId:  vn.DefinitionID,
	}

	tblDef, _, err := dc.deviceDefinitionCatalogService.GetTemplateByID(ctx, vn.DefinitionID)
	if err != nil && !errors.Is(err, gateways.ErrTemplateNotFound) {
		// A catalog outage (5xx, timeout, decode failure) is not the same as
		// the template not existing, and this is the path most decodes take.
		// Answering OK with empty trim, match quality, candidates, powertrain
		// and hardware template id would make an outage indistinguishable
		// from a genuinely template-less definition -- and would emit the
		// undocumented fourth match_quality ("") on the majority of
		// production traffic. The live path checks the same sentinel the same
		// way; fail the decode instead.
		return nil, errors.Wrapf(err, "failed to read template %s from catalog for cached decode of vin %s", vn.DefinitionID, vn.Vin)
	}
	if err != nil || tblDef == nil {
		// Genuinely not found (ErrTemplateNotFound): this is not good, somehow
		// it got decoded in past without a template existing for it.
		// MatchQuality stays empty rather than "model-only": the matcher did
		// not run at all, and claiming a quality it never computed would be
		// the kind of authoritative-looking wrong answer this migration exists
		// to stop.
		dc.logger.Warn().Err(err).Msgf("vin decoded for unexistent device definition: %s, vin: %s", vn.DefinitionID, vn.Vin)
		return resp, nil
	}

	// The model is the template's, not the vin_numbers row's: vin_numbers
	// stores the manufacturer name and the definition id but never the model,
	// and a response without it answered "" for every decode after the first
	// while a fresh decode of the same VIN answered the model.
	resp.Model = tblDef.Model

	resolved := services.MatchTrim(tblDef, dc.matchSignalsFromVinNumber(ctx, vn))
	resp.Trim = resolved.Trim
	resp.TemplateVersion = int32(resolved.TemplateVersion)
	resp.MatchQuality = string(resolved.Quality)
	resp.MatchCandidates = resolved.Candidates
	resp.MatchBy = resolved.MatchedBy
	resp.HardwareTemplateId = resolved.HardwareTemplateID
	// No heuristic fallback, deliberately: the live path in Handle reports no
	// powertrain when the resolved attributes carry none, and these two paths
	// answering the same VIN differently is the defect this function is
	// fixing, not a behaviour worth keeping on one side of it.
	if v, ok := resolved.Attributes[common.PowerTrainType].(string); ok {
		resp.Powertrain = v
	}

	observeTrimMatch(resolved, resp.Source)

	return resp, nil
}

// matchSignalsFromVinNumber rebuilds the signals a fresh decode of this VIN
// would have produced, out of what saveVinDecodeNumber persisted. Every source
// here is the same value the live path fed the matcher: drivly_data is the
// marshalled DrivlyVINResponse vinInfo.ManufacturerCode was read from, and the
// style row's name is the vinInfo.StyleName processDeviceStyle stored. A
// signal that was never persisted stays empty, which the matcher treats as
// "not available" rather than as a non-match.
func (dc DecodeVINQueryHandler) matchSignalsFromVinNumber(ctx context.Context, vn *models.VinNumber) services.MatchSignals {
	sig := services.MatchSignals{VIN: vn.Vin}

	if vn.DrivlyData.Valid {
		sig.ManufacturerCode = gjson.GetBytes(vn.DrivlyData.JSON, "manufacturerCode").String()
	}

	if vn.StyleID.Valid && vn.StyleID.String != "" {
		style, err := models.DeviceStyles(models.DeviceStyleWhere.ID.EQ(vn.StyleID.String)).One(ctx, dc.dbs().Reader)
		if err != nil {
			// Best effort: a missing style row costs us one selector, which
			// shows up as a lower match quality rather than as a wrong trim.
			dc.logger.Debug().Err(err).Msgf("could not load style %s for cached decode of vin %s", vn.StyleID.String, vn.Vin)
		} else {
			sig.StyleName = style.Name
		}
	}

	return sig
}

// observeTrimMatch records how a decode resolved, so the rate of exact vs
// ambiguous vs model-only is answerable from production rather than from
// inspection. The plan's riskiest assumption is that manufacturerCode reaches
// a decode as a usable signal at all -- it only ever arrives via drivly -- and
// this counter, broken down by decode source, is what makes that measurable
// instead of merely asserted.
func observeTrimMatch(resolved services.Resolved, source string) {
	if source == "" {
		source = "unknown"
	}
	metrics.TrimMatchQuality.With(prometheus.Labels{
		"quality": string(resolved.Quality),
		"source":  source,
	}).Inc()
}

// processDeviceStyle saves new styles if needed to db and returns the style database ID
//
// Every read here is checked for a failure that is not no-rows. sqlboiler's
// One() answers such a failure with a nil row and a wrapped error, so a reader
// that is merely unavailable for a moment -- a pool blip, a reset connection,
// the caller's context deadline expiring right after the catalog round trip
// above -- used to fall through both no-rows branches and nil-dereference the
// style on the way out, panicking the decode hot path. Handle already treats
// an error from here as "continue without a style id"; it just never got one.
func (dc DecodeVINQueryHandler) processDeviceStyle(ctx context.Context, vinInfo *coremodels.VINDecodingInfoData, definitionID, powertrain string) (string, error) {
	externalStyleID := stringutils.SlugString(vinInfo.StyleName)

	// Step 1: Try to find an existing style
	style, err := models.DeviceStyles(
		models.DeviceStyleWhere.DefinitionID.EQ(definitionID),
		models.DeviceStyleWhere.Source.EQ(string(vinInfo.Source)),
		models.DeviceStyleWhere.ExternalStyleID.EQ(externalStyleID),
	).One(ctx, dc.dbs().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", errors.Wrapf(err, "error querying device_styles for definition %s by external style id %s", definitionID, externalStyleID)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// Step 2: If not found, try searching by name
		style, err = models.DeviceStyles(
			models.DeviceStyleWhere.DefinitionID.EQ(definitionID),
			models.DeviceStyleWhere.Name.EQ(vinInfo.StyleName),
		).One(ctx, dc.dbs().Reader)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", errors.Wrapf(err, "error querying device_styles for definition %s by name %s", definitionID, vinInfo.StyleName)
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		// Step 3: Create a new style if it doesn't exist
		style = &models.DeviceStyle{
			ID:              ksuid.New().String(),
			DefinitionID:    definitionID,
			Name:            vinInfo.StyleName,
			ExternalStyleID: externalStyleID,
			Source:          string(vinInfo.Source),
			SubModel:        vinInfo.SubModel,
			Metadata:        vinInfo.MetaData,
		}

		// Resolve powertrain and add to metadata if applicable
		if powertrain != "" {
			metadataWithPT, metadataErr := sjson.SetBytes(vinInfo.MetaData.JSON, common.PowerTrainType, powertrain)
			if metadataErr == nil {
				style.Metadata = null.JSONFrom(metadataWithPT)
			}
		}

		// Insert the new style into the database
		errStyle := style.Insert(ctx, dc.dbs().Writer, boil.Infer())
		if errStyle != nil {
			return "", errors.Wrapf(errStyle, "error creating style with values: %+v", style)
		}
	}
	return style.ID, nil
}

func (dc DecodeVINQueryHandler) saveVinDecodeNumber(ctx context.Context, vinObj vin.VIN, vinInfo *coremodels.VINDecodingInfoData, resp *p_grpc.DecodeVinResponse) error {
	vinDecodeNumber := &models.VinNumber{
		Vin:              vinObj.String(),
		ManufacturerName: resp.Manufacturer,
		Wmi:              null.StringFrom(vinObj.Wmi()),
		SerialNumber:     vinObj.SerialNumber(),
		DecodeProvider:   null.StringFrom(string(vinInfo.Source)),
		Year:             int(resp.Year),
		DefinitionID:     resp.DefinitionId,
	}
	if vinObj.IsValidVIN() {
		vinDecodeNumber.VDS = null.StringFrom(vinObj.VDS())
		vinDecodeNumber.Vis = null.StringFrom(vinObj.VIS())
		vinDecodeNumber.CheckDigit = null.StringFrom(vinObj.CheckDigit())
	}

	// Optional fields based on response and VIN info
	if len(resp.DeviceStyleId) > 0 {
		vinDecodeNumber.StyleID = null.StringFrom(resp.DeviceStyleId)
	}

	switch vinInfo.Source {
	case coremodels.DrivlyProvider:
		if len(vinInfo.Raw) > 0 {
			vinDecodeNumber.DrivlyData = null.JSONFrom(vinInfo.Raw)
		}
	case coremodels.VincarioProvider:
		if len(vinInfo.Raw) > 0 {
			vinDecodeNumber.VincarioData = null.JSONFrom(vinInfo.Raw)
		}
	case coremodels.AutoIsoProvider:
		if len(vinInfo.Raw) > 0 {
			vinDecodeNumber.AutoisoData = null.JSONFrom(vinInfo.Raw)
		}
	case coremodels.DATGroupProvider:
		if len(vinInfo.Raw) > 0 {
			vinDecodeNumber.DatgroupData = null.JSONFrom(vinInfo.Raw)
		}
	case coremodels.Japan17VIN:
		if len(vinInfo.Raw) > 0 {
			vinDecodeNumber.Vin17Data = null.JSONFrom(vinInfo.Raw)
		}
	case coremodels.CarVXVIN:
		// we currently do not store the raw payload since seemed to not gain much for now
	}

	// Insert VIN decode number into the database
	if err := vinDecodeNumber.Insert(ctx, dc.dbs().Writer, boil.Infer()); err != nil {
		return errors.Wrapf(err, "error inserting vin_number with values: %+v", vinDecodeNumber)
	}
	return nil
}

// vinInfoFromKnown builds a vininfo object based on one passed in with Make from vin WMI, and passed in model and year set
func (dc DecodeVINQueryHandler) vinInfoFromKnown(ctx context.Context, vin vin.VIN, knownModel string, knownYear int32) (*coremodels.VINDecodingInfoData, error) {
	vinInfo := &coremodels.VINDecodingInfoData{}
	vinInfo.VIN = vin.String()
	wmis, err := models.Wmis(models.WmiWhere.Wmi.EQ(vin.Wmi())).All(ctx, dc.dbs().Reader)
	if err != nil {
		return nil, errors.Wrap(err, "vinInfoFromKnown: failed to read wmis for "+vin.Wmi())
	}
	vinInfo.Make, err = dc.makeFromWMIRows(ctx, vin.Wmi(), wmis, knownModel, knownYear)
	if err != nil {
		return nil, err
	}
	vinInfo.Year = knownYear
	vinInfo.Model = knownModel
	vinInfo.Source = "probably smartcar"

	if len(vinInfo.Model) == 0 || len(vinInfo.Make) == 0 || vinInfo.Year == 0 {
		return nil, fmt.Errorf("vinInfoFromKnown: unable to decode from known info")
	}

	return vinInfo, nil
}

// makeFromWMIRows names the manufacturer for a VIN whose wmis rows have
// already been read. A WMI can be shared by several marques of the same parent
// OEM, in which case the one that already has a template for this model-year
// is the right answer.
func (dc DecodeVINQueryHandler) makeFromWMIRows(ctx context.Context, wmiCode string, wmis models.WmiSlice, knownModel string, knownYear int32) (string, error) {
	if len(wmis) == 0 {
		// .All() reports an unknown WMI as an empty slice with a nil error --
		// unlike .One(), which returns sql.ErrNoRows -- so this has to be
		// checked explicitly. Indexing element zero here panicked the decode
		// handler for exactly the VINs this fallback exists to serve.
		return "", &exceptions.NotFoundError{Err: fmt.Errorf("vinInfoFromKnown: unknown WMI %s", wmiCode)}
	}
	if len(wmis) > 1 {
		// see if we can find an existing device definition for this WMI
		makeNamesForError := ""
		for _, wmi := range wmis {
			makeNamesForError += wmi.ManufacturerName + ", "
			definitionID, errID := common.DeviceDefinitionSlug(stringutils.SlugString(wmi.ManufacturerName), stringutils.SlugString(knownModel), int16(knownYear))
			if errID != nil {
				// No template can be stored at an id the worker refuses, so
				// this marque cannot be the one with a definition for the
				// model-year. Skipping it costs a catalog request that would
				// 404 anyway; the loop still reports every marque it weighed.
				continue
			}
			tmpl, _, err := dc.deviceDefinitionCatalogService.GetTemplateByID(ctx, definitionID)
			if err == nil && tmpl != nil {
				return wmi.ManufacturerName, nil
			}
		}
		// no matching DD's found. We don't have a good way to determine the right Make / OEM
		return "", fmt.Errorf("vinInfoFromKnown: unable to determine the right OEM between %sfor WMI %s", makeNamesForError, wmiCode)
	}
	return wmis[0].ManufacturerName, nil
}

func (dc DecodeVINQueryHandler) associateImagesToDeviceDefinition(ctx context.Context, definitionID, mk, model string, year int, prodID int, prodFormat int) error {

	img, err := dc.fuelAPIService.FetchDeviceImages(mk, model, year, prodID, prodFormat)
	if err != nil {
		dc.logger.Warn().Err(err).Msgf("unable to fetch device image for: %d %s %s", year, mk, model)
		return nil
	}

	var p models.Image

	// loop through all img (color variations)
	for _, device := range img.Images {
		p.ID = ksuid.New().String()
		p.DefinitionID = definitionID
		p.FuelAPIID = null.StringFrom(img.FuelAPIID)
		p.Width = null.IntFrom(img.Width)
		p.Height = null.IntFrom(img.Height)
		p.SourceURL = device.SourceURL
		//p.DimoS3URL = null.StringFrom("") // dont set it so it is null
		p.Color = device.Color
		p.NotExactImage = img.NotExactImage

		err = p.Upsert(ctx, dc.dbs().Writer, true, []string{models.ImageColumns.DefinitionID, models.ImageColumns.SourceURL}, boil.Infer(), boil.Infer())
		if err != nil {
			dc.logger.Warn().Err(err).Msgf("fail insert device image for: %s %d %s %s", definitionID, year, mk, model)
			continue
		}
	}

	return nil
}

// isLowConfidenceSource returns true for decode providers whose output has historically
// produced bad device definitions (e.g., Japanese chassis decoders returning body-style
// codes as model names). For these sources a decode refuses to create a template on a
// catalog miss and surfaces a NotFoundError so the client can fall back to manual
// make/model/year selection. High-confidence Western providers (Drivly, Vincario, DATGroup,
// Tesla) keep creating the template.
func isLowConfidenceSource(src coremodels.DecodeProviderEnum) bool {
	switch src {
	case coremodels.Japan17VIN, coremodels.CarVXVIN, coremodels.AutoIsoProvider, coremodels.ElevaKaufmannProvider:
		return true
	}
	return false
}
