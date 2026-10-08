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

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type projectWikiPage20261008093500 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	ProjectID    int64     `xorm:"bigint not null index" json:"project_id"`
	ParentPageID int64     `xorm:"bigint not null default 0 index" json:"parent_page_id"`
	Title        string    `xorm:"varchar(250) not null" json:"title"`
	Content      string    `xorm:"longtext null" json:"content"`
	Position     float64   `xorm:"double not null default 0" json:"position"`
	IsHome       bool      `xorm:"not null default false index" json:"is_home"`
	CreatedByID  int64     `xorm:"bigint not null" json:"created_by_id"`
	UpdatedByID  int64     `xorm:"bigint not null" json:"updated_by_id"`
	Created      time.Time `xorm:"created not null" json:"created"`
	Updated      time.Time `xorm:"updated not null" json:"updated"`
}

func (projectWikiPage20261008093500) TableName() string {
	return "project_wiki_pages"
}

type projectWikiPageRevision20261008093500 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	PageID      int64     `xorm:"bigint not null index" json:"page_id"`
	Title       string    `xorm:"varchar(250) not null" json:"title"`
	Content     string    `xorm:"longtext null" json:"content"`
	CreatedByID int64     `xorm:"bigint not null" json:"created_by_id"`
	Created     time.Time `xorm:"created not null" json:"created"`
}

func (projectWikiPageRevision20261008093500) TableName() string {
	return "project_wiki_page_revisions"
}

type projectWikiPageAttachment20261008093500 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	PageID      int64     `xorm:"bigint not null index" json:"page_id"`
	FileID      int64     `xorm:"bigint not null" json:"file_id"`
	CreatedByID int64     `xorm:"bigint not null" json:"created_by_id"`
	Created     time.Time `xorm:"created not null" json:"created"`
}

func (projectWikiPageAttachment20261008093500) TableName() string {
	return "project_wiki_page_attachments"
}

type projectView20261008093500 struct {
	ID                      int64     `xorm:"autoincr not null unique pk"`
	Title                   string    `xorm:"varchar(255) not null"`
	ProjectID               int64     `xorm:"not null index"`
	ViewKind                int       `xorm:"not null"`
	Position                float64   `xorm:"double null"`
	BucketConfigurationMode int       `xorm:"default 0"`
	Created                 time.Time `xorm:"created not null"`
	Updated                 time.Time `xorm:"updated not null"`
}

func (projectView20261008093500) TableName() string {
	return "project_views"
}

type projects20261008093500 struct {
	ID              int64 `xorm:"autoincr not null unique pk"`
	ParentProjectID int64 `xorm:"bigint null index"`
}

func (projects20261008093500) TableName() string {
	return "projects"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261008093500",
		Description: "Add project_wiki_pages, project_wiki_page_revisions, project_wiki_page_attachments tables, and add overview view to all projects",
		Migrate: func(tx *xorm.Engine) error {
			if err := tx.Sync(&projectWikiPage20261008093500{}, &projectWikiPageRevision20261008093500{}, &projectWikiPageAttachment20261008093500{}); err != nil {
				return err
			}

			// Add Overview view at position 50 for all projects
			// that don't already have one.
			var allProjects []*projects20261008093500
			if err := tx.Where("id > 0").Find(&allProjects); err != nil {
				return err
			}

			for _, p := range allProjects {
				hasOverview, err := tx.Where("project_id = ? AND view_kind = 4", p.ID).Exist(&projectView20261008093500{})
				if err != nil {
					return err
				}
				if !hasOverview {
					view := &projectView20261008093500{
						Title:                   "Overview",
						ProjectID:               p.ID,
						ViewKind:                4, // ProjectViewKindOverview
						Position:                50,
						BucketConfigurationMode: 0,
						Created:                 time.Now(),
						Updated:                 time.Now(),
					}
					if _, err := tx.Insert(view); err != nil {
						return err
					}
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
