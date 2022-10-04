package commands

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
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
	InsertedCount int32 `json:"insertedCount"`
}

func (*SyncNHTSARecallsCommand) Key() string { return "SyncNHTSARecallsCommand" }

type SyncNHTSARecallsCommandHandler struct {
	DBS     func() *db.ReaderWriter
	FileURL *string
}

func NewSyncNHTSARecallsCommandHandler(dbs func() *db.ReaderWriter, file *string) SyncNHTSARecallsCommandHandler {
	return SyncNHTSARecallsCommandHandler{DBS: dbs, FileURL: file}
}

func (ch SyncNHTSARecallsCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

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

	expectID := 1

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
			// TODO(zavaboy): Use logger here?
			fmt.Printf("NHTSA Recall record ID is %d, expected %d\n", id, expectID)
		}
		if len(scanFields) > 27 {
			// TODO(zavaboy): Use logger here?
			//fmt.Printf("NHTSA Recall record ID %d has %d columns, expected %d\n", id, len(scanFields), 27)
			//fmt.Println(scanFields[27:])
		}

		// TODO(zavaboy): Match to DD

		// TODO(zavaboy): Add to DB

		expectID = id + 1
	}

	fmt.Println("completed NHTSA Recalls sync")

	return SyncNHTSARecallsCommandResult{0}, nil
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
