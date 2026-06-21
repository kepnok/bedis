package core

type BedisCmd struct {
	Cmd string
	Args []string
}

type BedisCmds []*BedisCmd