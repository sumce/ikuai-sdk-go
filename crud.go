package ikuai

import "context"

// Crud 是任意 func_name 的通用 CRUD 服务，作为业务包的兜底通道：
// 即使某个功能尚未提供强类型封装，也可以通过它完整操作
// （show/add/edit/del/del_all/set）。
//
//	c := client.CRUD("dnat")
//	page, _ := c.List(ctx, ikuai.DefaultListOptions())
//	rowID, _ := c.Add(ctx, map[string]any{"enabled": "yes", ...})
type Crud struct {
	c  *Client
	fn string
}

// CRUD 返回指定 func_name 的通用 CRUD 服务。
func (c *Client) CRUD(funcName string) *Crud {
	return &Crud{c: c, fn: funcName}
}

// FuncName 返回本服务绑定的 func_name。
func (cr *Crud) FuncName() string { return cr.fn }

// Show 以自定义 param 调用 action=show，返回原始封装。
func (cr *Crud) Show(ctx context.Context, param any) (*Response, error) {
	return cr.c.Call(ctx, cr.fn, "show", param)
}

// List 以 ListOptions 调用 action=show，返回 {"total","data"} 分页
// 结果（行为 map）。
func (cr *Crud) List(ctx context.Context, opts *ListOptions) (*RawPage, error) {
	resp, err := cr.c.Call(ctx, cr.fn, "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return DecodeRawPage(resp.Data)
}

// ListInto 同 List，但把行解码为强类型 T。
func ListInto[T any](c *Client, funcName string, ctx context.Context, opts *ListOptions) (*Page[T], error) {
	resp, err := c.Call(ctx, funcName, "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return DecodePage[T](resp.Data)
}

// Count 只取 total：TYPE=total 的 show 调用。
func (cr *Crud) Count(ctx context.Context) (int64, error) {
	resp, err := cr.c.Call(ctx, cr.fn, "show", map[string]any{"TYPE": "total"})
	if err != nil {
		return 0, err
	}
	if n, ok := resp.Count(); ok {
		return n, nil
	}
	var p struct {
		Total int64 `json:"total"`
	}
	if err := resp.decodeData(&p); err != nil {
		return 0, err
	}
	return p.Total, nil
}

// Add 调用 action=add 新增一条记录，返回固件分配的 RowId。
func (cr *Crud) Add(ctx context.Context, fields any) (int64, error) {
	resp, err := cr.c.Call(ctx, cr.fn, "add", fields)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 调用 action=edit，参数自动合并 id。fields 为 map 或带 json tag
// 的 struct。
func (cr *Crud) Edit(ctx context.Context, id any, fields any) error {
	_, err := cr.c.Call(ctx, cr.fn, "edit", mergeIDField(id, fields))
	return err
}

// Delete 调用 action=del。单个 id 发送 {"id": id}；多个发送
// {"ids": [...]}；也接受 "1,2,3" 形式的字符串。
func (cr *Crud) Delete(ctx context.Context, ids ...any) error {
	ids = normalizeIDs(ids...)
	if len(ids) == 1 {
		_, err := cr.c.Call(ctx, cr.fn, "del", map[string]any{"id": ids[0]})
		return err
	}
	_, err := cr.c.Call(ctx, cr.fn, "del", map[string]any{"ids": ids})
	return err
}

// DeleteAll 调用 action=del_all 清空（部分 func 支持按条件清空，通过
// fields 传入过滤字段）。
func (cr *Crud) DeleteAll(ctx context.Context, fields ...map[string]any) error {
	var param any
	if len(fields) > 0 {
		param = fields[0]
	}
	_, err := cr.c.Call(ctx, cr.fn, "del_all", param)
	return err
}

// Set 调用 action=set（基础配置类 func 的保存动作，如 basic/lan/dns）。
func (cr *Crud) Set(ctx context.Context, fields any) error {
	_, err := cr.c.Call(ctx, cr.fn, "set", fields)
	return err
}

// IDParam 把 ID 列表渲染成 del 动作的标准参数：单个 → {"id": v}，
// 多个 → {"ids": [...]}，空列表 → {"ids": []}。业务包的 Delete 方法
// 统一用它，保证与固件约定一致。
func IDParam[T any](ids []T) map[string]any {
	switch len(ids) {
	case 0:
		return map[string]any{"ids": []any{}}
	case 1:
		return map[string]any{"id": ids[0]}
	default:
		out := make([]any, len(ids))
		for i, v := range ids {
			out[i] = v
		}
		return map[string]any{"ids": out}
	}
}

// mergeIDField 把 id 合并进 edit 参数。struct 通过 JSON 往返合并，map
// 直接注入。
func mergeIDField(id, fields any) map[string]any {
	out := map[string]any{"id": id}
	switch f := fields.(type) {
	case map[string]any:
		for k, v := range f {
			out[k] = v
		}
	default:
		// struct 等其它类型：编码为 JSON 再展开，保证与直传字段一致。
		raw, err := jsonMarshal(fields)
		if err == nil {
			var m map[string]any
			if err := jsonUnmarshal(raw, &m); err == nil {
				for k, v := range m {
					out[k] = v
				}
			}
		}
	}
	return out
}
