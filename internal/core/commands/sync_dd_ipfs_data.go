package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gosimple/slug"
	shell "github.com/ipfs/go-ipfs-api"
	files "github.com/ipfs/go-ipfs-files"
)

type SyncIPFSDataCommand struct {
}

type SyncIPFSDataCommandResult struct {
}

func (*SyncIPFSDataCommand) Key() string { return "SyncIPFSDataCommand" }

type SyncIPFSDataCommandHandler struct {
	DBS          func() *db.ReaderWriter
	IPFSEndpoint string
}

func NewSyncIPFSDataCommandHandler(dbs func() *db.ReaderWriter, IPFSEndpoint string) SyncIPFSDataCommandHandler {
	return SyncIPFSDataCommandHandler{DBS: dbs, IPFSEndpoint: IPFSEndpoint}
}

func (ch SyncIPFSDataCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	sh := shell.NewShell(ch.IPFSEndpoint)

	all, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Verified.EQ(true),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).All(ctx, ch.DBS().Reader)

	if err != nil {
		return nil, err
	}

	makes, err := models.DeviceMakes().All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, err
	}

	basePath := "/makes"

	_, err = sh.FileList(basePath)

	if err != nil && !strings.Contains(err.Error(), "invalid path") {
		fmt.Printf(err.Error())
		return nil, err
	}

	fmt.Printf("Creating %s directory", basePath)
	err = sh.FilesMkdir(ctx, basePath)

	if err != nil && !strings.Contains(err.Error(), "file already exists") {
		fmt.Printf("error creating %s directory %s", basePath, err.Error())
		return nil, err
	}

	fmt.Printf("Creation of makes folders")

	for _, v := range makes {
		// create make path
		path := fmt.Sprintf("%s/%s", basePath, slug.Make(v.Name))

		fmt.Printf("Creating make directory => %s", path)

		err := sh.FilesMkdir(ctx, path)
		if err != nil && !strings.Contains(err.Error(), "file already exists") {
			fmt.Printf("error creating make %s directory %s", basePath, err.Error())
			return nil, err
		}

		tsdBin, _ := json.Marshal(v)
		reader := bytes.NewReader(tsdBin)

		// create index.json
		fr := files.NewReaderFile(reader)
		slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
		fileReader := files.NewMultiFileReader(slf, true)

		indexFilePath := fmt.Sprintf("%s/index.json", path)

		fmt.Printf(indexFilePath)

		rb := sh.Request("files/write", indexFilePath)
		rb.Option("create", "true")

		err = rb.Body(fileReader).Exec(ctx, nil)
		if err != nil {
			return nil, err
		}

	}

	fmt.Printf("Creation of models folders")

	for _, definition := range all {
		path := fmt.Sprintf("%s/%s/%s",
			basePath,
			slug.Make(definition.R.DeviceMake.Name),
			slug.Make(definition.Model))

		fmt.Printf("Creating model directory => %s", path)

		err := sh.FilesMkdir(ctx, path)
		if err != nil && !strings.Contains(err.Error(), "file already exists") {
			fmt.Printf("error creating model %s directory %s", basePath, err.Error())
			return nil, err
		}

		tsdBin, _ := json.Marshal(definition)
		reader := bytes.NewReader(tsdBin)

		// create index.json
		fr := files.NewReaderFile(reader)
		slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
		fileReader := files.NewMultiFileReader(slf, true)

		indexFilePath := fmt.Sprintf("%s/index.json", path)

		fmt.Printf(indexFilePath)

		rb := sh.Request("files/write", indexFilePath)
		rb.Option("create", "true")

		err = rb.Body(fileReader).Exec(ctx, nil)
		if err != nil {
			return nil, err
		}
	}

	fmt.Printf("Creation of models/years folders")

	for _, definition := range all {
		path := fmt.Sprintf("%s/%s/%s/%d",
			basePath,
			slug.Make(definition.R.DeviceMake.Name),
			slug.Make(definition.Model),
			definition.Year)

		fmt.Printf("Creating model/year directory => %s", path)

		err := sh.FilesMkdir(ctx, path)
		if err != nil && !strings.Contains(err.Error(), "file already exists") {
			fmt.Printf("error creating model/year %s directory %s", basePath, err.Error())
			return nil, err
		}

		tsdBin, _ := json.Marshal(definition)
		reader := bytes.NewReader(tsdBin)

		// create index.json
		fr := files.NewReaderFile(reader)
		slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
		fileReader := files.NewMultiFileReader(slf, true)

		indexFilePath := fmt.Sprintf("%s/index.json", path)

		fmt.Printf(indexFilePath)

		rb := sh.Request("files/write", indexFilePath)
		rb.Option("create", "true")

		err = rb.Body(fileReader).Exec(ctx, nil)
		if err != nil {
			return nil, err
		}
	}

	fmt.Printf("Done !!")

	return SyncIPFSDataCommandResult{}, nil
}
