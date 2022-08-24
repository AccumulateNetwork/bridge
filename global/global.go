package global

import "github.com/AccumulateNetwork/bridge/schema"

var IsOnline bool                // is bridge is online or paused
var IsLeader bool                // if current node is a leader
var IsAudit bool                 // if current node is an audit
var LeaderDuration int64         // number of checks this node is a leader
var Tokens schema.Tokens         // slice of tokens
var BridgeFees schema.BridgeFees // slice of bridge fees
