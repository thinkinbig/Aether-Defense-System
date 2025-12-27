// Package mq provides message queue configuration and client wrappers.
package mq

// Config represents RocketMQ configuration.
type Config struct {
	// NameServer addresses (comma-separated or semicolon-separated)
	// Example: "127.0.0.1:9876" or "127.0.0.1:9876;192.168.1.1:9876"
	NameServer string `json:"nameServer" yaml:"nameServer"`
	// Producer group name
	Group string `json:"group" yaml:"group"`
	// Topic name for messages
	Topic string `json:"topic" yaml:"topic"`
	// Retry times for sending messages (default: 2)
	RetryTimes int `json:"retryTimes,omitempty" yaml:"retryTimes,omitempty"`
	// Timeout for sending messages in milliseconds (default: 3000)
	SendTimeout int `json:"sendTimeout,omitempty" yaml:"sendTimeout,omitempty"`
}
