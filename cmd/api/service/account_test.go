package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gorsk-gin/internal"

	"gorsk-gin/cmd/api/service"
	"gorsk-gin/internal/account"
	"gorsk-gin/internal/auth"

	"github.com/gin-gonic/gin"

	"gorsk-gin/internal/mock"
	"gorsk-gin/internal/mock/mockdb"
)

func TestCreate(t *testing.T) {
	cases := []struct {
		name       string
		req        string
		wantStatus int
		wantResp   *model.User
		adb        *mockdb.Account
		rbac       *mock.RBAC
	}{
		{
			name:       "Invalid request",
			req:        `{"first_name":"John","last_name":"Doe","username":"juzernejm","password":"hunter123","password_confirm":"hunter1234","email":"johndoe@gmail.com","company_id":1,"location_id":2,"role_id":3}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Fail on userSvc",
			req:  `{"first_name":"John","last_name":"Doe","username":"juzernejm","password":"hunter123","password_confirm":"hunter123","email":"johndoe@gmail.com","company_id":1,"location_id":2,"role_id":2}`,
			rbac: &mock.RBAC{
				AccountCreateFn: func(c *gin.Context, roleID, companyID, locationID int) bool {
					return false
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "Success",
			req:  `{"first_name":"John","last_name":"Doe","username":"juzernejm","password":"hunter123","password_confirm":"hunter123","email":"johndoe@gmail.com","company_id":1,"location_id":2,"role_id":2}`,
			rbac: &mock.RBAC{
				AccountCreateFn: func(c *gin.Context, roleID, companyID, locationID int) bool {
					return true
				},
			},
			adb: &mockdb.Account{
				CreateFn: func(c context.Context, usr *model.User) error {
					usr.ID = 1
					usr.CreatedAt = mock.TestTime(2018)
					usr.UpdatedAt = mock.TestTime(2018)
					return nil
				},
			},
			wantResp: &model.User{
				Base: model.Base{
					ID:        1,
					CreatedAt: mock.TestTime(2018),
					UpdatedAt: mock.TestTime(2018),
				},
				FirstName:  "John",
				LastName:   "Doe",
				Username:   "juzernejm",
				Email:      "johndoe@gmail.com",
				CompanyID:  1,
				LocationID: 2,
			},
			wantStatus: http.StatusOK,
		},
	}
	gin.SetMode(gin.TestMode)

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			rg := r.Group("/v1")
			service.NewAccount(account.New(tt.adb, nil, tt.rbac), rg)
			ts := httptest.NewServer(r)
			defer ts.Close()
			path := ts.URL + "/v1/users"
			res, err := http.Post(path, "application/json", bytes.NewBufferString(tt.req))
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if tt.wantResp != nil {
				response := new(model.User)
				if err := json.NewDecoder(res.Body).Decode(response); err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.wantResp, response)
			}
			assert.Equal(t, tt.wantStatus, res.StatusCode)
		})
	}
}

func TestChangePassword(t *testing.T) {
	cases := []struct {
		name       string
		req        string
		wantStatus int
		id         string
		udb        *mockdb.User
		adb        *mockdb.Account
		rbac       *mock.RBAC
	}{
		{
			name:       "Invalid request",
			req:        `{"new_password":"new_password","old_password":"my_old_password", "new_password_confirm":"new_password_cf"}`,
			wantStatus: http.StatusBadRequest,
			id:         "1",
		},
		{
			name: "Fail on RBAC",
			req:  `{"new_password":"newpassw","old_password":"oldpassw", "new_password_confirm":"newpassw"}`,
			rbac: &mock.RBAC{
				EnforceUserFn: func(c *gin.Context, id int) bool {
					return false
				},
			},
			id:         "1",
			wantStatus: http.StatusForbidden,
		},
		{
			name: "Success",
			req:  `{"new_password":"newpassw","old_password":"oldpassw", "new_password_confirm":"newpassw"}`,
			rbac: &mock.RBAC{
				EnforceUserFn: func(c *gin.Context, id int) bool {
					return true
				},
			},
			id: "1",
			udb: &mockdb.User{
				ViewFn: func(c context.Context, id int) (*model.User, error) {
					return &model.User{
						Password: auth.HashPassword("oldpassw"),
					}, nil
				},
			},
			adb: &mockdb.Account{
				ChangePasswordFn: func(c context.Context, usr *model.User) error {
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
	}
	gin.SetMode(gin.TestMode)
	client := &http.Client{}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			rg := r.Group("/v1")
			service.NewAccount(account.New(tt.adb, tt.udb, tt.rbac), rg)
			ts := httptest.NewServer(r)
			defer ts.Close()
			path := ts.URL + "/v1/users/" + tt.id + "/password"
			req, err := http.NewRequest("PATCH", path, bytes.NewBufferString(tt.req))
			if err != nil {
				t.Fatal(err)
			}
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			assert.Equal(t, tt.wantStatus, res.StatusCode)
		})
	}
}
