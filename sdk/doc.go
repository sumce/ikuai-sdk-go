// Package sdk 是 ikuai-sdk 的门面聚合包：一次性构造全部 12 个业务
// 服务，同时各业务包也可独立构造（按需引入）。
//
//	router, err := sdk.New(&sdk.Config{
//		BaseURL:  "http://192.168.1.1",
//		Username: "admin",
//		Password: "admin123",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	ctx := context.Background()
//	if err := router.Login(ctx); err != nil {
//		log.Fatal(err)
//	}
//	ov, _ := router.Overview.SystemOverview(ctx)
//
// 开发状态：已实现 —— 聚合全部 12 个业务包。
package sdk
