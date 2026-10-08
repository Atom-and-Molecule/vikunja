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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humaecho5"
	webfiles "code.vikunja.io/api/pkg/web/files"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

type projectWikiPageAttachmentListBody struct {
	Body Paginated[*models.ProjectWikiPageAttachment]
}

type projectWikiPageAttachmentUploadInput struct {
	ProjectID int64 `path:"project" doc:"The id of the project."`
	PageID    int64 `path:"page" doc:"The id of the wiki page."`
	RawBody   huma.MultipartFormFiles[struct {
		Files []huma.FormFile `form:"files" required:"true" doc:"One or more files to upload as attachments."`
	}]
}

type projectWikiPageAttachmentUploadBody struct {
	Body *webfiles.WikiAttachmentUploadResult
}

// RegisterProjectWikiPageAttachmentRoutes wires wiki page attachment endpoints onto Huma API.
func RegisterProjectWikiPageAttachmentRoutes(api huma.API) {
	tags := []string{"project"}

	Register(api, huma.Operation{
		OperationID: "project-wiki-page-attachments-list",
		Summary:     "List a wiki page's attachments",
		Description: "Returns attachment metadata for one wiki page, paginated. Requires read access to the project.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/wiki/pages/{page}/attachments",
		Tags:        tags,
	}, projectWikiPageAttachmentsList)

	Register(api, huma.Operation{
		OperationID: "project-wiki-page-attachments-upload",
		Summary:     "Upload wiki page attachments",
		Description: "Uploads one or more files as attachments to a wiki page via multipart/form-data under the \"files\" field. Requires write access to the project.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/wiki/pages/{page}/attachments",
		Tags:        tags,
		MaxBodyBytes: (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024,
	}, projectWikiPageAttachmentsUpload)

	Register(api, huma.Operation{
		OperationID: "project-wiki-page-attachments-download",
		Summary:     "Download a wiki page attachment",
		Description: "Returns raw bytes of one wiki page attachment. Requires read access to the project. Pass preview_size for image thumbnail.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/wiki/pages/{page}/attachments/{attachment}",
		Tags:        tags,
		Responses: map[string]*huma.Response{
			"200": {
				Description: "The attachment file bytes.",
				Content: map[string]*huma.MediaType{
					"application/octet-stream": {
						Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
					},
				},
			},
		},
	}, projectWikiPageAttachmentsDownload)

	Register(api, huma.Operation{
		OperationID: "project-wiki-page-attachments-delete",
		Summary:     "Delete a wiki page attachment",
		Description: "Deletes one attachment and its underlying file. Requires write access to the project.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/wiki/pages/{page}/attachments/{attachment}",
		Tags:        tags,
	}, projectWikiPageAttachmentsDelete)
}

func init() { AddRouteRegistrar(RegisterProjectWikiPageAttachmentRoutes) }

func projectWikiPageAttachmentsList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	PageID    int64 `path:"page"`
	ListParams
}) (*projectWikiPageAttachmentListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	page := &models.ProjectWikiPage{ID: in.PageID, ProjectID: in.ProjectID}
	if _, err := handler.DoReadOne(ctx, page, a); err != nil {
		return nil, translateDomainError(err)
	}

	result, _, total, err := handler.DoReadAll(ctx, &models.ProjectWikiPageAttachment{PageID: in.PageID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.ProjectWikiPageAttachment)
	if !ok {
		return nil, fmt.Errorf("projectWikiPageAttachments.ReadAll returned unexpected type %T", result)
	}
	return &projectWikiPageAttachmentListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func projectWikiPageAttachmentsUpload(ctx context.Context, in *projectWikiPageAttachmentUploadInput) (*projectWikiPageAttachmentUploadBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	page := &models.ProjectWikiPage{ID: in.PageID, ProjectID: in.ProjectID}
	if _, err := handler.DoReadOne(ctx, page, a); err != nil {
		return nil, translateDomainError(err)
	}

	s := db.NewSession()
	defer s.Close()

	formFiles := in.RawBody.Data().Files
	uploads := make([]*models.AttachmentToUpload, 0, len(formFiles))
	for _, file := range formFiles {
		uploads = append(uploads, &models.AttachmentToUpload{Reader: file, Filename: file.Filename, Size: uint64(file.Size)})
	}

	success, failures, err := models.UploadProjectWikiPageAttachments(s, a, in.PageID, uploads)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	return &projectWikiPageAttachmentUploadBody{Body: webfiles.BuildWikiUploadResult(success, failures)}, nil
}

func projectWikiPageAttachmentsDownload(ctx context.Context, in *struct {
	ProjectID    int64  `path:"project"`
	PageID       int64  `path:"page"`
	AttachmentID int64  `path:"attachment"`
	PreviewSize  string `query:"preview_size" enum:"sm,md,lg,xl"`
}) (*huma.StreamResponse, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	page := &models.ProjectWikiPage{ID: in.PageID, ProjectID: in.ProjectID}
	if _, err := handler.DoReadOne(ctx, page, a); err != nil {
		return nil, translateDomainError(err)
	}

	s := db.NewSession()
	defer s.Close()

	previewSize := models.GetPreviewSizeFromString(in.PreviewSize)
	pa, preview, err := models.LoadProjectWikiPageAttachmentForDownload(s, a, in.PageID, in.AttachmentID, previewSize)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		c := humaecho5.Unwrap(hctx)
		webfiles.WriteWikiAttachmentDownload((*c).Response(), (*c).Request(), pa, preview)
	}}, nil
}

func projectWikiPageAttachmentsDelete(ctx context.Context, in *struct {
	ProjectID    int64 `path:"project"`
	PageID       int64 `path:"page"`
	AttachmentID int64 `path:"attachment"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	page := &models.ProjectWikiPage{ID: in.PageID, ProjectID: in.ProjectID}
	if _, err := handler.DoReadOne(ctx, page, a); err != nil {
		return nil, translateDomainError(err)
	}

	if err := handler.DoDelete(ctx, &models.ProjectWikiPageAttachment{ID: in.AttachmentID, PageID: in.PageID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
