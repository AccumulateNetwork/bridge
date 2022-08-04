package global

import "github.com/AccumulateNetwork/bridge/schema"

// minimum amount of checks to consider node as a leader
const LEADER_MIN_DURATION = 2

var IsLeader bool
var LeaderDuration int64
var Tokens schema.Tokens
var BridgeFees schema.BridgeFees
