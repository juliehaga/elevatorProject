package network 

import (
	"../elevStateMap"
	"../config"
	"time"
)


type ElevMapMsg struct {
	ID int
	elevMap   elevStateMap.ElevStateMap
}



func testSendfunc(elevMapTx chan ElevMapMsg) {
		currentMap := elevStateMap.GetLocalMap()
		elevMapMsg := ElevMapMsg{config.My_ID, currentMap}
		for {
			elevMapTx <- elevMapMsg
			time.Sleep(1 * time.Second)
		}
}