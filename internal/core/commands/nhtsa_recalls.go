package commands

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/null/v8"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
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

const NHTSARecallsMatchingVersion = "2022.10.06.0"
const NHTSARecallsColumnCount = 27

type NHTSARecallMetadata struct {
	MatchingVersion string   `json:"matchingVersion,omitempty"`
	MatchType       string   `json:"matchType,omitempty"`
	AdditionalData  []string `json:"additionalData,omitempty"`
}

func (ch SyncNHTSARecallsCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	_ = query.(*SyncNHTSARecallsCommand)

	filePath, err := ch.DownloadFileToTemp("", *ch.FileURL)

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

	fmt.Print("...")

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

	fmt.Print("...")

	//allDD, err := ch.DDRepo.GetAll(ctx, true)
	//if err != nil {
	//	log.Fatal(err)
	//}
	allMakes, err := ch.MakesRepo.GetAll(ctx)
	if err != nil {
		log.Fatal(err)
	}
	allUnmatched, err := ch.RecallsRepo.GetAllWithoutDD(ctx, NHTSARecallsMatchingVersion)
	if err != nil {
		log.Fatal(err)
	}

	for _, recall := range *allUnmatched {
		recallMakeUC := strings.ToUpper(recall.DataMaketxt)
		makeMatch := models.DeviceMake{}
		makeMatchType := "NONE"
		//wordMakes := []*models.DeviceMake{}
		// match make
		for _, mk := range allMakes {
			recallMake := recallMakeUC
			makeName := strings.ToUpper(mk.Name)
			if recallMake == makeName {
				// exact match
				makeMatch = *mk
				makeMatchType = "EXACT"
				break
			}
			re1 := regexp.MustCompile(`[_\W]+`)               // match non-alphanumerics
			recallMake = re1.ReplaceAllString(recallMake, "") // remove non-alphanumerics
			makeName = re1.ReplaceAllString(makeName, "")     // remove non-alphanumerics
			if recallMake == makeName {
				// exact match (alphanumerics only)
				makeMatch = *mk
				makeMatchType = "ALPHANUM"
				break
			}
			//// WIP
			//if strings.ContainsAny(recallMakeUC, " -") {
			//	recallMakeWords := re1.Split(recallMakeUC, -1)
			//	makeNameWords := re1.Split(strings.ToUpper(mk.Name), -1)
			//	wordsTotal := len(recallMakeWords)
			//	wordsMatched := 0
			//	for _, word := range recallMakeWords {
			//		for _, word2 := range makeNameWords {
			//			if word == word2 {
			//				wordsMatched++
			//				break
			//			}
			//		}
			//	}
			//	if float32(wordsMatched)/float32(wordsTotal) >= 0.5 {
			//		wordMakes = append(wordMakes, mk)
			//	}
			//}
		}

		// TODO: Improve matching? Takes some time
		if len(makeMatch.ID) > 0 {
			dd, err := ch.DDRepo.GetByMakeModelAndYears(ctx, makeMatch.Name, recall.DataModeltxt, recall.DataYeartxt, false)
			if err != nil {
				log.Fatal(err)
			}
			if dd != nil {
				metadata := NHTSARecallMetadata{}
				if recall.Metadata.Valid {
					err := json.Unmarshal(recall.Metadata.JSON, metadata)
					if err != nil {
						log.Fatal(err)
					}
				}
				metadata.MatchingVersion = NHTSARecallsMatchingVersion
				metadata.MatchType = makeMatchType
				metadataJSON, err := json.Marshal(metadata)
				if err != nil {
					log.Fatal(err)
				}
				mdJSON := null.JSONFrom(metadataJSON)
				err = ch.RecallsRepo.SetDDAndMetadata(ctx, *recall, &dd.ID, &mdJSON) // TODO: Needs more testing
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("\rMatched data record ID %d with DD ID %s", recall.DataRecordID, dd.ID)
				matchCount++
			}
			//continue
		}

		//// WIP
		//for i, wm := range wordMakes {
		//	fmt.Println("word", i, recall.DataMaketxt, wm.Name)
		//}
	}

	fmt.Println("\r...")
	fmt.Print("\033[1A\033[K") // clear line
	fmt.Println("\rFinished matching")

	fmt.Println("Completed NHTSA Recalls sync")

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
