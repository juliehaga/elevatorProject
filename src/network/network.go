package network 

import (
	"./broadcast"
	"./localip"
	"./peers"
)


type ElevMapMsg struct {
	ID int
	elevMap   elevStateMap
}


func init()