// Package mq provides a unified abstraction for message queue components.
//
// Key components:
//   - message.go: Defines message attributes (Topic, Payload, Extra, Tag, etc.)
//     enabling different MQ drivers to implement specific features via composition.
//   - pubsub.go: Defines standard roles including Consumer, Producer, Observer,
//     and Factory.
//   - resource.go: Provides universal resource management, allowing resources
//     created by the Factory to be managed through a centralized ResourceManager.
package mq
