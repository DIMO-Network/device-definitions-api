package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GetDeviceDefinitionImagesByIDsQuery struct {
	DefinitionID []string `json:"definitionId" validate:"required"`
}

func (*GetDeviceDefinitionImagesByIDsQuery) Key() string {
	return "GetDeviceDefinitionImagesByIDsQuery"
}

type GetDeviceDefinitionImagesByIDsQueryHandler struct {
	log *zerolog.Logger
	dbs func() *db.ReaderWriter
}

func NewGetDeviceDefinitionImagesByIDsQueryHandler(dbs func() *db.ReaderWriter, log *zerolog.Logger) GetDeviceDefinitionImagesByIDsQueryHandler {
	return GetDeviceDefinitionImagesByIDsQueryHandler{
		log: log,
		dbs: dbs,
	}
}

func (ch GetDeviceDefinitionImagesByIDsQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionImagesByIDsQuery)

	if len(qry.DefinitionID) == 0 {
		return nil, &exceptions.ValidationError{
			Err: errors.New("Device Definition Ids is required"),
		}
	}

	response := &grpc.GetDeviceImagesResponse{Images: make([]*grpc.DeviceImage, 0)}
	all, err := models.Images(models.ImageWhere.DefinitionID.IN(qry.DefinitionID), qm.OrderBy(models.ImageColumns.DefinitionID)).All(ctx, ch.dbs().Reader)
	if err != nil {
		return nil, err
	}
	// filter for one image in each width/height size in preffered color
	for _, image := range all {
		// see if response.Images already has this image width, if not, add it, if it does, then does this image have a color we prefere?
		if ei := findImage(response.Images, image.Width.Int, image.DefinitionID); ei == nil {
			response.Images = append(response.Images, &grpc.DeviceImage{
				DefinitionId: image.DefinitionID,
				ImageUrl:     image.SourceURL,
				Width:        int32(image.Width.Int),
				Height:       int32(image.Height.Int),
				Color:        image.Color,
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

func findImage(images []*grpc.DeviceImage, width int, definitionID string) *grpc.DeviceImage {
	for _, image := range images {
		if image.Width == int32(width) && image.DefinitionId == definitionID {
			return image
		}
	}
	return nil
}
