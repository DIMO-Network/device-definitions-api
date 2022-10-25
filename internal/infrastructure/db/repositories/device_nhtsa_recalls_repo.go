//go:generate mockgen -source device_nhtsa_recalls_repo.go -destination mocks/device_nhtsa_recalls_repo_mock.go -package mocks

package repositories

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
)

type DeviceNHTSARecallsRepository interface {
	Create(ctx context.Context, deviceDefinitionID null.String, data []string, metadata null.JSON, hash []byte) (*models.DeviceNhtsaRecall, error)
	GetLastDataRecordID(ctx context.Context) (*null.Int, error)
	MatchDeviceDefinition(ctx context.Context, matchingVersion string) (int64, error)
}

type deviceNHTSARecallsRepository struct {
	DBS func() *db.ReaderWriter
}

func NewDeviceNHTSARecallsRepository(dbs func() *db.ReaderWriter) DeviceNHTSARecallsRepository {
	return &deviceNHTSARecallsRepository{DBS: dbs}
}

func (r *deviceNHTSARecallsRepository) Create(ctx context.Context, deviceDefinitionID null.String, row []string, metadata null.JSON, hash []byte) (*models.DeviceNhtsaRecall, error) {

	if !deviceDefinitionID.IsZero() && deviceDefinitionID.String == "" {
		deviceDefinitionID = null.String{}
	}

	if len(row) == 0 {
		return nil, errors.New("NHTSA Recall record can not be empty")
	}
	drID, err := strconv.Atoi(row[0])
	if err != nil {
		return nil, errors.New("NHTSA Recall record ID must be a number")
	}
	if len(row) < 27 {
		return nil, errors.Errorf("NHTSA Recall record ID %d has %d columns, expected %d at minimum", drID, len(row), 27)
	}

	dnr := &models.DeviceNhtsaRecall{
		ID:                   ksuid.New().String(),
		DeviceDefinitionID:   deviceDefinitionID,
		DataRecordID:         drID,
		DataCampno:           row[1],
		DataMaketxt:          row[2],
		DataModeltxt:         row[3],
		DataYeartxt:          r.nullableInt(row[4]).Int,
		DataMfgcampno:        row[5],
		DataCompname:         row[6],
		DataMfgname:          row[7],
		DataBgman:            r.nullableDate(row[8]),
		DataEndman:           r.nullableDate(row[9]),
		DataRcltypecd:        row[10],
		DataPotaff:           r.nullableInt(row[11]),
		DataOdate:            r.nullableDate(row[12]),
		DataInfluencedBy:     row[13],
		DataMFGTXT:           row[14],
		DataRcdate:           r.nullableDate(row[15]).Time,
		DataDatea:            r.nullableDate(row[16]).Time,
		DataRpno:             row[17],
		DataFMVSS:            row[18],
		DataDescDefect:       row[19],
		DataConequenceDefect: row[20],
		DataCorrectiveAction: row[21],
		DataNotes:            row[22],
		DataRCLCMPTID:        row[23],
		DataMFRCompName:      row[24],
		DataMFRCompDesc:      row[25],
		DataMFRCompPtno:      row[26],
		Metadata:             metadata,
		Hash:                 hash,
	}
	err = dnr.Insert(ctx, r.DBS().Writer, boil.Infer())
	if err != nil {
		// ignore duplicate key errors
		if strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint "device_nhtsa_recalls_hash"`) {
			return nil, nil
		}
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
