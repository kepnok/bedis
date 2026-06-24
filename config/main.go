package config

var Host string = "0.0.0.0"
var Port int = 7379
var KeysLimit int = 5

var EvictionStratery string = "simple-first"
var AOFFile string = "./bedis-main.aof"