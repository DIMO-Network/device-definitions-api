package queries

import (
	"context"

	interfaces "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/interfaces/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetAllQuery struct {
}

type GetAllQueryResult struct {
	Make   string            `json:"make"`
	Models []GetDeviceModels `json:"models"`
}

func (*GetAllQuery) Key() string { return "GetAllQuery" }

type GetDeviceModels struct {
	Model string               `json:"model"`
	Years []GetDeviceModelYear `json:"years"`
}

type GetDeviceModelYear struct {
	Year               int16  `json:"year"`
	DeviceDefinitionID string `json:"id"`
}

type GetAllQueryHandler struct {
	Repository     interfaces.IDeviceDefinitionRepository
	MakeRepository interfaces.IDeviceMakeRepository
}

func NewGetAllQueryHandler(repository interfaces.IDeviceDefinitionRepository, makeRepository interfaces.IDeviceMakeRepository) GetAllQueryHandler {
	return GetAllQueryHandler{
		Repository:     repository,
		MakeRepository: makeRepository,
	}
}

func (ch GetAllQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	all, _ := ch.Repository.GetAll(ctx)
	allMakes, _ := ch.MakeRepository.GetAll(ctx)

	var result []GetAllQueryResult
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
			result = append(result, GetAllQueryResult{
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

func indexOfMake(makes []GetAllQueryResult, make string) int {
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
