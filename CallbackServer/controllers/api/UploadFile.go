////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq" // Needed for gorp
	"github.com/zenazn/goji/web"

	"qdserver/CallbackServer/command"
	"qdserver/CallbackServer/system"
	"qdserver/lib/models"
	"qdserver/lib/taskqueue"
	"qdserver/lib/utils"
)

// UploadFile receives a file from the client
func (controller *Controller) UploadFile(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)
	ch := controller.GetQueueChannel(c)

	type eventFromClient struct {
		SystemUUID        string
		CustomerUUID      string
		CurrentClientTime int64
		Sha256            string // Sha256 of the file
		FileType          string // May be "executable" or "catalog"
	}

	var event eventFromClient
	var systemID int64
	var fileUploaded bool

	var err error
	// TODO I should be smarter about my file uploads so I don't need to burn so much memory
	const maxSize = 1024 * 1024 * 100 // 100MB max
	if err = r.ParseMultipartForm(maxSize); nil != err {
		log.Errorf("Unable to read multi part form, %v", err)
		return "", http.StatusBadRequest
	}

	// Get event data associated with file upload
	for _, vheaders := range r.MultipartForm.Value {
		for _, vhdr := range vheaders {
			err = json.Unmarshal([]byte(vhdr), &event)
			if err != nil {
				log.Errorf("Unable to unmarshal json, %v", err)
				return "", http.StatusBadRequest
			}

			systemID, err = utils.GetSystemIDFromUUID(db, event.SystemUUID, event.CustomerUUID)
			if err != nil {
				log.Errorf("Unable to parse get System ID, %v", err)
				return "", http.StatusBadRequest
			}

			log.Infof("System ID: %d\n", systemID)
		}
	}

	// Ensure we actually got some event data in the looping above
	if systemID == 0 {
		log.Errorf("No event data associated with upload, %v", err)
		return "", http.StatusBadRequest
	}

	// TODO I shouldn't need all these loops as I'm only uploading one file at a time
	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			// Sanity check
			if fileUploaded {
				log.Errorf("Already uploaded one file, agent tried uploading multiple files: %v", err)
				return "", http.StatusBadRequest
			}

			// Open uploaded
			var infile multipart.File
			if infile, err = hdr.Open(); nil != err {
				log.Errorf("Unable to open file, %v", err)
				return "", http.StatusBadRequest
			}

			// Sanity check filename
			sha256HexString := hdr.Filename
			match, _ := regexp.MatchString("^[a-f0-9]{64}$", sha256HexString)
			if !match {
				log.Errorf("Incorrectly formatted file name, %v", sha256HexString)
				return "", http.StatusBadRequest
			}

			// Convert given sha256 to a byte array
			sha256ByteArray, err := hex.DecodeString(sha256HexString)
			if err != nil {
				log.Errorf("Unable to convert hex string to bytes, %v", err)
				return "", http.StatusBadRequest
			}

			// Create temp file to write to
			var outfile *os.File
			outfile, err = ioutil.TempFile(c.Env["Config"].(*system.Configuration).UploadPath, sha256HexString)
			if err != nil {
				log.Errorf("Unable to create temp file, %v", err)
				return "", http.StatusBadRequest
			}

			defer outfile.Close()
			// Only deleting file in case everything works, so not deferring the deletion here

			// Write file to disk
			var written int64
			if written, err = io.Copy(outfile, infile); nil != err {
				log.Errorf("Unable to write file to disk, %v", err)
				return "", http.StatusBadRequest
			}

			// Check SHA256
			sha256OfUpload, err := utils.Sha256File(outfile)
			if err != nil {
				log.Errorf("Unable to hash uploaded file, %v", err)
				return "", http.StatusBadRequest
			}
			if bytes.Compare(sha256OfUpload, sha256ByteArray) != 0 {
				log.Errorf("Hash of uploaded file did not match expected value, %v", err)
				return "", http.StatusBadRequest
			}

			//
			// Record in the DB that we have a copy
			//
			if event.FileType == "exe" {
				// Upload to S3
				err = utils.UploadToS3HashPath(c.Env["Config"].(*system.Configuration).Aws.ExeUpload, sha256HexString, outfile)
				if err != nil {
					// I can probably continue on, but for now I'll bail on error
					log.Errorf("Unable to upload to S3 due to %v", err)
					return "", http.StatusBadRequest
				}

				// Find the entry in the DB for this file
				var executableFile models.ExecutableFile
				err = db.SelectOne(&executableFile, `SELECT *
					FROM ExecutableFiles
					WHERE Sha256=:sha256`,
					map[string]interface{}{
						"sha256": sha256ByteArray,
					})
				if err != nil {
					log.Errorf("Unable to find executable (%s) in DB (agent uploaded a file we've never seen), %v", sha256HexString, err)
					return "", http.StatusBadRequest
				}

				// Update it to say we have a copy
				executableFile.UploadDate = utils.DBTimeNow()
				_, err = db.Update(&executableFile)
				if err != nil {
					log.Errorf("Can't update exe: %v", err)
					return "", http.StatusBadRequest
				}

				// Sanity check
				if written != int64(executableFile.Size) {
					log.Errorf("Uploaded file does not equal expectd size (wrote: %d, expected: %d)", written, int64(executableFile.Size))
					return "", http.StatusBadRequest
				}

				err = TaskQueue.CreateTaskToAnalyzeFile(ch, executableFile.ID)
				if err != nil {
					log.Errorf("Failed to create task for file %d, %v", executableFile.ID, err)
					return "", http.StatusBadRequest
				}
			} else if event.FileType == "catalog" {
				// Upload to S3
				err = utils.UploadToS3HashPath(c.Env["Config"].(*system.Configuration).Aws.CatalogUpload, sha256HexString, outfile)
				if err != nil {
					// I can probably continue on, but for now I'll bail on error
					log.Errorf("Unable to upload to S3 due to %v", err)
					return "", http.StatusBadRequest
				}

				// Find the entry in the DB for this file
				var catalogFile models.CatalogFile
				err = db.SelectOne(&catalogFile, `SELECT *
								FROM CatalogFiles
								WHERE Sha256=:sha256`,
					map[string]interface{}{
						"sha256": sha256ByteArray,
					})
				if err != nil {
					log.Errorf("Unable to find catalog (%s) in DB (agent uploaded a file we've never seen), %v", sha256HexString, err)
					return "", http.StatusBadRequest
				}

				// Update it to say we have a copy
				catalogFile.UploadDate = utils.DBTimeNow()
				_, err = db.Update(&catalogFile)
				if err != nil {
					log.Errorf("Can't update exe: %v", err)
					return "", http.StatusBadRequest
				}

				// Sanity check
				if written != int64(catalogFile.Size) {
					log.Errorf("Uploaded file does not equal expectd size (wrote: %d, expected: %d)", written, int64(catalogFile.Size))
					return "", http.StatusBadRequest
				}

				err = TaskQueue.CreateTaskToAnalyzeCatalog(ch, catalogFile.ID)
				if err != nil {
					log.Errorf("Failed to create task for file %d, %v", catalogFile.ID, err)
					return "", http.StatusBadRequest
				}
			} else {
				log.Errorf("Unknown file type %s, %v", event.FileType, err)
				return "", http.StatusBadRequest
			}

			// Everything worked, and the file is on S3, so delete our copy
			os.Remove(outfile.Name())
			fileUploaded = true
		}
	}

	if !fileUploaded {
		log.Errorf("No file uploaded: %v", err)
		return "", http.StatusBadRequest
	}

	return GenerateResponseToAgent(controller.GetDatabase(c), systemID, command.Nop())
}
