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

type projectView20261008135000 struct {
	ID                      int64     `xorm:"autoincr not null unique pk"`
	Title                   string    `xorm:"varchar(255) not null"`
	ProjectID               int64     `xorm:"not null index"`
	ViewKind                int       `xorm:"not null"`
	Position                float64   `xorm:"double null"`
	BucketConfigurationMode int       `xorm:"default 0"`
	Created                 time.Time `xorm:"created not null"`
	Updated                 time.Time `xorm:"updated not null"`
}

func (projectView20261008135000) TableName() string {
	return "project_views"
}

type projects20261008135000 struct {
	ID int64 `xorm:"autoincr not null unique pk"`
}

func (projects20261008135000) TableName() string {
	return "projects"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261008135000",
		Description: "Add overview view to all projects that do not already have one",
		Migrate: func(tx *xorm.Engine) error {
			var projects []*projects20261008135000
			if err := tx.Where("id > 0").Find(&projects); err != nil {
				return err
			}

			for _, p := range projects {
				hasOverview, err := tx.Where("project_id = ? AND view_kind = 4", p.ID).Exist(&projectView20261008135000{})
				if err != nil {
					return err
				}
				if !hasOverview {
					view := &projectView20261008135000{
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
