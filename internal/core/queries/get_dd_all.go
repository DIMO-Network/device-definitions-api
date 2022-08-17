package queries

import (
	"context"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetAllDeviceDefinitionQuery struct {
}

type GetAllDeviceDefinitionQueryResult struct {
	Make   string            `json:"make"`
	Models []GetDeviceModels `json:"models"`
}

func (*GetAllDeviceDefinitionQuery) Key() string { return "GetAllDeviceDefinitionQuery" }

type GetDeviceModels struct {
	Model string               `json:"model"`
	Years []GetDeviceModelYear `json:"years"`
}

type GetDeviceModelYear struct {
	Year               int16  `json:"year"`
	DeviceDefinitionID string `json:"id"`
}

type GetAllDeviceDefinitionQueryHandler struct {
	Repository     repositories.DeviceDefinitionRepository
	MakeRepository repositories.DeviceMakeRepository
}

func NewGetAllDeviceDefinitionQueryHandler(repository repositories.DeviceDefinitionRepository, makeRepository repositories.DeviceMakeRepository) GetAllDeviceDefinitionQueryHandler {
	return GetAllDeviceDefinitionQueryHandler{
		Repository:     repository,
		MakeRepository: makeRepository,
	}
}

func (ch GetAllDeviceDefinitionQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	all, _ := ch.Repository.GetAll(ctx, true)
	allMakes, _ := ch.MakeRepository.GetAll(ctx)

	var result []GetAllDeviceDefinitionQueryResult
	for _, dd := range all {
		makeName := ""
		for _, mk := range allMakes {
			if mk.ID == dd.DeviceMakeID {
				makeName = mk.Name
				break
			}
		}
		idx := indexOfMake(result, makeName)
		// append make if not found
		if idx == -1 {
			result = append(result, GetAllDeviceDefinitionQueryResult{
				Make:   makeName,
				Models: []GetDeviceModels{{Model: dd.Model, Years: []GetDeviceModelYear{{Year: dd.Year, DeviceDefinitionID: dd.ID}}}},
			})
		} else {
			// attach model or year to existing make, lookup model
			idx2 := indexOfModel(result[idx].Models, dd.Model)
			if idx2 == -1 {
				// append model if not found
				result[idx].Models = append(result[idx].Models, GetDeviceModels{
					Model: dd.Model,
					Years: []GetDeviceModelYear{{Year: dd.Year, DeviceDefinitionID: dd.ID}},
				})
			} else {
				// make and model already found, just add year
				result[idx].Models[idx2].Years = append(result[idx].Models[idx2].Years, GetDeviceModelYear{Year: dd.Year, DeviceDefinitionID: dd.ID})
			}
		}
	}

	return result, nil
}

func indexOfMake(makes []GetAllDeviceDefinitionQueryResult, make string) int {
	for i, root := range makes {
		if root.Make == make {
			return i
		}
	}
	return -1
}

func indexOfModel(models []GetDeviceModels, model string) int {
	for i, m := range models {
		if m.Model == model {
			return i
		}
	}
	return -1
}
