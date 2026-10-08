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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectWikiPage_Permissions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// In fixtures, project 1 is owned by user 1
	owner := &user.User{ID: 1}
	// user 2 has no access to project 1
	unauthorizedUser := &user.User{ID: 2}

	page := &ProjectWikiPage{
		ProjectID: 1,
		Title:     "Main Docs",
		Content:   "<p>Hello world</p>",
	}

	canCreate, err := page.CanCreate(s, owner)
	require.NoError(t, err)
	assert.True(t, canCreate)

	canCreateUnauthorized, err := page.CanCreate(s, unauthorizedUser)
	require.NoError(t, err)
	assert.False(t, canCreateUnauthorized)

	err = page.Create(s, owner)
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	s2 := db.NewSession()
	defer s2.Close()

	canRead, _, err := page.CanRead(s2, owner)
	require.NoError(t, err)
	assert.True(t, canRead)

	canReadUnauthorized, _, err := page.CanRead(s2, unauthorizedUser)
	require.NoError(t, err)
	assert.False(t, canReadUnauthorized)

	canUpdate, err := page.CanUpdate(s2, owner)
	require.NoError(t, err)
	assert.True(t, canUpdate)

	canUpdateUnauthorized, err := page.CanUpdate(s2, unauthorizedUser)
	require.NoError(t, err)
	assert.False(t, canUpdateUnauthorized)

	canDelete, err := page.CanDelete(s2, owner)
	require.NoError(t, err)
	assert.True(t, canDelete)

	canDeleteUnauthorized, err := page.CanDelete(s2, unauthorizedUser)
	require.NoError(t, err)
	assert.False(t, canDeleteUnauthorized)
}

func TestProjectWikiPage_HomeAndPosition(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	usr := &user.User{ID: 1}

	page1 := &ProjectWikiPage{
		ProjectID: 1,
		Title:     "Page One",
		Content:   "First page",
	}
	err := page1.Create(s, usr)
	require.NoError(t, err)
	assert.True(t, page1.IsHome, "First page should automatically be home")
	assert.Equal(t, float64(100), page1.Position)

	page2 := &ProjectWikiPage{
		ProjectID: 1,
		Title:     "Page Two",
		Content:   "Second page",
	}
	err = page2.Create(s, usr)
	require.NoError(t, err)
	assert.False(t, page2.IsHome, "Second page should not automatically be home")
	assert.Equal(t, float64(200), page2.Position)

	// Set page2 as home
	page2.IsHome = true
	err = page2.Update(s, usr)
	require.NoError(t, err)

	p1Check := &ProjectWikiPage{ID: page1.ID, ProjectID: 1}
	err = p1Check.ReadOne(s, usr)
	require.NoError(t, err)
	assert.False(t, p1Check.IsHome, "Page one should no longer be home")

	p2Check := &ProjectWikiPage{ID: page2.ID, ProjectID: 1}
	err = p2Check.ReadOne(s, usr)
	require.NoError(t, err)
	assert.True(t, p2Check.IsHome, "Page two should now be home")
}

func TestProjectWikiPage_CycleDetectionAndReparent(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	usr := &user.User{ID: 1}

	root := &ProjectWikiPage{ProjectID: 1, Title: "Root"}
	require.NoError(t, root.Create(s, usr))

	child := &ProjectWikiPage{ProjectID: 1, Title: "Child", ParentPageID: root.ID}
	require.NoError(t, child.Create(s, usr))

	grandchild := &ProjectWikiPage{ProjectID: 1, Title: "Grandchild", ParentPageID: child.ID}
	require.NoError(t, grandchild.Create(s, usr))

	// Self-parenting attempt
	root.ParentPageID = root.ID
	err := root.Update(s, usr)
	assert.Error(t, err)
	assert.True(t, IsErrProjectWikiPageInvalidParent(err))

	// Indirect cycle attempt: make root parented to grandchild
	root.ParentPageID = grandchild.ID
	err = root.Update(s, usr)
	assert.Error(t, err)
	assert.True(t, IsErrProjectWikiPageInvalidParent(err))

	// Deleting child should re-parent grandchild to root
	require.NoError(t, child.Delete(s, usr))

	gcCheck := &ProjectWikiPage{ID: grandchild.ID, ProjectID: 1}
	require.NoError(t, gcCheck.ReadOne(s, usr))
	assert.Equal(t, root.ID, gcCheck.ParentPageID, "Grandchild should have been moved up to root")
}

func TestProjectWikiPage_Revisions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	owner := &user.User{ID: 1}
	unauthorizedUser := &user.User{ID: 2}

	page := &ProjectWikiPage{
		ProjectID: 1,
		Title:     "Version 1 Title",
		Content:   "Version 1 Content",
	}
	require.NoError(t, page.Create(s, owner))

	// Update page to create revision 1
	page.Title = "Version 2 Title"
	page.Content = "Version 2 Content"
	require.NoError(t, page.Update(s, owner))

	// Update page to create revision 2
	page.Title = "Version 3 Title"
	page.Content = "Version 3 Content"
	require.NoError(t, page.Update(s, owner))

	// List revisions
	revObj := &ProjectWikiPageRevision{PageID: page.ID}
	res, count, total, err := revObj.ReadAll(s, owner, "", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, int64(2), total)

	revs, ok := res.([]*ProjectWikiPageRevision)
	require.True(t, ok)
	require.Len(t, revs, 2)

	// Newest revision first: revs[0] has Version 2, revs[1] has Version 1
	assert.Equal(t, "Version 2 Title", revs[0].Title)
	assert.Equal(t, "Version 2 Content", revs[0].Content)
	assert.Equal(t, "Version 1 Title", revs[1].Title)
	assert.Equal(t, "Version 1 Content", revs[1].Content)

	// Read single revision
	single := &ProjectWikiPageRevision{ID: revs[1].ID, PageID: page.ID}
	require.NoError(t, single.ReadOne(s, owner))
	assert.Equal(t, "Version 1 Title", single.Title)

	// Unauthorized user cannot read revisions
	_, _, _, err = revObj.ReadAll(s, unauthorizedUser, "", 0, 0)
	assert.Error(t, err)
	assert.True(t, IsErrGenericForbidden(err))

	canReadUnauthorized, _, err := single.CanRead(s, unauthorizedUser)
	require.NoError(t, err)
	assert.False(t, canReadUnauthorized)

	// Deleting page removes revisions
	require.NoError(t, page.Delete(s, owner))
	_, countAfterDelete, _, err := revObj.ReadAll(s, owner, "", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, countAfterDelete)
}

func TestProjectWikiPage_Attachments(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	files.InitTestFileFixtures(t)
	s := db.NewSession()
	defer s.Close()

	owner := &user.User{ID: 1}
	unauthorizedUser := &user.User{ID: 2}

	page := &ProjectWikiPage{
		ProjectID: 1,
		Title:     "Page With Attachments",
		Content:   "Testing file attachments",
	}
	require.NoError(t, page.Create(s, owner))
	require.NoError(t, s.Commit())

	s2 := db.NewSession()
	defer s2.Close()

	att := &ProjectWikiPageAttachment{
		PageID: page.ID,
	}

	// Permission checks
	canCreate, err := att.CanCreate(s2, owner)
	require.NoError(t, err)
	assert.True(t, canCreate)

	canCreateUnauthorized, err := att.CanCreate(s2, unauthorizedUser)
	require.NoError(t, err)
	assert.False(t, canCreateUnauthorized)

	// Create attachment
	content := []byte("wiki document attachment content")
	err = att.NewAttachment(s2, bytes.NewReader(content), "doc.txt", uint64(len(content)), owner)
	require.NoError(t, err)
	assert.True(t, att.ID > 0)
	assert.True(t, att.FileID > 0)
	require.NoError(t, s2.Commit())

	s3 := db.NewSession()
	defer s3.Close()

	// Read permissions
	attRead := &ProjectWikiPageAttachment{ID: att.ID, PageID: page.ID}
	canRead, _, err := attRead.CanRead(s3, owner)
	require.NoError(t, err)
	assert.True(t, canRead)

	canReadUnauthorized, _, err := attRead.CanRead(s3, unauthorizedUser)
	require.NoError(t, err)
	assert.False(t, canReadUnauthorized)

	// ReadAll
	listAtt := &ProjectWikiPageAttachment{PageID: page.ID}
	res, count, total, err := listAtt.ReadAll(s3, owner, "", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, int64(1), total)

	items, ok := res.([]*ProjectWikiPageAttachment)
	require.True(t, ok)
	require.Len(t, items, 1)
	assert.Equal(t, att.ID, items[0].ID)
	assert.NotNil(t, items[0].File)
	assert.Equal(t, "doc.txt", items[0].File.Name)

	// ReadAll forbidden
	_, _, _, err = listAtt.ReadAll(s3, unauthorizedUser, "", 0, 0)
	assert.Error(t, err)
	assert.True(t, IsErrGenericForbidden(err))

	// ReadOne
	single := &ProjectWikiPageAttachment{ID: att.ID, PageID: page.ID}
	require.NoError(t, single.ReadOne(s3, owner))
	assert.Equal(t, att.ID, single.ID)
	assert.Equal(t, "doc.txt", single.File.Name)

	// Load for download
	loaded, _, err := LoadProjectWikiPageAttachmentForDownload(s3, owner, page.ID, att.ID, "")
	require.NoError(t, err)
	assert.Equal(t, att.ID, loaded.ID)

	_, _, err = LoadProjectWikiPageAttachmentForDownload(s3, unauthorizedUser, page.ID, att.ID, "")
	assert.Error(t, err)

	// Delete attachment
	require.NoError(t, single.Delete(s3, owner))
	err = single.ReadOne(s3, owner)
	assert.Error(t, err)
	assert.True(t, IsErrProjectWikiPageAttachmentDoesNotExist(err))

	// Create another attachment and verify cascade on page delete
	att2 := &ProjectWikiPageAttachment{PageID: page.ID}
	err = att2.NewAttachment(s3, bytes.NewReader(content), "doc2.txt", uint64(len(content)), owner)
	require.NoError(t, err)
	require.NoError(t, s3.Commit())

	s4 := db.NewSession()
	defer s4.Close()

	require.NoError(t, page.Delete(s4, owner))
	_, countAfterPageDelete, _, err := listAtt.ReadAll(s4, owner, "", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, countAfterPageDelete)
}


