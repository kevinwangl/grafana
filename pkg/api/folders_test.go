package api

import (
	"encoding/json"
	"testing"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/alerting"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFoldersApiEndpoint(t *testing.T) {
	Convey("Given a folder which does not have an acl", t, func() {
		fakeFolder := m.NewDashboardFolder("Folder")
		fakeFolder.Id = 1
		fakeFolder.IsFolder = true
		fakeFolder.HasAcl = false

		bus.AddHandler("test", func(query *m.GetDashboardQuery) error {
			query.Result = fakeFolder
			return nil
		})

		viewerRole := m.ROLE_VIEWER
		editorRole := m.ROLE_EDITOR

		aclMockResp := []*m.DashboardAclInfoDTO{
			{Role: &viewerRole, Permission: m.PERMISSION_VIEW},
			{Role: &editorRole, Permission: m.PERMISSION_EDIT},
		}

		bus.AddHandler("test", func(query *m.GetDashboardAclInfoListQuery) error {
			query.Result = aclMockResp
			return nil
		})

		bus.AddHandler("test", func(query *m.GetTeamsByUserQuery) error {
			query.Result = []*m.Team{}
			return nil
		})

		cmd := m.SaveDashboardCommand{
			IsFolder: true,
			Dashboard: simplejson.NewFromAny(map[string]interface{}{
				"title": fakeFolder.Title,
				"id":    fakeFolder.Id,
			}),
		}

		Convey("When user is an Org Viewer", func() {
			role := m.ROLE_VIEWER

			loggedInUserScenarioWithRole("When calling GET on", "GET", "/api/folders/1", "/api/folders/:id", role, func(sc *scenarioContext) {
				folder := getFolderShouldReturn200(sc)

				Convey("Should not be able to edit or save folder", func() {
					So(folder.CanEdit, ShouldBeFalse)
					So(folder.CanSave, ShouldBeFalse)
					So(folder.CanAdmin, ShouldBeFalse)
				})
			})

			loggedInUserScenarioWithRole("When calling DELETE on", "DELETE", "/api/folders/1", "/api/folders/:id", role, func(sc *scenarioContext) {
				callDeleteFolder(sc)
				So(sc.resp.Code, ShouldEqual, 403)
			})

			postDashboardScenario("When calling POST on", "/api/folders", "/api/folders", role, cmd, func(sc *scenarioContext) {
				callCreateFolder(sc)
				So(sc.resp.Code, ShouldEqual, 403)
			})
		})

		Convey("When user is an Org Editor", func() {
			role := m.ROLE_EDITOR

			loggedInUserScenarioWithRole("When calling GET on", "GET", "/api/folders/1", "/api/folders/:id", role, func(sc *scenarioContext) {
				folder := getFolderShouldReturn200(sc)

				Convey("Should be able to edit or save folder", func() {
					So(folder.CanEdit, ShouldBeTrue)
					So(folder.CanSave, ShouldBeTrue)
					So(folder.CanAdmin, ShouldBeFalse)
				})
			})

			loggedInUserScenarioWithRole("When calling DELETE on", "DELETE", "/api/folders/1", "/api/folders/:id", role, func(sc *scenarioContext) {
				callDeleteFolder(sc)
				So(sc.resp.Code, ShouldEqual, 200)
			})

			postDashboardScenario("When calling POST on", "/api/folders", "/api/folders", role, cmd, func(sc *scenarioContext) {
				callCreateFolder(sc)
				So(sc.resp.Code, ShouldEqual, 200)
			})
		})
	})

	Convey("Given a folder which have an acl", t, func() {
		fakeFolder := m.NewDashboardFolder("Folder")
		fakeFolder.Id = 1
		fakeFolder.IsFolder = true
		fakeFolder.HasAcl = true

		bus.AddHandler("test", func(query *m.GetDashboardQuery) error {
			query.Result = fakeFolder
			return nil
		})

		aclMockResp := []*m.DashboardAclInfoDTO{
			{
				DashboardId: 1,
				Permission:  m.PERMISSION_EDIT,
				UserId:      200,
			},
		}

		bus.AddHandler("test", func(query *m.GetDashboardAclInfoListQuery) error {
			query.Result = aclMockResp
			return nil
		})

		bus.AddHandler("test", func(query *m.GetTeamsByUserQuery) error {
			query.Result = []*m.Team{}
			return nil
		})

		cmd := m.SaveDashboardCommand{
			IsFolder: true,
			Dashboard: simplejson.NewFromAny(map[string]interface{}{
				"title": fakeFolder.Title,
				"id":    fakeFolder.Id,
			}),
		}

		Convey("When user is an Org Viewer and has no permissions for this folder", func() {
			role := m.ROLE_VIEWER

			loggedInUserScenarioWithRole("When calling GET on", "GET", "/api/folders/1", "/api/folders/:id", role, func(sc *scenarioContext) {
				sc.handlerFunc = GetFolderById
				sc.fakeReqWithParams("GET", sc.url, map[string]string{}).exec()

				Convey("Should be denied access", func() {
					So(sc.resp.Code, ShouldEqual, 403)
				})
			})

			loggedInUserScenarioWithRole("When calling DELETE on", "DELETE", "/api/folders/1", "/api/folders/:id", role, func(sc *scenarioContext) {
				callDeleteFolder(sc)
				So(sc.resp.Code, ShouldEqual, 403)
			})

			postDashboardScenario("When calling POST on", "/api/folders", "/api/folders", role, cmd, func(sc *scenarioContext) {
				callCreateFolder(sc)
				So(sc.resp.Code, ShouldEqual, 403)
			})
		})

		Convey("When user is an Org Editor and has no permissions for this folder", func() {
			role := m.ROLE_EDITOR

			loggedInUserScenarioWithRole("When calling GET on", "GET", "/api/folders/1", "/api/folders/:id", role, func(sc *scenarioContext) {
				sc.handlerFunc = GetFolderById
				sc.fakeReqWithParams("GET", sc.url, map[string]string{}).exec()

				Convey("Should be denied access", func() {
					So(sc.resp.Code, ShouldEqual, 403)
				})
			})

			loggedInUserScenarioWithRole("When calling DELETE on", "DELETE", "/api/folders/1", "/api/folders/:id", role, func(sc *scenarioContext) {
				callDeleteFolder(sc)
				So(sc.resp.Code, ShouldEqual, 403)
			})

			postDashboardScenario("When calling POST on", "/api/folders", "/api/folders", role, cmd, func(sc *scenarioContext) {
				callCreateFolder(sc)
				So(sc.resp.Code, ShouldEqual, 403)
			})
		})
	})
}

func getFolderShouldReturn200(sc *scenarioContext) dtos.Folder {
	sc.handlerFunc = GetFolderById
	sc.fakeReqWithParams("GET", sc.url, map[string]string{}).exec()

	So(sc.resp.Code, ShouldEqual, 200)

	folder := dtos.Folder{}
	err := json.NewDecoder(sc.resp.Body).Decode(&folder)
	So(err, ShouldBeNil)

	return folder
}

func callDeleteFolder(sc *scenarioContext) {
	bus.AddHandler("test", func(cmd *m.DeleteDashboardCommand) error {
		return nil
	})

	sc.handlerFunc = DeleteDashboard
	sc.fakeReqWithParams("DELETE", sc.url, map[string]string{}).exec()
}

func callCreateFolder(sc *scenarioContext) {
	bus.AddHandler("test", func(cmd *alerting.ValidateDashboardAlertsCommand) error {
		return nil
	})

	bus.AddHandler("test", func(cmd *m.SaveDashboardCommand) error {
		cmd.Result = &m.Dashboard{Id: 1, Slug: "folder", Version: 2}
		return nil
	})

	bus.AddHandler("test", func(cmd *alerting.UpdateDashboardAlertsCommand) error {
		return nil
	})

	sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()
}

func callUpdateFolder(sc *scenarioContext) {
	bus.AddHandler("test", func(cmd *alerting.ValidateDashboardAlertsCommand) error {
		return nil
	})

	bus.AddHandler("test", func(cmd *m.SaveDashboardCommand) error {
		cmd.Result = &m.Dashboard{Id: 1, Slug: "folder", Version: 3}
		return nil
	})

	bus.AddHandler("test", func(cmd *alerting.UpdateDashboardAlertsCommand) error {
		return nil
	})

	sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()
}
