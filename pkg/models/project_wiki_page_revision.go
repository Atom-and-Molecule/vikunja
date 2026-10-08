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
	"fmt"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// ProjectWikiPageRevision represents a historical revision of a wiki page.
type ProjectWikiPageRevision struct {
	// The unique, numeric id of this revision.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"revision" readOnly:"true" doc:"The unique, numeric id of this revision."`
	// The wiki page this revision belongs to.
	PageID int64 `xorm:"bigint not null index" json:"page_id" param:"wiki_page" readOnly:"true" doc:"The wiki page this revision belongs to."`
	// The title of the page at this revision.
	Title string `xorm:"varchar(250) not null" json:"title" readOnly:"true" doc:"The title of the page at this revision."`
	// The content of the page at this revision.
	Content string `xorm:"longtext null" json:"content" readOnly:"true" doc:"The content of the page at this revision."`

	CreatedByID int64      `xorm:"bigint not null" json:"-"`
	CreatedBy   *user.User `xorm:"-" json:"created_by" readOnly:"true" doc:"The user who authored this revision."`

	// A timestamp when this revision was recorded.
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this revision was recorded."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*ProjectWikiPageRevision) TableName() string {
	return "project_wiki_page_revisions"
}

// ReadAll returns all revisions for a page, ordered newest first.
func (r *ProjectWikiPageRevision) ReadAll(s *xorm.Session, a web.Auth, _ string, page int, perPage int) (result any, resultCount int, totalCount int64, err error) {
	canRead, _, err := r.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	q := s.Where("page_id = ?", r.PageID)
	if limit > 0 {
		q = q.Limit(limit, start)
	}

	revisions := []*ProjectWikiPageRevision{}
	err = q.OrderBy("id desc").Find(&revisions)
	if err != nil {
		return nil, 0, 0, err
	}

	for _, rev := range revisions {
		if rev.CreatedByID > 0 {
			rev.CreatedBy, _ = user.GetUserByID(s, rev.CreatedByID)
		}
	}

	totalCount, err = s.Where("page_id = ?", r.PageID).Count(&ProjectWikiPageRevision{})
	if err != nil {
		return nil, 0, 0, err
	}

	return revisions, len(revisions), totalCount, nil
}

// ReadOne returns a single revision.
func (r *ProjectWikiPageRevision) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	exists, err := s.Where("id = ? AND page_id = ?", r.ID, r.PageID).Get(r)
	if err != nil {
		return err
	}
	if !exists {
		return &ErrProjectWikiPageRevisionDoesNotExist{RevisionID: r.ID}
	}

	if r.CreatedByID > 0 {
		r.CreatedBy, _ = user.GetUserByID(s, r.CreatedByID)
	}

	return nil
}

// CanRead checks read permission against the parent wiki page.
func (r *ProjectWikiPageRevision) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	if isInstanceAdmin(s, a) {
		return true, int(PermissionAdmin), nil
	}
	if r.PageID == 0 && r.ID > 0 {
		rev := &ProjectWikiPageRevision{ID: r.ID}
		exists, err := s.ID(r.ID).Get(rev)
		if err != nil {
			return false, 0, err
		}
		if !exists {
			return false, 0, &ErrProjectWikiPageRevisionDoesNotExist{RevisionID: r.ID}
		}
		r.PageID = rev.PageID
	}
	page := &ProjectWikiPage{ID: r.PageID}
	return page.CanRead(s, a)
}

// ErrProjectWikiPageRevisionDoesNotExist represents an error when a revision is not found.
type ErrProjectWikiPageRevisionDoesNotExist struct {
	RevisionID int64
}

// IsErrProjectWikiPageRevisionDoesNotExist checks if an error is ErrProjectWikiPageRevisionDoesNotExist.
func IsErrProjectWikiPageRevisionDoesNotExist(err error) bool {
	_, ok := err.(ErrProjectWikiPageRevisionDoesNotExist)
	return ok
}

func (err ErrProjectWikiPageRevisionDoesNotExist) Error() string {
	return fmt.Sprintf("Project wiki page revision with id %d does not exist", err.RevisionID)
}

// ErrCodeProjectWikiPageRevisionDoesNotExist holds the unique world-error code of this error.
const ErrCodeProjectWikiPageRevisionDoesNotExist = 20003

// HTTPError holds the http error description.
func (err ErrProjectWikiPageRevisionDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeProjectWikiPageRevisionDoesNotExist,
		Message:  "The project wiki page revision does not exist.",
	}
}
