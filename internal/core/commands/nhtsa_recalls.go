package commands

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/null/v8"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

type SyncNHTSARecallsCommand struct {
}

type SyncNHTSARecallsCommandResult struct {
	InsertedCount int `json:"insertedCount"`
	MatchCount    int `json:"matchCount"`
}

func (*SyncNHTSARecallsCommand) Key() string { return "SyncNHTSARecallsCommand" }

type SyncNHTSARecallsCommandHandler struct {
	DBS         func() *db.ReaderWriter
	RecallsRepo repositories.DeviceNHTSARecallsRepository
	DDRepo      repositories.DeviceDefinitionRepository
	MakesRepo   repositories.DeviceMakeRepository
	FileURL     *string
}

func NewSyncNHTSARecallsCommandHandler(dbs func() *db.ReaderWriter, recallsRepo repositories.DeviceNHTSARecallsRepository, ddRepo repositories.DeviceDefinitionRepository, makesRepo repositories.DeviceMakeRepository, file *string) SyncNHTSARecallsCommandHandler {
	return SyncNHTSARecallsCommandHandler{DBS: dbs, RecallsRepo: recallsRepo, DDRepo: ddRepo, MakesRepo: makesRepo, FileURL: file}
}

const NHTSARecallsMatchingVersion = "2022.10.18.0"
const NHTSARecallsColumnCount = 27

type NHTSARecallMetadata struct {
	MatchingVersion string   `json:"matchingVersion,omitempty"`
	MatchType       string   `json:"matchType,omitempty"`
	MatchedMake     []string `json:"matchedMake,omitempty"`
	MatchedModel    []string `json:"matchedModel,omitempty"`
	AdditionalData  []string `json:"additionalData,omitempty"`
}

func (ch SyncNHTSARecallsCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	defer fmt.Println("Completed NHTSA Recalls sync")

	_ = query.(*SyncNHTSARecallsCommand)

	filePath, err := ch.DownloadFileToTemp("", *ch.FileURL)

	fmt.Printf("Tmp file: %s\n", *filePath)

	r, err := zip.OpenReader(*filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = r.Close()
		if err != nil {
			log.Fatal(err)
		}
		err = os.Remove(*filePath)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Removed tmp file")
	}()

	// Only one file is expected inside the ZIP file
	if len(r.File) != 1 {
		return nil, fmt.Errorf("found %d files in NHTSA Recalls ZIP file, expected 1 file", len(r.File))
	}

	rc, err := r.File[0].Open()
	if err != nil {
		return nil, err
	}
	defer func(rc io.ReadCloser) {
		err = rc.Close()
		if err != nil {
			panic(err)
		}
	}(rc)

	lastID, err := ch.RecallsRepo.GetLastDataRecordID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	expectID := 1
	insertCount := 0
	matchCount := 0

	fmt.Print("\rReading file...")

	// The file is expected to be a tab separated value (TSV) file without a header row
	scanLine := bufio.NewScanner(bufio.NewReader(rc))
	scanLine.Split(bufio.ScanLines)
	for scanLine.Scan() {
		scanFields := strings.Split(scanLine.Text(), "\t")
		id, err := strconv.Atoi(scanFields[0])
		if err != nil {
			return nil, err
		}
		if expectID != id {
			log.Printf("NHTSA Recall record ID is %d, expected %d\n", id, expectID)
		}

		mdJSON := null.JSON{}
		if len(scanFields) > NHTSARecallsColumnCount {
			//log.Printf("NHTSA Recall record ID %d has %d columns, expected %d\n", id, len(scanFields), NHTSARecallsColumnCount)
			//fmt.Println(scanFields[NHTSARecallsColumnCount:])
			md := NHTSARecallMetadata{
				AdditionalData: scanFields[NHTSARecallsColumnCount:],
			}
			medtadataJSON, err := json.Marshal(md)
			if err != nil {
				log.Fatal(err)
			}
			mdJSON = null.JSONFrom(medtadataJSON)
		}

		// Add to DB if last ID is less than this ID
		if lastID.IsZero() || lastID.Int < id {
			_, err = ch.RecallsRepo.Create(ctx, null.String{}, scanFields, mdJSON)
			if err != nil {
				log.Print(err)
				fmt.Println("Aborting...")
				break
			}
			fmt.Printf("\rInserted data record ID %d", id)
			insertCount++
		}

		expectID = id + 1
	}

	if lastID.IsZero() {
		fmt.Printf("\rInserted %d rows\n", insertCount)
	} else {
		fmt.Printf("\rInserted %d rows after data record ID %d\n", insertCount, lastID.Int)
	}

	fmt.Print("Finding matching device definitions...")

	matchedCount, err := ch.RecallsRepo.MatchDeviceDefinition(ctx, NHTSARecallsMatchingVersion)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println("\r...")
	//fmt.Print("\033[1A\033[K") // clear line
	fmt.Printf("\rProcessed %d records using device definition matching version %s\n", matchedCount, NHTSARecallsMatchingVersion)

	return SyncNHTSARecallsCommandResult{insertCount, matchCount}, nil
}

func (ch SyncNHTSARecallsCommandHandler) DownloadFileToTemp(filename string, url string) (localpath *string, err error) {

	// Default filename to same as the filename in the URL
	if filename == "" {
		r, _ := http.NewRequest("GET", url, nil)
		filename = path.Base(r.URL.Path)
		filenameArr := strings.Split(filename, ".")
		filenameArr[len(filenameArr)-1] = "*." + filenameArr[len(filenameArr)-1]
		filename = strings.Join(filenameArr, ".")
	}

	// Create the file
	out, err := os.CreateTemp("", filename)
	if err != nil {
		return nil, err
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	err = out.Close()
	if err != nil {
		return nil, err
	}

	outpath := out.Name()

	return &outpath, nil
}
