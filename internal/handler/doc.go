// Package handler 承载业务语义的最顶层：解析「这条消息要做什么」，编排 service / room 等。
// 下层 game 只负责收发与按 route 分发，不实现具体业务。
package handler
