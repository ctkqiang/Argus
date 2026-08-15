// Package event 实现 AIEvent 审计事件的持久化控制器，接收网关上报的事件流，
// 写入数据库/对象存储并提供检索 API。支持事件分级归档与合规保留策略。
package event
