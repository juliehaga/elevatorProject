package elevStateMap

import(
	"../config"
	"../elevio"	
	"fmt"
	"sync"
)


var LocalMap ElevStateMap
var mutex = &sync.Mutex{}


type OrderType int
const (
	OT_NoOrder                = 0
	OT_OrderExists            = 1
	OT_OrderAccepted          = 2
	OT_OrderCompleted		  = 3
)

type ElevDir int

const (
	ED_Up   		ElevDir = 1
	ED_Down                 = -1
)



type ElevInfo struct{
	CurrentFloor int
	CurrentDir ElevDir
	Connected bool
	Door bool
	Orders [config.NUM_FLOORS][config.NUM_BUTTONS] OrderType
}


type ElevStateMap [config.NUM_ELEVS]ElevInfo




func InitElevStateMap(){
	for e:= 0; e < config.NUM_ELEVS; e++{
		LocalMap[e].CurrentFloor = 0
		LocalMap[e].CurrentDir = ED_Down
		LocalMap[e].Connected = false
		LocalMap[e].Door = false
			
		for f := 0; f < config.NUM_FLOORS; f++{
			for b :=0; b < config.NUM_BUTTONS; b++{
				LocalMap[e].Orders[f][b] = OT_NoOrder
			
			}
		}

		LocalMap[e].Orders[0][elevio.BT_HallDown] = -1
		LocalMap[e].Orders[3][elevio.BT_HallUp] = -1
	}
}


func GetLocalMap() ElevStateMap{
	mutex.Lock()
	currentMap := LocalMap
	mutex.Unlock()
	return currentMap
}


func UpdateLocalMap(changedMap ElevStateMap) ElevStateMap{
	LocalMap[config.My_ID].CurrentFloor = changedMap[config.My_ID].CurrentFloor
	LocalMap[config.My_ID].CurrentDir = changedMap[config.My_ID].CurrentDir
	LocalMap[config.My_ID].Connected = changedMap[config.My_ID].Connected
	LocalMap[config.My_ID].Door = changedMap[config.My_ID].Door
	for f:= 0; f < config.NUM_FLOORS; f++{
		for b:= 0; b < config.NUM_BUTTONS; b++{
			LocalMap[config.My_ID].Orders[f][b] = changedMap[config.My_ID].Orders[f][b]
		}
	}
	return LocalMap
}

func PrintMap(){
	
	for e := 0; e < config.NUM_ELEVS; e++ {
		fmt.Printf("\n \n \nSTATE MAP FOR ELEV %v\n", e)
		fmt.Printf("Current floor: %v \n", LocalMap[e].CurrentFloor)
		fmt.Printf("Current dir: %v \n", LocalMap[e].CurrentDir)
		fmt.Printf("Connected: %v \n", LocalMap[e].Connected)
		fmt.Printf("Door: %v \n", LocalMap[e].Door)
		
		for f:= 0; f < config.NUM_FLOORS; f++{
			for b:= 0; b < config.NUM_BUTTONS; b++{
				fmt.Printf("%v", LocalMap[e].Orders[f][b])
				fmt.Printf("\t")
			}
			fmt.Printf("\n")
		}
	}

}