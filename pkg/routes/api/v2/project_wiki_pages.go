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

package apiv2

import (
	"context"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
)

type projectWikiPageListBody struct {
	Body Paginated[*models.ProjectWikiPage]
}

// RegisterProjectWikiPageRoutes wires ProjectWikiPage CRUD onto the Huma API.
func RegisterProjectWikiPageRoutes(api huma.API) {
	tags := []string{"project_wiki_pages"}

	Register(api, huma.Operation{
		OperationID: "project-wiki-pages-list",
		Summary:     "List wiki pages of a project",
		Description: "Returns all wiki pages of the given project. Requires read access to the project.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/wiki/pages",
		Tags:        tags,
	}, projectWikiPagesList)

	Register(api, huma.Operation{
		OperationID: "project-wiki-pages-read",
		Summary:     "Get a single wiki page of a project",
		Description: "Returns one wiki page of a project. The page must belong to the project in the path. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/wiki/pages/{page}",
		Tags:        tags,
	}, projectWikiPagesRead)

	Register(api, huma.Operation{
		OperationID: "project-wiki-pages-create",
		Summary:     "Create a wiki page in a project",
		Description: "Creates a wiki page in the given project. The parent project is taken from the URL, not the body. The first page in a project automatically becomes the home page.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/wiki/pages",
		Tags:        tags,
	}, projectWikiPagesCreate)

	Register(api, huma.Operation{
		OperationID: "project-wiki-pages-update",
		Summary:     "Update a wiki page of a project",
		Description: "Updates a project wiki page's title, content, parent, position, or home status. Use PATCH for a partial update.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/wiki/pages/{page}",
		Tags:        tags,
	}, projectWikiPagesUpdate)

	Register(api, huma.Operation{
		OperationID: "project-wiki-pages-delete",
		Summary:     "Delete a wiki page of a project",
		Description: "Deletes a project wiki page. Any direct child pages are re-parented to the deleted page's parent.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/wiki/pages/{page}",
		Tags:        tags,
	}, projectWikiPagesDelete)
}

func init() { AddRouteRegistrar(RegisterProjectWikiPageRoutes) }

func projectWikiPagesList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ListParams
}) (*projectWikiPageListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.ProjectWikiPage{ProjectID: in.ProjectID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.ProjectWikiPage)
	if !ok {
		return nil, fmt.Errorf("projectWikiPages.ReadAll returned unexpected type %T (expected []*models.ProjectWikiPage)", result)
	}
	return &projectWikiPageListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

type projectWikiPageReadBody struct {
	models.ProjectWikiPage
	MaxPermission models.Permission `json:"max_permission" readOnly:"true" doc:"The maximum permission the requesting user has on this wiki page (0=read, 1=read/write, 2=admin)."`
}

func projectWikiPagesRead(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ID        int64 `path:"page"`
	conditional.Params
}) (*singleReadBody[projectWikiPageReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	page := &models.ProjectWikiPage{ID: in.ID, ProjectID: in.ProjectID}
	maxPermission, err := handler.DoReadOne(ctx, page, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	body := &projectWikiPageReadBody{ProjectWikiPage: *page, MaxPermission: models.Permission(maxPermission)}
	return conditionalReadResponse(&in.Params, body, page.Updated, maxPermission)
}

func projectWikiPagesCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	Body      models.ProjectWikiPage
}) (*singleBody[models.ProjectWikiPage], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ProjectID = in.ProjectID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectWikiPage]{Body: &in.Body}, nil
}

func projectWikiPagesUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ID        int64 `path:"page"`
	Body      projectWikiPageReadBody
}) (*singleBody[models.ProjectWikiPage], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	page := &in.Body.ProjectWikiPage
	page.ID = in.ID
	page.ProjectID = in.ProjectID
	if err := handler.DoUpdate(ctx, page, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectWikiPage]{Body: page}, nil
}

func projectWikiPagesDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ID        int64 `path:"page"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.ProjectWikiPage{ID: in.ID, ProjectID: in.ProjectID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
