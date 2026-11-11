package ikuai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

// Response is the common envelope of every /Action/call result:
//
//	{"Result":30000,"ErrMsg":"Success","Data":{...}}
//	{"Result":30000,"ErrMsg":"Success","RowId":29}
type Response struct {
	Result int             `json:"Result"`
	ErrMsg string          `json:"ErrMsg"`
	Data   json.RawMessage `json:"Data"`
	// RowId is set by "add" actions and carries the id of the new row.
	RowId int64 `json:"RowId"`
}

// OK reports whether the call succeeded (Result == 30000).
func (r *Response) OK() bool { return r != nil && r.Result == CodeSuccess }

// Count returns the "total" field of a show payload when present.
func (r *Response) Count() (int64, bool) {
	var p struct {
		Total int64 `json:"total"`
	}
	if r == nil || len(r.Data) == 0 || r.Data[0] != '{' {
		return 0, false
	}
	if err := json.Unmarshal(r.Data, &p); err != nil {
		return 0, false
	}
	return p.Total, true
}

// decodeData unmarshals the Data member into out (which may be nil to
// ignore the payload).
func (r *Response) decodeData(out any) error {
	if out == nil {
		return nil
	}
	if r == nil || len(r.Data) == 0 || bytes.Equal(r.Data, []byte("null")) {
		return nil
	}
	return json.Unmarshal(r.Data, out)
}

var emptyValueTail = regexp.MustCompile(`("(?:Data|data)"\s*:\s*)([,}\]])`)

// sanitizeJSON repairs quirks seen in real iKuai firmware responses,
// e.g. auth_code:show replying `{"Result":30000,"ErrMsg":"Success","Data":}`
// (an empty value with no literal). It returns the input unchanged when
// it already parses.
func sanitizeJSON(b []byte) []byte {
	if json.Valid(b) {
		return b
	}
	// `"Data":}` / `"Data":,` / `"Data":]` → 空值补 null，保留分隔符。
	// 注意 ${1} 定界：$1null 会被 Go regexp 解析成名为 "1n" 的分组。
	return emptyValueTail.ReplaceAll(b, []byte(`${1}null${2}`))
}

// unmarshalStrict-ish helper used by the client: decode the envelope,
// sanitizing when the first attempt fails.
func decodeEnvelope(body []byte) (*Response, error) {
	var resp Response
	if err := json.Unmarshal(body, &resp); err != nil {
		fixed := sanitizeJSON(body)
		if err := json.Unmarshal(fixed, &resp); err != nil {
			return nil, fmt.Errorf("ikuai: cannot decode response envelope: %w (body: %.200s)", err, body)
		}
		resp.Data = json.RawMessage(sanitizeJSON(resp.Data))
	}
	return &resp, nil
}
