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
package ikuai
