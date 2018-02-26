package elevStateMap

import(
	"../config"
	
	
)


var LocalMap ElevStateMap


type OrderType int
const (
	OT_NoOrder                = 0
	OT_OrderExists            = 1
	OT_OrderAccepted          = 2
	OT_OrderServed			  = 3
)

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ElevInfo struct{
	CurrentFloor int
	CurrentDir MotorDirection
	Connected bool
	Door bool
	Orders [config.NUM_FLOORS][config.NUM_BUTTONS] OrderType
}


type ElevStateMap [config.NUM_ELEVS]ElevInfo




func InitElevStateMap(){
	for e:= 0; e < config.NUM_ELEVS; e++{
		LocalMap[e].CurrentFloor = 0
		LocalMap[e].CurrentDir = MD_Up
		LocalMap[e].Connected = true
		LocalMap[e].Door = false
			
		for f := 0; f < config.NUM_FLOORS; f++{
			for b :=0; b < config.NUM_BUTTONS; b++{
				LocalMap[e].Orders[f][b] = OT_NoOrder
			
			}
		}
	}
}


func GetLocalMap() ElevStateMap{
	//vil vi få problemer med at map forandres mens vi returnerer det
	return LocalMap
}



	
