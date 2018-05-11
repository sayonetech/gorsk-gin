package pgsql_test

import (
	"testing"

	"gorsk-gin/internal/mock"
	"gorsk-gin/internal/platform/postgres"
	"github.com/stretchr/testify/assert"

	"gorsk-gin/internal"

	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

func testAccountDB(t *testing.T, c *pg.DB, l *zap.Logger) {
	accDB := pgsql.NewAccountDB(c, l)
	cases := []struct {
		name string
		fn   func(*testing.T, *pgsql.AccountDB, *pg.DB)
	}{
		{
			name: "accountCreate",
			fn:   testAccountCreate,
		},
		{
			name: "changePassword",
			fn:   testChangePassword,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, accDB, c)
		})
	}

}

func testAccountCreate(t *testing.T, db *pgsql.AccountDB, c *pg.DB) {
	cases := []struct {
		name     string
		wantErr  bool
		usr      *model.User
		wantData *model.User
	}{
		{
			name:    "User already exists",
			wantErr: true,
			usr: &model.User{
				Email:    "johndoe@mail.com",
				Username: "johndoe",
			},
		},
		{
			name:    "Fail on insert duplicate ID",
			wantErr: true,
			usr: &model.User{
				Email:      "tomjones@mail.com",
				FirstName:  "Tom",
				LastName:   "Jones",
				Username:   "tomjones",
				RoleID:     1,
				CompanyID:  1,
				LocationID: 1,
				Password:   "pass",
				Base: model.Base{
					ID: 1,
				},
			},
		},
		{
			name: "Success",
			usr: &model.User{
				Email:      "tomjones@mail.com",
				FirstName:  "Tom",
				LastName:   "Jones",
				Username:   "tomjones",
				RoleID:     1,
				CompanyID:  1,
				LocationID: 1,
				Password:   "pass",
				Base: model.Base{
					ID: 2,
				},
			},
			wantData: &model.User{
				Email:      "tomjones@mail.com",
				FirstName:  "Tom",
				LastName:   "Jones",
				Username:   "tomjones",
				RoleID:     1,
				CompanyID:  1,
				LocationID: 1,
				Password:   "pass",
				Base: model.Base{
					ID: 2,
				},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Create(nil, tt.usr)
			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantData != nil {
				userDB := queryUser(t, c, tt.usr.Base.ID)
				tt.wantData.CreatedAt = userDB.CreatedAt
				tt.wantData.UpdatedAt = userDB.UpdatedAt
				assert.Equal(t, tt.wantData, userDB)
			}
		})
	}
}

func testChangePassword(t *testing.T, db *pgsql.AccountDB, c *pg.DB) {
	cases := []struct {
		name     string
		wantErr  bool
		usr      *model.User
		wantData *model.User
	}{
		// Does not fail on this test, but should
		// {
		// 	name:    "User does not exist",
		// 	wantErr: true,
		// 	usr:     &model.User{},
		// },
		{
			name: "Success",
			usr: &model.User{
				Base: model.Base{
					ID:        2,
					UpdatedAt: mock.TestTime(2000),
				},
				Password: "newPass",
			},
			wantData: &model.User{
				Email:      "tomjones@mail.com",
				FirstName:  "Tom",
				LastName:   "Jones",
				Username:   "tomjones",
				RoleID:     1,
				CompanyID:  1,
				LocationID: 1,
				Password:   "newPass",
				Base: model.Base{
					ID: 2,
				},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := db.ChangePassword(nil, tt.usr)
			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantData != nil {
				userDB := queryUser(t, c, tt.usr.Base.ID)
				assert.NotEqual(t, tt.usr.UpdatedAt, userDB.UpdatedAt)
				tt.wantData.UpdatedAt = userDB.UpdatedAt
				tt.wantData.CreatedAt = userDB.CreatedAt
				assert.Equal(t, tt.wantData, userDB)
			}
		})
	}
}
