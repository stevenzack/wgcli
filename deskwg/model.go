package main

type Status int

const (
	EmptyAccessKey Status = iota
	Disconnected
	Connecting
	Connected 
	Deleting
)

const (
	HongKong  = "香港"
	Singapore = "新加坡"
	German    = "德国"
	Korea     = "韩国"
	Japan     = "日本"
)

var nameToRegion = map[string]string{
	HongKong:  "hk",
	Singapore: "sg",
	German:    "de",
	Korea:     "kr",
	Japan:     "jp",
}
