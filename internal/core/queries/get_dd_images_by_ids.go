package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetDeviceDefinitionImagesByIdsQuery struct {
	DeviceDefinitionID []string `json:"deviceDefinitionId" validate:"required"`
}

func (*GetDeviceDefinitionImagesByIdsQuery) Key() string {
	return "GetDeviceDefinitionImagesByIdsQuery"
}

type GetDeviceDefinitionImagesByIdsQueryHandler struct {
	log *zerolog.Logger
	dbs func() *db.ReaderWriter
}

func NewGetDeviceDefinitionImagesByIdsQueryHandler(dbs func() *db.ReaderWriter, log *zerolog.Logger) GetDeviceDefinitionImagesByIdsQueryHandler {
	return GetDeviceDefinitionImagesByIdsQueryHandler{
		log: log,
		dbs: dbs,
	}
}

func (ch GetDeviceDefinitionImagesByIdsQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionImagesByIdsQuery)

	if len(qry.DeviceDefinitionID) == 0 {
		return nil, &exceptions.ValidationError{
			Err: errors.New("Device Definition Ids is required"),
		}
	}

	response := &grpc.GetDeviceImagesResponse{Images: make([]*grpc.DeviceImage, 0)}
	all, err := models.Images(models.ImageWhere.DeviceDefinitionID.IN(qry.DeviceDefinitionID), qm.OrderBy(models.ImageColumns.DeviceDefinitionID)).All(ctx, ch.dbs().Reader)
	if err != nil {
		return nil, err
	}
	// filter for one image in each width/height size in preffered color
	for _, image := range all {
		// see if response.Images already has this image width, if not, add it, if it does, then does this image have a color we prefere?
		if ei := findImage(response.Images, image.Width.Int, image.DeviceDefinitionID); ei == nil {
			response.Images = append(response.Images, &grpc.DeviceImage{
				DeviceDefinitionId: image.DeviceDefinitionID,
				ImageUrl:           image.SourceURL,
				Width:              int32(image.Width.Int),
				Height:             int32(image.Height.Int),
				Color:              image.Color,
			})
		} else {
			// ei is a pointer so i should be able to just modify values in it
			if ei.Color == "Silver" || ei.Color == "White" || ei.Color == "Red" {
				continue
			}
			if image.Color == "Silver" || image.Color == "White" || image.Color == "Red" {
				ei.Color = image.Color
				ei.ImageUrl = image.SourceURL
			}
		}
	}

	return response, nil
}

func findImage(images []*grpc.DeviceImage, width int, ddID string) *grpc.DeviceImage {
	for _, image := range images {
		if image.Width == int32(width) && image.DeviceDefinitionId == ddID {
			return image
		}
	}
	return nil
}
