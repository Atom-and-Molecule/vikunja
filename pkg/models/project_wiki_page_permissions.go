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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

func (wp *ProjectWikiPage) resolveProjectID(s *xorm.Session) error {
	if wp.ProjectID > 0 || wp.ID == 0 {
		return nil
	}
	loaded := &ProjectWikiPage{ID: wp.ID}
	exists, err := s.Where("id = ?", wp.ID).Cols("project_id").Get(loaded)
	if err != nil {
		return err
	}
	if exists {
		wp.ProjectID = loaded.ProjectID
	}
	return nil
}

// CanRead checks whether the authenticated user has read permissions on the wiki page.
func (wp *ProjectWikiPage) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	if isInstanceAdmin(s, a) {
		return true, int(PermissionAdmin), nil
	}
	if err := wp.resolveProjectID(s); err != nil {
		return false, 0, err
	}
	p := &Project{ID: wp.ProjectID}
	return p.CanRead(s, a)
}

// CanCreate checks whether the authenticated user has write permissions to create wiki pages in the project.
func (wp *ProjectWikiPage) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if isInstanceAdmin(s, a) {
		return true, nil
	}
	if err := wp.resolveProjectID(s); err != nil {
		return false, err
	}
	p := &Project{ID: wp.ProjectID}
	return p.CanWrite(s, a)
}

// CanUpdate checks whether the authenticated user has write permissions to update the wiki page.
func (wp *ProjectWikiPage) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	if isInstanceAdmin(s, a) {
		return true, nil
	}
	if err := wp.resolveProjectID(s); err != nil {
		return false, err
	}
	p := &Project{ID: wp.ProjectID}
	return p.CanWrite(s, a)
}

// CanDelete checks whether the authenticated user has write permissions to delete the wiki page.
func (wp *ProjectWikiPage) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	if isInstanceAdmin(s, a) {
		return true, nil
	}
	if err := wp.resolveProjectID(s); err != nil {
		return false, err
	}
	p := &Project{ID: wp.ProjectID}
	return p.CanWrite(s, a)
}
