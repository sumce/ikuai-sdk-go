// Package ikuai is a full-featured Go SDK for the iKuaiOS (爱快) router
// web API.
//
// iKuaiOS exposes its management plane through a small set of HTTP
// endpoints that are shared by the official web UI:
//
//	POST /Action/login   JSON body {"username","passwd","pass","remember_password"}
//	POST /Action/call    JSON body {"func_name","action","param"}
//	GET  /Action/tag     UI page beacon (id=...)
//	GET  /Action/rtcollect/websocket  realtime traffic stream
//
// Every management feature of the web console (network, NAT, QoS,
// firewall, VPN, AC/wireless, auth accounting, monitoring, tools, ...)
// is a "func_name" invoked through /Action/call with an "action" such
// as show/add/edit/del/del_all/set.
//
// # Quick start
//
//	c, err := ikuai.NewClient("http://192.168.1.1", "admin", "password")
//	if err != nil {
//		log.Fatal(err)
//	}
//	ctx := context.Background()
//	if err := c.Login(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer c.Logout(context.WithoutCancel(ctx))
//
//	// Typed helper for verified endpoints.
//	home, err := c.Monitor.Homepage(ctx, "sysstat")
//
//	// Generic CRUD for every func_name in the web console.
//	page, err := c.NAT.DNAT().List(ctx, ikuai.DefaultListOptions())
//
//	// Escape hatch: any func_name/action/param combination.
//	resp, err := c.Call(ctx, "dnat", "show", map[string]any{"TYPE": "total,data", "limit": "0,20"})
//
// A login session is maintained automatically: calls that fail with the
// "no login authentication" error (Result 10014) trigger a transparent
// re-login and one retry.
//
// # Guarantees
//
//   - Concurrency: *Client and every service built on it are safe for
//     concurrent use by multiple goroutines. The session is mutex
//     guarded and re-login is single-flight. Each Client instance is
//     fully independent — create as many as needed.
//   - No global state: the SDK never mutates http.DefaultClient, the
//     environment or any process-wide setting, and library code never
//     panics; every failure is returned as an error value.
//   - Zero dependencies: the module depends only on the Go standard
//     library, which keeps the supply chain minimal.
//   - Errors: every non-success API result is a structured *APIError.
//     Use errors.As(err, &ae) to inspect it and errors.Is with a code
//     target for programmatic matching, e.g.
//     errors.Is(err, &ikuai.APIError{Code: ikuai.CodeNoLoginAuth}).
//     Sentinel ikuai.ErrInvalidParam marks invalid arguments and is
//     always wrapped with %w so the original chain survives.
package ikuai
