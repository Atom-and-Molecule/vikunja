// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package files

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/models"
)

// WikiAttachmentUploadResult is the outcome of a wiki page attachment upload.
type WikiAttachmentUploadResult struct {
	Errors  []AttachmentUploadError              `json:"errors" doc:"Per-file failures. A file that fails here does not fail the whole request; the others still upload."`
	Success []*models.ProjectWikiPageAttachment `json:"success" doc:"The attachments that were created successfully."`
}

// BuildWikiUploadResult turns domain function results into the wire DTO.
func BuildWikiUploadResult(success []*models.ProjectWikiPageAttachment, failures []error) *WikiAttachmentUploadResult {
	r := &WikiAttachmentUploadResult{Success: success}
	for _, err := range failures {
		r.Errors = append(r.Errors, toAttachmentUploadError(err))
	}
	return r
}

// WriteWikiAttachmentDownload streams the wiki page attachment or preview to the HTTP response.
func WriteWikiAttachmentDownload(w http.ResponseWriter, r *http.Request, pa *models.ProjectWikiPageAttachment, preview []byte) {
	defer func() { _ = pa.File.File.Close() }()

	if preview != nil {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(preview)))
		_, _ = w.Write(preview)
		return
	}

	WriteFileDownload(w, r, pa.File)
}
