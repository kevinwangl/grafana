package api

import (
	"fmt"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/middleware"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/guardian"
)

func getFolderHelper(orgId int64, slug string, id int64) (*m.Dashboard, Response) {
	query := m.GetDashboardQuery{Slug: slug, Id: id, OrgId: orgId}
	if err := bus.Dispatch(&query); err != nil {
		if err == m.ErrDashboardNotFound {
			err = m.ErrFolderNotFound
		}

		return nil, ApiError(404, "Folder not found", err)
	}

	if !query.Result.IsFolder {
		return nil, ApiError(404, "Folder not found", m.ErrFolderNotFound)
	}

	return query.Result, nil
}

func folderGuardianResponse(err error) Response {
	if err != nil {
		return ApiError(500, "Error while checking folder permissions", err)
	}

	return ApiError(403, "Access denied to this folder", nil)
}

func GetFolderById(c *middleware.Context) Response {
	folder, rsp := getFolderHelper(c.OrgId, "", c.ParamsInt64(":id"))
	if rsp != nil {
		return rsp
	}

	guardian := guardian.NewDashboardGuardian(folder.Id, c.OrgId, c.SignedInUser)
	if canView, err := guardian.CanView(); err != nil || !canView {
		fmt.Printf("%v", err)
		return folderGuardianResponse(err)
	}

	canEdit, _ := guardian.CanEdit()
	canSave, _ := guardian.CanSave()
	canAdmin, _ := guardian.CanAdmin()

	// Finding creator and last updater of the folder
	updater, creator := "Anonymous", "Anonymous"
	if folder.UpdatedBy > 0 {
		updater = getUserLogin(folder.UpdatedBy)
	}
	if folder.CreatedBy > 0 {
		creator = getUserLogin(folder.CreatedBy)
	}

	dto := dtos.Folder{
		Id:        folder.Id,
		Title:     folder.Title,
		Slug:      folder.Slug,
		HasAcl:    folder.HasAcl,
		CanSave:   canSave,
		CanEdit:   canEdit,
		CanAdmin:  canAdmin,
		CreatedBy: creator,
		Created:   folder.Created,
		UpdatedBy: updater,
		Updated:   folder.Updated,
		Version:   folder.Version,
	}

	return Json(200, dto)
}
