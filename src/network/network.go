package network 

import (
	"../elevStateMap"
	"../config"
)


type ElevMapMsg struct {
	ID int
	elevMap   elevStateMap.ElevStateMap
}



func SendElevMap(elevMapTx chan ElevMapMsg, elevMap elevStateMap.ElevStateMap) {
		elevMapMsg := ElevMapMsg{config.My_ID, elevMap}
		elevMapTx <- elevMapMsg
}