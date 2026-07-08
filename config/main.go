package config

var Host string = "0.0.0.0"
var Port int = 7379
var KeysLimit int = 100

var EvictionRatio float64 = 0.04

var EvictionStratery string = "allkeys-lru"
var AOFFile string = "./bedis-main.aof"