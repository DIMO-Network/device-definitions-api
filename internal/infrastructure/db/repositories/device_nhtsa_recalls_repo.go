//go:generate mockgen -source device_nhtsa_recalls_repo.go -destination mocks/device_nhtsa_recalls_repo_mock.go -package mocks

package repositories

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"strconv"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DeviceNHTSARecallsRepository interface {
	Create(ctx context.Context, deviceDefinitionID null.String, data []string, metadata null.JSON) (*models.DeviceNhtsaRecall, error)
	GetLastDataRecordID(ctx context.Context) (*null.Int, error)
	MatchDeviceDefinition(ctx context.Context, matchingVersion string) (int64, error)
	GetByID(ctx context.Context, id string) (*models.DeviceNhtsaRecall, error)
	SetDDAndMetadata(ctx context.Context, recall models.DeviceNhtsaRecall, deviceDefinitionID *string, metadata *null.JSON) error
}

type deviceNHTSARecallsRepository struct {
	DBS func() *db.ReaderWriter
}

func NewDeviceNHTSARecallsRepository(dbs func() *db.ReaderWriter) DeviceNHTSARecallsRepository {
	return &deviceNHTSARecallsRepository{DBS: dbs}
}

func (r *deviceNHTSARecallsRepository) Create(ctx context.Context, deviceDefinitionID null.String, data []string, metadata null.JSON) (*models.DeviceNhtsaRecall, error) {

	if !deviceDefinitionID.IsZero() && deviceDefinitionID.String == "" {
		deviceDefinitionID = null.String{}
	}

	if len(data) == 0 {
		return nil, errors.New("NHTSA Recall record can not be empty")
	}
	drID, err := strconv.Atoi(data[0])
	if err != nil {
		return nil, errors.New("NHTSA Recall record ID must be a number")
	}
	if len(data) < 27 {
		return nil, errors.Errorf("NHTSA Recall record ID %d has %d columns, expected %d at minimum", drID, len(data), 27)
	}
	dnr := &models.DeviceNhtsaRecall{
		ID:                   ksuid.New().String(),
		DeviceDefinitionID:   deviceDefinitionID,
		DataRecordID:         drID,
		DataCampno:           data[1],
		DataMaketxt:          data[2],
		DataModeltxt:         data[3],
		DataYeartxt:          r.nullableInt(data[4]).Int,
		DataMfgcampno:        data[5],
		DataCompname:         data[6],
		DataMfgname:          data[7],
		DataBgman:            r.nullableDate(data[8]),
		DataEndman:           r.nullableDate(data[9]),
		DataRcltypecd:        data[10],
		DataPotaff:           r.nullableInt(data[11]),
		DataOdate:            r.nullableDate(data[12]),
		DataInfluencedBy:     data[13],
		DataMFGTXT:           data[14],
		DataRcdate:           r.nullableDate(data[15]).Time,
		DataDatea:            r.nullableDate(data[16]).Time,
		DataRpno:             data[17],
		DataFMVSS:            data[18],
		DataDescDefect:       data[19],
		DataConequenceDefect: data[20],
		DataCorrectiveAction: data[21],
		DataNotes:            data[22],
		DataRCLCMPTID:        data[23],
		DataMFRCompName:      data[24],
		DataMFRCompDesc:      data[25],
		DataMFRCompPtno:      data[26],
		Metadata:             metadata,
	}
	err = dnr.Insert(ctx, r.DBS().Writer, boil.Infer())
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return dnr, nil
}

func (r *deviceNHTSARecallsRepository) GetLastDataRecordID(ctx context.Context) (*null.Int, error) {
	recall, err := models.DeviceNhtsaRecalls(
		qm.OrderBy("data_record_id DESC"),
		qm.Limit(1),
	).One(ctx, r.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &null.Int{}, nil
		}
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}
	ret := null.IntFrom(recall.DataRecordID)
	return &ret, nil
}

func (r *deviceNHTSARecallsRepository) MatchDeviceDefinition(ctx context.Context, matchingVersion string) (int64, error) {
	updateMatching := `UPDATE device_nhtsa_recalls
		SET
			device_definition_id = matches.dd_id,
			metadata = COALESCE(metadata,'{}'::jsonb) || json_build_object(
				'matchingVersion',$1::text,
				'matchType',matches.match_type
				)::jsonb,
		    updated_at = NOW()
		FROM (
			SELECT
				dnr.id,
				dd.id AS "dd_id",
				CASE
					WHEN dd.model = dnr.data_modeltxt
						THEN 'EXACT'
					WHEN dd.model ILIKE dnr.data_modeltxt
						THEN 'EXACT CI'
					WHEN dd.model IS NOT NULL
						THEN 'ALPHANUM CI'
					ELSE 'NONE'
					END AS "match_type"
			FROM device_nhtsa_recalls dnr
			LEFT JOIN device_makes dm
				ON regexp_replace(dm.name, '\W+', '', 'g') ILIKE regexp_replace(dnr.data_maketxt, '\W+', '', 'g')
			LEFT JOIN device_definitions dd
				ON dm.id = dd.device_make_id
					   AND dd.year = dnr.data_yeartxt
					   AND regexp_replace(dd.model, '\W+', '', 'g') ILIKE regexp_replace(dnr.data_modeltxt, '\W+', '', 'g')
			WHERE (dnr.metadata ->> 'matchingVersion') <> $1::text OR dnr.metadata IS NULL
			ORDER BY
				dnr.data_record_id ASC,
				dd.model ilike dnr.data_modeltxt
			 ) matches
		WHERE matches.id = device_nhtsa_recalls.id`
	result, err := r.DBS().Writer.Exec(updateMatching, matchingVersion)
	if err != nil {
		return 0, &exceptions.InternalError{
			Err: errors.Wrap(err, "failed to exec sql"),
		}
	}
	matchedCount, err := result.RowsAffected()
	if err != nil {
		return 0, &exceptions.InternalError{
			Err: errors.Wrap(err, "filed to get affected row count"),
		}
	}
	return matchedCount, nil
}

func (r *deviceNHTSARecallsRepository) GetByID(ctx context.Context, id string) (*models.DeviceNhtsaRecall, error) {
	recall, err := models.DeviceNhtsaRecalls(
		qm.Where("id = ?", id),
	).One(ctx, r.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.DeviceNhtsaRecall{}, nil
		}
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}
	return recall, nil
}

func (r *deviceNHTSARecallsRepository) SetDDAndMetadata(ctx context.Context, recall models.DeviceNhtsaRecall, deviceDefinitionID *string, metadata *null.JSON) error {
	if deviceDefinitionID == nil {
		recall.DeviceDefinitionID = null.StringFromPtr(deviceDefinitionID)
	}
	if metadata == nil {
		recall.Metadata = *metadata
	}
	_, err := recall.Update(ctx, r.DBS().Writer, boil.Infer())
	if err != nil {
		return &exceptions.InternalError{
			Err: err,
		}
	}
	return nil
}

func (r *deviceNHTSARecallsRepository) nullableInt(value string) null.Int {
	if i, err := strconv.Atoi(value); err == nil {
		return null.IntFrom(i)
	}
	return null.Int{}
}
func (r *deviceNHTSARecallsRepository) nullableDate(value string) null.Time {
	if t, err := time.Parse("20060102", value); err == nil {
		return null.TimeFrom(t)
	}
	return null.Time{}
}