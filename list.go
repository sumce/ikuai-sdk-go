package ikuai

import (
	"fmt"
	"sort"
	"strings"
)

// Filter is one FILTERn entry of a show param, expressed as the iKuai
// triple "field,operator,value" (e.g. "read,==,0" for unread system
// messages). Op values observed in the web console: ==, !=, like, >, <.
type Filter struct {
	Field string
	Op    string
	Value any
}

func (f Filter) encode() string {
	return fmt.Sprintf("%s,%s,%v", f.Field, f.Op, f.Value)
}

// ListOptions describes a paginated, filtered, ordered show request —
// the shape every list-style func_name understands. Zero-value options
// match the web console defaults (first 20 rows, unfiltered).
type ListOptions struct {
	// Type overrides the TYPE selector; defaults to "total,data".
	Type string
	// Offset and Limit form the "offset,rows" limit string. Limit 0
	// means "no page size" (iKuai interprets "N,0" as all rows).
	Offset int
	Limit  int
	// OrderBy is a raw field name (e.g. "ip_addr_int", "id").
	OrderBy string
	// Order is "asc" or "desc".
	Order string
	// Finds restricts keyword search to these comma-separated fields
	// (e.g. "ip_addr,mac,comment,username,hostname").
	Finds string
	// Keywords is the search text for Finds.
	Keywords string
	// Filters become FILTER1..n in order.
	Filters []Filter
	// Extra merges arbitrary keys into the param, overriding anything
	// built above (used for func-specific selectors such as
	// "interface", "datetype", "time_range", "math", ...).
	Extra map[string]any
}

// DefaultListOptions returns the web console default: first page of 20
// rows with totals.
func DefaultListOptions() *ListOptions {
	return &ListOptions{Type: "total,data", Limit: 20}
}

// BuildParam renders ListOptions into the /Action/call param map.
// Returns nil when every field is empty (some show actions require an
// empty param). 业务包通过它构造分页/过滤/排序参数。
func (o *ListOptions) BuildParam() map[string]any {
	if o == nil {
		return nil
	}
	p := map[string]any{}
	typeDefault := o.Type
	if typeDefault == "" && (o.Offset != 0 || o.Limit != 0 || o.OrderBy != "" || o.Order != "" ||
		o.Finds != "" || o.Keywords != "" || len(o.Filters) > 0) {
		typeDefault = "total,data"
	}
	if typeDefault != "" {
		p["TYPE"] = typeDefault
	}
	if o.Offset != 0 || o.Limit != 0 {
		p["limit"] = fmt.Sprintf("%d,%d", o.Offset, o.Limit)
	}
	if o.OrderBy != "" {
		p["ORDER_BY"] = o.OrderBy
	}
	if o.Order != "" {
		p["ORDER"] = o.Order
	}
	if o.Finds != "" {
		p["FINDS"] = o.Finds
	}
	if o.Keywords != "" {
		p["KEYWORDS"] = o.Keywords
	}
	for i, f := range o.Filters {
		p[fmt.Sprintf("FILTER%d", i+1)] = f.encode()
	}
	for k, v := range o.Extra {
		p[k] = v
	}
	return p
}

// RawPage is the decoded "total,data" show payload with rows kept as
// generic maps (use ListInto for typed decoding).
type RawPage struct {
	Total int64
	Data  []map[string]any
}

// ListResult is the union of show payload shapes seen in the firmware:
// {"total":N,"data":[...]} and occasionally the keys hoisted to the top
// level of Data.
type listResult struct {
	Total int64            `json:"total"`
	Data  []map[string]any `json:"data"`
}

// DecodeRawPage 解码 "total,data" 形态的 Data，行保留为通用 map。
// 需要强类型时使用 DecodePage[T]。
func DecodeRawPage(data []byte) (*RawPage, error) {
	var lr listResult
	if err := jsonUnmarshal(sanitizeJSON(data), &lr); err != nil {
		return nil, fmt.Errorf("ikuai: decode page: %w (data: %.200s)", err, data)
	}
	return &RawPage{Total: lr.Total, Data: lr.Data}, nil
}

// Page is the typed variant of a "total,data" payload: rows decoded
// into T.
type Page[T any] struct {
	Total int64
	Rows  []T
}

// DecodePage 解码行数据为强类型 T。接受 {"total":N,"data":[...]} 与
// 裸数组 [...] 两种形态；T 的 json tag 必须与固件字段一致。
func DecodePage[T any](data []byte) (*Page[T], error) {
	raw := sanitizeJSON(data)
	if len(raw) > 0 && raw[0] == '[' {
		var rows []T
		if err := jsonUnmarshal(raw, &rows); err != nil {
			return nil, fmt.Errorf("ikuai: decode rows: %w", err)
		}
		return &Page[T]{Rows: rows}, nil
	}
	var lr struct {
		Total int64 `json:"total"`
		Data  []T   `json:"data"`
	}
	if err := jsonUnmarshal(raw, &lr); err != nil {
		return nil, fmt.Errorf("ikuai: decode page: %w (data: %.200s)", err, data)
	}
	return &Page[T]{Total: lr.Total, Rows: lr.Data}, nil
}

// normalizeIDs converts the many id shapes callers may pass (int,
// string, []int, []string, "1,2,3") into the list used by del params.
func normalizeIDs(ids ...any) []any {
	out := make([]any, 0, len(ids))
	for _, id := range ids {
		switch v := id.(type) {
		case string:
			if strings.Contains(v, ",") {
				for _, part := range strings.Split(v, ",") {
					if part = strings.TrimSpace(part); part != "" {
						out = append(out, part)
					}
				}
				continue
			}
			out = append(out, v)
		case []string:
			for _, s := range v {
				out = append(out, s)
			}
		case []int:
			for _, i := range v {
				out = append(out, i)
			}
		case []int64:
			for _, i := range v {
				out = append(out, i)
			}
		case nil:
			continue
		default:
			out = append(out, v)
		}
	}
	return out
}

// mergeFields deep-copies optional base and overlay maps for edit calls.
func mergeFields(base, overlay map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range overlay {
		out[k] = v
	}
	// Deterministic maps are not required by the firmware, but sorted
	// insertion keeps debugging predictable.
	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return out
}
