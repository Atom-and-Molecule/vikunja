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

package models

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// ProjectWikiPageAttachment represents a file attached to a wiki page.
type ProjectWikiPageAttachment struct {
	ID          int64      `xorm:"bigint autoincr not null unique pk" json:"id" param:"attachment" readOnly:"true" doc:"The unique, numeric id of this attachment."`
	PageID      int64      `xorm:"bigint not null index" json:"page_id" param:"wiki_page" readOnly:"true" doc:"The wiki page this attachment belongs to."`
	FileID      int64      `xorm:"bigint not null" json:"-"`
	CreatedByID int64      `xorm:"bigint not null" json:"-"`
	CreatedBy   *user.User `xorm:"-" json:"created_by" readOnly:"true" doc:"The user who uploaded this attachment."`
	File        *files.File `xorm:"-" json:"file" readOnly:"true" doc:"Metadata of the uploaded file."`
	Created     time.Time  `xorm:"created not null" json:"created" readOnly:"true" doc:"Timestamp when this attachment was uploaded."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*ProjectWikiPageAttachment) TableName() string {
	return "project_wiki_page_attachments"
}

// NewAttachment creates a new wiki page attachment.
func (pa *ProjectWikiPageAttachment) NewAttachment(s *xorm.Session, f io.ReadSeeker, realname string, realsize uint64, a web.Auth) error {
	file, err := files.CreateWithSession(s, f, realname, realsize, a)
	if err != nil {
		if files.IsErrFileIsTooLarge(err) {
			return ErrTaskAttachmentIsTooLarge{Size: realsize}
		}
		return err
	}
	pa.File = file
	pa.FileID = file.ID

	u, err := GetUserOrLinkShareUser(s, a)
	if err != nil {
		_ = file.Delete(s)
		return err
	}
	pa.CreatedBy = u
	pa.CreatedByID = u.ID

	_, err = s.Insert(pa)
	if err != nil {
		_ = file.Delete(s)
		return err
	}
	return nil
}

// UploadProjectWikiPageAttachments stores multiple uploaded files for a wiki page.
func UploadProjectWikiPageAttachments(s *xorm.Session, a web.Auth, pageID int64, uploads []*AttachmentToUpload) (success []*ProjectWikiPageAttachment, failures []error, err error) {
	pa := &ProjectWikiPageAttachment{PageID: pageID}
	canCreate, err := pa.CanCreate(s, a)
	if err != nil {
		return nil, nil, err
	}
	if !canCreate {
		return nil, nil, ErrGenericForbidden{}
	}

	for _, upload := range uploads {
		attachment := &ProjectWikiPageAttachment{PageID: pageID}
		err = attachment.NewAttachment(s, upload.Reader, upload.Filename, upload.Size, a)
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", upload.Filename, err))
			continue
		}
		success = append(success, attachment)
	}

	return success, failures, nil
}

// LoadProjectWikiPageAttachmentForDownload verifies access and loads the attachment for downloading or previewing.
func LoadProjectWikiPageAttachmentForDownload(s *xorm.Session, a web.Auth, pageID, attachmentID int64, previewSize PreviewSize) (pa *ProjectWikiPageAttachment, preview []byte, err error) {
	pa = &ProjectWikiPageAttachment{ID: attachmentID, PageID: pageID}
	canRead, _, err := pa.CanRead(s, a)
	if err != nil {
		return nil, nil, err
	}
	if !canRead {
		return nil, nil, ErrGenericForbidden{}
	}

	err = pa.ReadOne(s, a)
	if err != nil {
		return nil, nil, err
	}

	if previewSize != "" && pa.File != nil && pa.File.MimeType != "" && strings.HasPrefix(pa.File.MimeType, "image/") {
		preview = pa.GetPreview(previewSize)
	}

	return pa, preview, nil
}

// ReadAll returns attachments for a wiki page.
func (pa *ProjectWikiPageAttachment) ReadAll(s *xorm.Session, a web.Auth, _ string, page int, perPage int) (result any, resultCount int, totalCount int64, err error) {
	canRead, _, err := pa.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	q := s.Where("page_id = ?", pa.PageID)
	totalCount, err = q.Clone().Count(&ProjectWikiPageAttachment{})
	if err != nil {
		return nil, 0, 0, err
	}

	attachments := []*ProjectWikiPageAttachment{}
	q = q.OrderBy("id desc")
	if perPage > 0 && page > 0 {
		q = q.Limit(perPage, (page-1)*perPage)
	}

	err = q.Find(&attachments)
	if err != nil {
		return nil, 0, 0, err
	}

	for _, att := range attachments {
		if att.FileID > 0 {
			f := &files.File{ID: att.FileID}
			if err := f.ReadOne(s); err == nil {
				att.File = f
			}
		}
		if att.CreatedByID > 0 {
			att.CreatedBy, _ = user.GetUserByID(s, att.CreatedByID)
		}
	}

	return attachments, len(attachments), totalCount, nil
}

// ReadOne returns a single attachment.
func (pa *ProjectWikiPageAttachment) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	exists, err := s.Where("id = ? AND page_id = ?", pa.ID, pa.PageID).Get(pa)
	if err != nil {
		return err
	}
	if !exists {
		return &ErrProjectWikiPageAttachmentDoesNotExist{AttachmentID: pa.ID}
	}

	if pa.FileID > 0 {
		f := &files.File{ID: pa.FileID}
		if err := f.ReadOne(s); err == nil {
			pa.File = f
		}
	}
	if pa.CreatedByID > 0 {
		pa.CreatedBy, _ = user.GetUserByID(s, pa.CreatedByID)
	}
	return nil
}

// Delete removes an attachment and its underlying file.
func (pa *ProjectWikiPageAttachment) Delete(s *xorm.Session, _ web.Auth) (err error) {
	exists, err := s.Where("id = ? AND page_id = ?", pa.ID, pa.PageID).Get(pa)
	if err != nil {
		return err
	}
	if !exists {
		return &ErrProjectWikiPageAttachmentDoesNotExist{AttachmentID: pa.ID}
	}

	if pa.FileID > 0 {
		f := &files.File{ID: pa.FileID}
		_ = f.Delete(s)
	}
	_, err = s.Where("id = ?", pa.ID).Delete(&ProjectWikiPageAttachment{})
	return err
}

// CanRead checks read permission against the parent wiki page.
func (pa *ProjectWikiPageAttachment) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	if isInstanceAdmin(s, a) {
		return true, int(PermissionAdmin), nil
	}
	if pa.PageID == 0 && pa.ID > 0 {
		att := &ProjectWikiPageAttachment{ID: pa.ID}
		exists, err := s.ID(pa.ID).Get(att)
		if err != nil {
			return false, 0, err
		}
		if !exists {
			return false, 0, &ErrProjectWikiPageAttachmentDoesNotExist{AttachmentID: pa.ID}
		}
		pa.PageID = att.PageID
	}
	page := &ProjectWikiPage{ID: pa.PageID}
	return page.CanRead(s, a)
}

// CanCreate checks create access (requires write access to the page).
func (pa *ProjectWikiPageAttachment) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if isInstanceAdmin(s, a) {
		return true, nil
	}
	page := &ProjectWikiPage{ID: pa.PageID}
	return page.CanUpdate(s, a)
}

// CanDelete checks delete access (requires write access to the page).
func (pa *ProjectWikiPageAttachment) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	if isInstanceAdmin(s, a) {
		return true, nil
	}
	page := &ProjectWikiPage{ID: pa.PageID}
	return page.CanUpdate(s, a)
}

func cacheKeyForWikiAttachmentPreview(id int64, size PreviewSize) string {
	return "wiki_attachment_preview_" + strconv.FormatInt(id, 10) + "_size_" + string(size)
}

// GetPreview generates and returns a cached preview thumbnail of an image attachment.
func (pa *ProjectWikiPageAttachment) GetPreview(previewSize PreviewSize) []byte {
	cacheKey := cacheKeyForWikiAttachmentPreview(pa.ID, previewSize)

	result, err := keyvalue.RememberValue(cacheKey, func() ([]byte, error) {
		data, err := io.ReadAll(pa.File.File)
		if err != nil {
			return nil, err
		}

		const maxPixels = 50_000_000
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if cfg.Width*cfg.Height > maxPixels {
			return nil, fmt.Errorf("image dimensions %dx%d exceed maximum of %d pixels", cfg.Width, cfg.Height, maxPixels)
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		resizedImg := resizeImage(img, previewSize.GetSize())
		buf := &bytes.Buffer{}
		if err := png.Encode(buf, resizedImg); err != nil {
			return nil, err
		}
		return io.ReadAll(buf)
	})
	if err != nil {
		return nil
	}
	return result
}

// ErrProjectWikiPageAttachmentDoesNotExist represents a not found error.
type ErrProjectWikiPageAttachmentDoesNotExist struct {
	AttachmentID int64
}

// IsErrProjectWikiPageAttachmentDoesNotExist checks if error is ErrProjectWikiPageAttachmentDoesNotExist.
func IsErrProjectWikiPageAttachmentDoesNotExist(err error) bool {
	_, ok := err.(ErrProjectWikiPageAttachmentDoesNotExist)
	return ok
}

func (err ErrProjectWikiPageAttachmentDoesNotExist) Error() string {
	return fmt.Sprintf("Project wiki page attachment with id %d does not exist", err.AttachmentID)
}

// ErrCodeProjectWikiPageAttachmentDoesNotExist is the world error code.
const ErrCodeProjectWikiPageAttachmentDoesNotExist = 20004

// HTTPError implements web.HTTPErrorProcessor.
func (err ErrProjectWikiPageAttachmentDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeProjectWikiPageAttachmentDoesNotExist,
		Message:  "The project wiki page attachment does not exist.",
		Args:     web.Map{"attachment_id": err.AttachmentID},
	}
}
