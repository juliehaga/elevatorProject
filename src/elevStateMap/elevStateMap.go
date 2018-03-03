package elevStateMap

import(
	"../config"
	"../elevio"	
)


var LocalMap ElevStateMap


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
		LocalMap[e].Connected = true
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
	//vil vi få problemer med at map forandres mens vi returnerer det
	return LocalMap
}

func ClearOrder(floor int){
	LocalMap[config.My_ID].Orders[floor][elevio.BT_Cab] = OT_OrderCompleted

	if LocalMap[config.My_ID].CurrentDir == ED_Up{
		LocalMap[config.My_ID].Orders[floor][elevio.BT_HallUp] = OT_OrderCompleted
	} else {
		LocalMap[config.My_ID].Orders[floor][elevio.BT_HallDown] = OT_OrderCompleted
	}
}

