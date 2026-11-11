package accounting

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestCouponList 覆盖 coupon.show 的成功路径与参数断言。
func TestCouponList(t *testing.T) {
	tests := []struct {
		name  string
		opts  *ikuai.ListOptions
		data  map[string]any
		wantP map[string]any
	}{
		{
			name:  "nil-opts 默认分页(真机空表形态)",
			opts:  nil,
			data:  map[string]any{"total": 0, "data": []any{}},
			wantP: map[string]any{"TYPE": "total,data", "limit": "0,20"},
		},
		{
			name: "带行数据解码(Raw 原样保留)",
			opts: &ikuai.ListOptions{Offset: 0, Limit: 5},
			data: map[string]any{
				"total": 1,
				"data": []any{map[string]any{
					"id": 1, "code": "A1B2C3", "expires": 1791741676, "timeout": 60,
				}},
			},
			wantP: map[string]any{"TYPE": "total,data", "limit": "0,5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("coupon", "show", tt.data)
			page, err := New(ts.Client()).Coupon.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if tt.name == "nil-opts 默认分页(真机空表形态)" {
				if page.Total != 0 || len(page.Rows) != 0 {
					t.Fatalf("空列表解码不符: %+v", page)
				}
			} else {
				if page.Total != 1 || len(page.Rows) != 1 {
					t.Fatalf("行数不符: %+v", page)
				}
				m, err := page.Rows[0].Map()
				if err != nil {
					t.Fatalf("Map() 失败: %v", err)
				}
				if m["code"] != "A1B2C3" {
					t.Fatalf("行内容不符: %v", m)
				}
			}
			wantParam(t, paramOf(t, ts.CallsFor("coupon")[0]), tt.wantP)
		})
	}
}

// TestCouponGenerate 覆盖 add 写动作的参数断言（仅离线）。
func TestCouponGenerate(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("coupon", "add", map[string]any{})
	if _, err := New(ts.Client()).Coupon.Generate(context.Background(), CouponBuildParam{
		Number: 100, Length: 8, Type: CouponTypeNumberLetters,
		ValidPeriod: 30, Timeout: 120,
	}); err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
	rec := ts.CallsFor("coupon")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	if p["coupon_number"] != float64(100) || p["coupon_length"] != float64(8) ||
		p["coupon_type"] != CouponTypeNumberLetters || p["valid_period"] != float64(30) ||
		p["timeout"] != float64(120) {
		t.Fatalf("add param 字段不符: %v", p)
	}
}

// TestCouponDelete 覆盖 del 写动作的参数断言（仅离线）。
func TestCouponDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("coupon", "del", map[string]any{})
	if err := New(ts.Client()).Coupon.Delete(context.Background(), 3); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	rec := ts.CallsFor("coupon")[0]
	if rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"id": float64(3)})
}

// TestCouponError 覆盖固件错误路径。
func TestCouponError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("coupon", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95000, ErrMsg: "失败", FuncName: "coupon", Action: action}
	})
	svc := New(ts.Client()).Coupon
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("List 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95000 {
		t.Fatalf("期望 APIError(95000), got %v", err)
	}
	if _, err := svc.Generate(ctx, CouponBuildParam{}); err == nil {
		t.Fatal("Generate 应失败")
	}
	if err := svc.Delete(ctx, 1); err == nil {
		t.Fatal("Delete 应失败")
	}
}
