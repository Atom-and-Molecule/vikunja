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

package webtests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectWikiPage(t *testing.T) {
	owned := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/1/wiki/pages",
		idParam:  "page",
		t:        t,
	}
	require.NoError(t, owned.ensureEnv())

	forbidden := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/2/wiki/pages",
		idParam:  "page",
		t:        t,
		echo:     owned.echo,
	}

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testCreate(`{"title":"Getting Started","content":"<p>Welcome</p>"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)

			var body struct {
				ID      int64  `json:"id"`
				Title   string `json:"title"`
				Content string `json:"content"`
				IsHome  bool   `json:"is_home"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.True(t, body.ID > 0)
			assert.Equal(t, "Getting Started", body.Title)
			assert.Equal(t, "<p>Welcome</p>", body.Content)
			assert.True(t, body.IsHome, "First created wiki page should be home")
		})

		t.Run("Forbidden project", func(t *testing.T) {
			rec, err := forbidden.testCreate(`{"title":"Forbidden Page"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})

		t.Run("Forbidden project", func(t *testing.T) {
			rec, err := forbidden.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		recCreate, err := owned.testCreate(`{"title":"Read Me","content":"Detail"}`)
		require.NoError(t, err)
		var created struct {
			ID int64 `json:"id"`
		}
		require.NoError(t, json.Unmarshal(recCreate.Body.Bytes(), &created))

		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testReadOne(created.ID)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var readBody struct {
				ID            int64  `json:"id"`
				Title         string `json:"title"`
				MaxPermission int    `json:"max_permission"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &readBody))
			assert.Equal(t, created.ID, readBody.ID)
			assert.Equal(t, "Read Me", readBody.Title)
		})

		t.Run("Nonexistent", func(t *testing.T) {
			rec, err := owned.testReadOne(999999)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	})

	t.Run("Update", func(t *testing.T) {
		recCreate, err := owned.testCreate(`{"title":"Before Update","content":"Old"}`)
		require.NoError(t, err)
		var created struct {
			ID int64 `json:"id"`
		}
		require.NoError(t, json.Unmarshal(recCreate.Body.Bytes(), &created))

		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testUpdate(created.ID, fmt.Sprintf(`{"id":%d,"title":"After Update","content":"New"}`, created.ID))
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var updated struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
			assert.Equal(t, "After Update", updated.Title)
			assert.Equal(t, "New", updated.Content)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		recCreate, err := owned.testCreate(`{"title":"To Delete","content":"Bye"}`)
		require.NoError(t, err)
		var created struct {
			ID int64 `json:"id"`
		}
		require.NoError(t, json.Unmarshal(recCreate.Body.Bytes(), &created))

		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testDelete(created.ID)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)

			recGet, err := owned.testReadOne(created.ID)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, recGet.Code)
		})
	})

	t.Run("Revisions", func(t *testing.T) {
		recCreate, err := owned.testCreate(`{"title":"Initial Title","content":"Initial Content"}`)
		require.NoError(t, err)
		var created struct {
			ID int64 `json:"id"`
		}
		require.NoError(t, json.Unmarshal(recCreate.Body.Bytes(), &created))

		_, err = owned.testUpdate(created.ID, fmt.Sprintf(`{"id":%d,"title":"Second Title","content":"Second Content"}`, created.ID))
		require.NoError(t, err)
		_, err = owned.testUpdate(created.ID, fmt.Sprintf(`{"id":%d,"title":"Third Title","content":"Third Content"}`, created.ID))
		require.NoError(t, err)

		revTest := webHandlerTestV2{
			user:     &testuser1,
			basePath: fmt.Sprintf("/api/v2/projects/1/wiki/pages/%d/revisions", created.ID),
			idParam:  "revision",
			t:        t,
			echo:     owned.echo,
		}

		t.Run("List", func(t *testing.T) {
			rec, err := revTest.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var revs []struct {
				ID      int64  `json:"id"`
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &revs))
			require.Len(t, revs, 2)
			assert.Equal(t, "Second Title", revs[0].Title)
			assert.Equal(t, "Initial Title", revs[1].Title)

			t.Run("ReadOne", func(t *testing.T) {
				recOne, err := revTest.testReadOne(revs[1].ID)
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, recOne.Code)

				var revOne struct {
					ID      int64  `json:"id"`
					Title   string `json:"title"`
					Content string `json:"content"`
				}
				require.NoError(t, json.Unmarshal(recOne.Body.Bytes(), &revOne))
				assert.Equal(t, revs[1].ID, revOne.ID)
				assert.Equal(t, "Initial Title", revOne.Title)
				assert.Equal(t, "Initial Content", revOne.Content)
			})
		})
	})
}
