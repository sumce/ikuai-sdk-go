package accounting

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestPackagesCRUD 覆盖 ppp_package show/add/edit/del 的成功路径与参数断言。
func TestPackagesCRUD(t *testing.T) {
	tests := []struct {
		name  string
		run   func(t *testing.T, svc *PackagesService) error
		check func(t *testing.T, ts *testserver.Server)
	}{
		{
			name: "List nil-opts 默认分页(真机空表形态)",
			run: func(t *testing.T, svc *PackagesService) error {
				page, err := svc.List(context.Background(), nil)
				if err != nil {
					return err
				}
				if page.Total != 1 || len(page.Rows) != 1 || page.Rows[0].PackName != "包月套餐" {
					t.Fatalf("列表解码不符: %+v", page)
				}
				return nil
			},
			check: func(t *testing.T, ts *testserver.Server) {
				wantParam(t, paramOf(t, ts.CallsFor("ppp_package")[0]),
					map[string]any{"TYPE": "total,data", "limit": "0,20"})
			},
		},
		{
			name: "Add 返回 RowId",
			run: func(t *testing.T, svc *PackagesService) error {
				id, err := svc.Add(context.Background(), Package{
					PackName: "包月套餐", PackTime: 1, PeriodUnit: PeriodUnitMonth,
					Price: 30.5, UpSpeed: 1024, DownSpeed: 2048,
				})
				if err != nil {
					return err
				}
				_ = id // testserver 静态响应不携带 RowId，参数断言见 check
				return nil
			},
			check: func(t *testing.T, ts *testserver.Server) {
				rec := ts.CallsFor("ppp_package")[0]
				if rec.Action != "add" {
					t.Fatalf("action = %q, 期望 add", rec.Action)
				}
				p := paramOf(t, rec)
				if p["packname"] != "包月套餐" || p["period_unit"] != PeriodUnitMonth ||
					p["price"] != 30.5 || p["up_speed"] != float64(1024) {
					t.Fatalf("add param 字段不符: %v", p)
				}
			},
		},
		{
			name: "Edit 携带 id",
			run: func(t *testing.T, svc *PackagesService) error {
				return svc.Edit(context.Background(), Package{ID: 3, PackName: "改名", Price: 20})
			},
			check: func(t *testing.T, ts *testserver.Server) {
				rec := ts.CallsFor("ppp_package")[0]
				if rec.Action != "edit" {
					t.Fatalf("action = %q, 期望 edit", rec.Action)
				}
				p := paramOf(t, rec)
				if p["id"] != float64(3) || p["packname"] != "改名" {
					t.Fatalf("edit param 字段不符: %v", p)
				}
			},
		},
		{
			name: "Delete 单个 → id",
			run: func(t *testing.T, svc *PackagesService) error {
				return svc.Delete(context.Background(), 5)
			},
			check: func(t *testing.T, ts *testserver.Server) {
				rec := ts.CallsFor("ppp_package")[0]
				if rec.Action != "del" {
					t.Fatalf("action = %q, 期望 del", rec.Action)
				}
				wantParam(t, paramOf(t, rec), map[string]any{"id": float64(5)})
			},
		},
		{
			name: "Delete 多个 → ids",
			run: func(t *testing.T, svc *PackagesService) error {
				return svc.Delete(context.Background(), 5, 6)
			},
			check: func(t *testing.T, ts *testserver.Server) {
				rec := ts.CallsFor("ppp_package")[0]
				wantParam(t, paramOf(t, rec), map[string]any{"ids": []any{float64(5), float64(6)}})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ppp_package", "show", map[string]any{
				"total": 1,
				"data": []any{map[string]any{
					"id": 1, "packname": "包月套餐", "packtime": 1,
					"period_unit": "month", "price": 30.5,
					"up_speed": 1024, "down_speed": 2048,
				}},
			})
			ts.SetResponse("ppp_package", "add", map[string]any{})
			ts.SetResponse("ppp_package", "edit", map[string]any{})
			ts.SetResponse("ppp_package", "del", map[string]any{})
			svc := New(ts.Client()).Packages
			if err := tt.run(t, svc); err != nil {
				t.Fatalf("执行失败: %v", err)
			}
			tt.check(t, ts)
		})
	}
}

// TestPackagesError 覆盖固件错误路径。
func TestPackagesError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ppp_package", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95000, ErrMsg: "失败", FuncName: "ppp_package", Action: action}
	})
	svc := New(ts.Client()).Packages
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("List 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95000 {
		t.Fatalf("期望 APIError(95000), got %v", err)
	}
	if _, err := svc.Add(ctx, Package{}); err == nil {
		t.Fatal("Add 应失败")
	}
	if err := svc.Edit(ctx, Package{ID: 1}); err == nil {
		t.Fatal("Edit 应失败")
	}
	if err := svc.Delete(ctx, 1); err == nil {
		t.Fatal("Delete 应失败")
	}
}
