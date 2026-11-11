// Package overview 实现爱快 Web 控制台「系统概况」大类。
//
// 系统概况是控制台首页仪表盘（无子菜单），对应固件 func_name=homepage，
// action=show，通过 TYPE 选择器获取 sysstat（CPU/内存/在线用户/流量/
// 版本）与 ac_status（AP 统计）数据块。
//
// 开发状态：已实现（系统概况 大类 1 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package overview
