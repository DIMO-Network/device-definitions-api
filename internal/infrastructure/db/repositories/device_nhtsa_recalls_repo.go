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
	GetLastDataRecordID(ctx context.Context) (null.Int, error)
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

func (r *deviceNHTSARecallsRepository) GetLastDataRecordID(ctx context.Context) (null.Int, error) {
	recall, err := models.DeviceNhtsaRecalls(
		qm.OrderBy("data_record_id DESC"),
		qm.Limit(1),
	).One(ctx, r.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return null.Int{}, nil
		}
		return null.Int{}, err
	}
	return null.IntFrom(recall.DataRecordID), nil
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
