// Package confws 提供可配置的 WebSocket 服务端:
//
//   - 服务配置: Endpoint / Option(监听、超时、TLS、容量等)
//   - 客户端管理: Client / ClientManager(登记、摘表、踢除、关闭)
//   - 回调接口(用法与 error 语义见各 Handler 类型注释):
//   - Connection: Upgrade 前门禁;error → HTTP 401、不建 WS
//   - Establish:  登记后业务握手;error → Close Client、不进消息循环
//   - Message 及读写失败、断开钩子
package confws
