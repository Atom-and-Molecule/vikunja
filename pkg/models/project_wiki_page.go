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
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// ProjectWikiPage represents a wiki page within a project.
type ProjectWikiPage struct {
	// The unique, numeric id of this wiki page.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"wiki_page" readOnly:"true" doc:"The unique, numeric id of this wiki page."`
	// The project this wiki page belongs to. Taken from the URL path; ignored on write.
	ProjectID int64 `xorm:"bigint not null index" json:"project_id" param:"project" readOnly:"true" doc:"The project this wiki page belongs to."`
	// The parent wiki page id for nested pages. 0 indicates a root-level page.
	ParentPageID int64 `xorm:"bigint not null default 0 index" json:"parent_page_id" doc:"The parent wiki page id for nested pages. 0 indicates a root page."`
	// The title of the wiki page.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250" doc:"The title of the wiki page."`
	// The HTML content of the wiki page.
	Content string `xorm:"longtext null" json:"content" doc:"The content of the wiki page."`
	// The position of this page among its siblings for ordering.
	Position float64 `xorm:"double not null default 0" json:"position" doc:"The position of this page among its siblings for ordering."`
	// Whether this page is the primary home page for the project wiki.
	IsHome bool `xorm:"not null default false index" json:"is_home" doc:"Whether this page is the primary home page for the project wiki."`

	CreatedByID int64      `xorm:"bigint not null" json:"-"`
	CreatedBy   *user.User `xorm:"-" json:"created_by" readOnly:"true" doc:"The user who created this wiki page."`
	UpdatedByID int64      `xorm:"bigint not null" json:"-"`
	UpdatedBy   *user.User `xorm:"-" json:"updated_by" readOnly:"true" doc:"The user who last updated this wiki page."`

	// A timestamp when this wiki page was created.
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this wiki page was created."`
	// A timestamp when this wiki page was last updated.
	Updated time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"A timestamp when this wiki page was last updated."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*ProjectWikiPage) TableName() string {
	return "project_wiki_pages"
}

// ReadAll gets all wiki pages for a project.
func (wp *ProjectWikiPage) ReadAll(s *xorm.Session, _ web.Auth, search string, page int, perPage int) (result any, resultCount int, totalCount int64, err error) {
	q := s.Where("project_id = ?", wp.ProjectID)
	if search != "" {
		q = q.Where("title LIKE ?", "%"+search+"%")
	}

	totalCount, err = q.Clone().Count(&ProjectWikiPage{})
	if err != nil {
		return nil, 0, 0, err
	}

	pages := []*ProjectWikiPage{}
	q = q.OrderBy("parent_page_id asc, position asc, id asc")
	if perPage > 0 && page > 0 {
		q = q.Limit(perPage, (page-1)*perPage)
	}

	err = q.Find(&pages)
	if err != nil {
		return nil, 0, 0, err
	}

	for _, p := range pages {
		if p.CreatedByID > 0 {
			p.CreatedBy, _ = user.GetUserByID(s, p.CreatedByID)
		}
		if p.UpdatedByID > 0 {
			p.UpdatedBy, _ = user.GetUserByID(s, p.UpdatedByID)
		}
	}

	return pages, len(pages), totalCount, nil
}

// ReadOne gets a single wiki page.
func (wp *ProjectWikiPage) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	exists, err := s.Where("id = ? AND project_id = ?", wp.ID, wp.ProjectID).Get(wp)
	if err != nil {
		return err
	}
	if !exists {
		return &ErrProjectWikiPageDoesNotExist{WikiPageID: wp.ID}
	}

	if wp.CreatedByID > 0 {
		wp.CreatedBy, _ = user.GetUserByID(s, wp.CreatedByID)
	}
	if wp.UpdatedByID > 0 {
		wp.UpdatedBy, _ = user.GetUserByID(s, wp.UpdatedByID)
	}

	return nil
}

// Create creates a new wiki page in a project.
func (wp *ProjectWikiPage) Create(s *xorm.Session, a web.Auth) (err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}
	wp.CreatedByID = u.ID
	wp.UpdatedByID = u.ID

	if wp.ParentPageID > 0 {
		parent := &ProjectWikiPage{ID: wp.ParentPageID, ProjectID: wp.ProjectID}
		exists, err := s.Where("id = ? AND project_id = ?", parent.ID, parent.ProjectID).Get(parent)
		if err != nil {
			return err
		}
		if !exists {
			return &ErrProjectWikiPageInvalidParent{ParentPageID: wp.ParentPageID}
		}
	}

	count, err := s.Where("project_id = ?", wp.ProjectID).Count(&ProjectWikiPage{})
	if err != nil {
		return err
	}
	if count == 0 {
		wp.IsHome = true
	} else if wp.IsHome {
		_, err = s.Where("project_id = ?", wp.ProjectID).Cols("is_home").Update(&ProjectWikiPage{IsHome: false})
		if err != nil {
			return err
		}
	}

	if wp.Position == 0 {
		var maxPos float64
		_, err = s.Where("project_id = ? AND parent_page_id = ?", wp.ProjectID, wp.ParentPageID).
			Select("COALESCE(MAX(position), 0)").
			Table(&ProjectWikiPage{}).
			Get(&maxPos)
		if err != nil {
			return err
		}
		wp.Position = maxPos + 100
	}

	_, err = s.Insert(wp)
	if err != nil {
		return err
	}

	wp.CreatedBy = u
	wp.UpdatedBy = u
	return nil
}

// Update updates an existing wiki page.
func (wp *ProjectWikiPage) Update(s *xorm.Session, a web.Auth) (err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}
	wp.UpdatedByID = u.ID

	existing := &ProjectWikiPage{ID: wp.ID, ProjectID: wp.ProjectID}
	exists, err := s.Where("id = ? AND project_id = ?", existing.ID, existing.ProjectID).Get(existing)
	if err != nil {
		return err
	}
	if !exists {
		return &ErrProjectWikiPageDoesNotExist{WikiPageID: wp.ID}
	}

	if wp.ParentPageID > 0 {
		if wp.ParentPageID == wp.ID {
			return &ErrProjectWikiPageInvalidParent{ParentPageID: wp.ParentPageID}
		}

		// Cycle detection: ensure wp.ID is not an ancestor of wp.ParentPageID
		currParentID := wp.ParentPageID
		for currParentID > 0 {
			if currParentID == wp.ID {
				return &ErrProjectWikiPageInvalidParent{ParentPageID: wp.ParentPageID}
			}
			ancestor := &ProjectWikiPage{ID: currParentID, ProjectID: wp.ProjectID}
			exists, err := s.Where("id = ? AND project_id = ?", ancestor.ID, ancestor.ProjectID).Get(ancestor)
			if err != nil {
				return err
			}
			if !exists {
				return &ErrProjectWikiPageInvalidParent{ParentPageID: wp.ParentPageID}
			}
			currParentID = ancestor.ParentPageID
		}
	}

	if wp.IsHome && !existing.IsHome {
		_, err = s.Where("project_id = ? AND id != ?", wp.ProjectID, wp.ID).Cols("is_home").Update(&ProjectWikiPage{IsHome: false})
		if err != nil {
			return err
		}
	}

	_, err = s.ID(wp.ID).Cols("parent_page_id", "title", "content", "position", "is_home", "updated_by_id").Update(wp)
	return err
}

// Delete deletes a wiki page and moves its direct children up to its parent.
func (wp *ProjectWikiPage) Delete(s *xorm.Session, _ web.Auth) (err error) {
	existing := &ProjectWikiPage{ID: wp.ID, ProjectID: wp.ProjectID}
	exists, err := s.Where("id = ? AND project_id = ?", existing.ID, existing.ProjectID).Get(existing)
	if err != nil {
		return err
	}
	if !exists {
		return &ErrProjectWikiPageDoesNotExist{WikiPageID: wp.ID}
	}

	// Move children up to deleted page's parent
	_, err = s.Where("project_id = ? AND parent_page_id = ?", wp.ProjectID, wp.ID).
		Cols("parent_page_id").
		Update(&ProjectWikiPage{ParentPageID: existing.ParentPageID})
	if err != nil {
		return err
	}

	_, err = s.Where("id = ? AND project_id = ?", wp.ID, wp.ProjectID).Delete(&ProjectWikiPage{})
	if err != nil {
		return err
	}

	if existing.IsHome {
		var nextHome ProjectWikiPage
		hasNext, err := s.Where("project_id = ? AND parent_page_id = 0", wp.ProjectID).
			OrderBy("position asc, id asc").
			Get(&nextHome)
		if err != nil {
			return err
		}
		if hasNext {
			nextHome.IsHome = true
			_, err = s.ID(nextHome.ID).Cols("is_home").Update(&nextHome)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
