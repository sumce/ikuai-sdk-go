// Package testserver 提供所有业务包测试共用的「假爱快路由器」。
//
// 它以 httptest 实现了固件的核心端点：
//
//	POST /Action/login  —— 校验 username + md5(passwd)，成功回
//	                       Result=10000 并下发 Set-Cookie: sess_key；
//	                       失败回 Result=10001。
//	POST /Action/call   —— 校验会话 Cookie（无效回 Result=10014），
//	                       记录每次调用，并按注册的分发器返回数据。
//
// 业务包测试的标准用法（表驱动）：
//
//	ts := testserver.New(t)
//	ts.SetResponse("dnat", "show", testserver.OK(map[string]any{"total": 1, "data": [...]})
//	defer ts.Close()
//	c := ts.Client()
//	...调用业务包方法，用 ts.CallsFor("dnat") 断言发出的参数...
package testserver

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"

	"github.com/sumce/ikuai-sdk-go"
)

// 测试路由器的默认账号密码。
const (
	TestUsername = "admin"
	TestPassword = "admin123"
)

// CallRecord 记录一次 /Action/call 请求。
type CallRecord struct {
	Method   string
	Path     string
	FuncName string
	Action   string
	Param    json.RawMessage
	Cookie   string
	Body     []byte
}

// HandlerFunc 是按 func_name 注册的动态分发器：
// 返回 (Data 值, nil) 时响应 {"Result":30000,"ErrMsg":"Success","Data":<值>}；
// 返回非 nil error 时，若为 *ikuai.APIError 则以其中 Code/ErrMsg 响应，
// 否则以 Result=90000 响应。
type HandlerFunc func(action string, param json.RawMessage) (any, error)

// Server 是假路由器。
type Server struct {
	*httptest.Server

	mu        sync.Mutex
	sessKey   string
	handlers  map[string]HandlerFunc
	static    map[string]any // "func:action" -> 响应 Data
	calls     []CallRecord
	strict    bool // true：未注册的 func_name 直接报错，测试立即暴露
	authCheck bool // true：校验会话 Cookie
}

var sessKeyRe = regexp.MustCompile(`sess_key=([0-9a-fA-F]+)`)

// New 创建并启动假路由器。默认账号 testserver.TestUsername /
// testserver.TestPassword，严格模式 + 会话校验开启。
func New() *Server {
	s := &Server{
		handlers:  map[string]HandlerFunc{},
		static:    map[string]any{},
		strict:    true,
		authCheck: true,
	}
	s.sessKey = newSessKey()
	s.Server = httptest.NewServer(http.HandlerFunc(s.serveHTTP))
	return s
}

// Client 返回指向本假路由器的 *ikuai.Client（未登录）。
func (s *Server) Client(opts ...ikuai.Option) *ikuai.Client {
	c, err := ikuai.NewClient(s.URL, TestUsername, TestPassword, opts...)
	if err != nil {
		panic(fmt.Sprintf("testserver: 构造 Client 失败: %v", err))
	}
	return c
}

// SetResponse 注册 func/action 的静态 Data 响应（Result=30000）。
func (s *Server) SetResponse(funcName, action string, data any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.static[funcName+":"+action] = data
}

// HandleFuncName 注册动态分发器，一个 func_name 的所有 action 都走它。
func (s *Server) HandleFuncName(funcName string, h HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[funcName] = h
}

// Strict 控制未注册 func_name 的行为；默认 true（测试立即失败暴露）。
func (s *Server) Strict(strict bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.strict = strict
}

// AuthCheck 控制是否校验会话 Cookie；默认 true。关闭后 /Action/call
// 不再要求登录（仅用于极端边角测试）。
func (s *Server) AuthCheck(check bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authCheck = check
}

// ExpireSessions 使当前所有会话失效（下一次调用得到 Result=10014，
// 用于测试自动重登）。
func (s *Server) ExpireSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessKey = newSessKey()
}

// Calls 返回全部 /Action/call 记录（拷贝）。
func (s *Server) Calls() []CallRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]CallRecord, len(s.calls))
	copy(out, s.calls)
	return out
}

// CallsFor 返回指定 func_name 的调用记录。
func (s *Server) CallsFor(funcName string) []CallRecord {
	var out []CallRecord
	for _, c := range s.Calls() {
		if c.FuncName == funcName {
			out = append(out, c)
		}
	}
	return out
}

// Reset 清空调用记录（保留注册的分发器）。
func (s *Server) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = nil
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/Action/login" && r.Method == http.MethodPost:
		s.handleLogin(w, r)
	case r.URL.Path == "/Action/call" && r.Method == http.MethodPost:
		s.handleCall(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Passwd   string `json:"passwd"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	sum := md5.Sum([]byte(TestPassword))
	if req.Username != TestUsername || req.Passwd != hex.EncodeToString(sum[:]) {
		writeJSON(w, map[string]any{"Result": 10001, "ErrMsg": "账号或密码错误"})
		return
	}
	s.mu.Lock()
	key := newSessKey()
	s.sessKey = key
	s.mu.Unlock()
	w.Header().Add("Set-Cookie", "sess_key="+key+"; path=/;")
	writeJSON(w, map[string]any{"Result": 10000, "ErrMsg": "Success"})
}

func (s *Server) handleCall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FuncName string          `json:"func_name"`
		Action   string          `json:"action"`
		Param    json.RawMessage `json:"param"`
	}
	body := readAll(r)
	_ = json.Unmarshal(body, &req)
	s.mu.Lock()
	s.calls = append(s.calls, CallRecord{
		Method: r.Method, Path: r.URL.Path,
		FuncName: req.FuncName, Action: req.Action,
		Param: req.Param, Cookie: r.Header.Get("Cookie"), Body: body,
	})
	authOK := !s.authCheck
	if !authOK {
		if m := sessKeyRe.FindStringSubmatch(r.Header.Get("Cookie")); m != nil {
			authOK = m[1] == s.sessKey
		}
	}
	static, hasStatic := s.static[req.FuncName+":"+req.Action]
	handler, hasHandler := s.handlers[req.FuncName]
	strict := s.strict
	s.mu.Unlock()

	if !authOK {
		writeJSON(w, map[string]any{"Result": 10014, "ErrMsg": "no login authentication"})
		return
	}
	if hasStatic {
		writeJSON(w, envelope(30000, "Success", static))
		return
	}
	if hasHandler {
		data, err := handler(req.Action, req.Param)
		if err != nil {
			if ae, ok := err.(*ikuai.APIError); ok {
				writeJSON(w, map[string]any{"Result": ae.Code, "ErrMsg": ae.ErrMsg})
				return
			}
			writeJSON(w, map[string]any{"Result": 90000, "ErrMsg": err.Error()})
			return
		}
		writeJSON(w, envelope(30000, "Success", data))
		return
	}
	if strict {
		writeJSON(w, map[string]any{
			"Result": 90001,
			"ErrMsg": fmt.Sprintf("testserver: func_name %q 未注册（测试需先 SetResponse/HandleFuncName）", req.FuncName),
		})
		return
	}
	writeJSON(w, envelope(30000, "Success", map[string]any{}))
}

func envelope(code int, msg string, data any) map[string]any {
	return map[string]any{"Result": code, "ErrMsg": msg, "Data": data}
}

// OK 是便捷构造：一个 Result=30000 的完整响应体（不走 Data 包装时用）。
func OK(data any) map[string]any { return envelope(30000, "Success", data) }

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	_ = json.NewEncoder(w).Encode(v)
}

func readAll(r *http.Request) []byte {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil
	}
	_ = r.Body.Close()
	return body
}

func newSessKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
