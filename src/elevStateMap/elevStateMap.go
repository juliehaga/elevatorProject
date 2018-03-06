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
	OT_OrderPlaced            = 1
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
	IDLE bool
	Orders [config.NUM_FLOORS][config.NUM_BUTTONS] OrderType
}


type ElevStateMap [config.NUM_ELEVS]ElevInfo




func InitElevStateMap(){
	for e:= 0; e < config.NUM_ELEVS; e++{
		LocalMap[e].CurrentFloor = 3
		LocalMap[e].IDLE = true
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
	LocalMap[config.My_ID].CurrentFloor = elevio.GetFloor()

	LocalMap[config.My_ID].Connected = true
	fmt.Printf("init elev state MAp\n")
}


func GetLocalMap() ElevStateMap{
	mutex.Lock()
	currentMap := LocalMap
	mutex.Unlock()
	return currentMap
}


func UpdateLocalMap(changedMap ElevStateMap){
	floorWithOpenDoor := -1


	LocalMap[config.My_ID].CurrentFloor = changedMap[config.My_ID].CurrentFloor
	LocalMap[config.My_ID].CurrentDir = changedMap[config.My_ID].CurrentDir
	LocalMap[config.My_ID].Connected = changedMap[config.My_ID].Connected
	LocalMap[config.My_ID].Door = changedMap[config.My_ID].Door

	for e:= 0; e < config.NUM_ELEVS; e++{
		if changedMap[e].Door == true{
			floorWithOpenDoor = changedMap[e].CurrentFloor
		}
		LocalMap[e].Connected = changedMap[e].Connected
		for f:= 0; f < config.NUM_FLOORS; f++{
			LocalMap[config.My_ID].Orders[f][elevio.BT_Cab] = changedMap[config.My_ID].Orders[f][elevio.BT_Cab]
			for b:= elevio.BT_HallUp; b < elevio.BT_Cab; b++{

				if changedMap[e].Orders[f][b] == OT_OrderPlaced && LocalMap[config.My_ID].Orders[f][b] == OT_NoOrder{
						LocalMap[e].Orders[f][b] = changedMap[config.My_ID].Orders[f][b]
					} else if changedMap[e].Orders[f][b] == OT_NoOrder && LocalMap[config.My_ID].Orders[f][b] == OT_OrderPlaced && floorWithOpenDoor == f{
						LocalMap[e].Orders[f][b] = changedMap[config.My_ID].Orders[f][b]
					}
			}
		}
	}

}


func UpdateMapFromNetwork(recievedMap ElevStateMap, newOrderChan chan elevio.ButtonEvent, buttonLampChan chan elevio.ButtonLamp){
	floorWithOpenDoor := -1
	fmt.Print("MAP I ORDRE FROM NETWORK\n\n")
	//PrintMap(recievedMap)
	for e:= 0; e < config.NUM_ELEVS; e++{
		if recievedMap[e].Door == true{
			floorWithOpenDoor = recievedMap[e].CurrentFloor
		}

		//sjekk om heis e er i live
		if recievedMap[e].Connected == true{

			if e != config.My_ID {
				LocalMap[e].CurrentFloor = recievedMap[e].CurrentFloor
				LocalMap[e].CurrentDir = recievedMap[e].CurrentDir
				LocalMap[e].Door = recievedMap[e].Door
			}
			for f:= 0; f < config.NUM_FLOORS; f++{
				for b:= elevio.BT_HallUp; b < elevio.BT_Cab; b++{
					if recievedMap[e].Orders[f][b] == OT_OrderPlaced && LocalMap[e].Orders[f][b] == OT_NoOrder{
						newOrderChan <- elevio.ButtonEvent{f, b}
						fmt.Printf("Order from network\n\n")
						LocalMap[e].Orders[f][b] = OT_OrderPlaced
					} else if recievedMap[e].Orders[f][b] == OT_NoOrder && LocalMap[e].Orders[f][b] == OT_OrderPlaced && floorWithOpenDoor == f{
						fmt.Printf("Ordered completed from netowrk floor %v, button %v\n", floor, button)
						//clear orders from all elevators

						buttonLampChan <- elevio.ButtonLamp{f, b, false}
						for elev := 0; elev < config.NUM_ELEVS; elev++{

							LocalMap[elev].Orders[f][b] = OT_NoOrder
						}
					}
				}
			}
		}
	}
}





func PrintMap(elevMap ElevStateMap){
	for e := 0; e < config.NUM_ELEVS; e++ {
		fmt.Printf("\n \n \nSTATE MAP FOR ELEV %v\n", e)
		fmt.Printf("Current floor: %v \n", elevMap[e].CurrentFloor)
		fmt.Printf("Current dir: %v \n", elevMap[e].CurrentDir)
		fmt.Printf("Connected: %v \n", elevMap[e].Connected)
		fmt.Printf("Door: %v \n", elevMap[e].Door)
		
		for f:= 0; f < config.NUM_FLOORS; f++{
			for b:= 0; b < config.NUM_BUTTONS; b++{
				fmt.Printf("%v", elevMap[e].Orders[f][b])
				fmt.Printf("\t")
			}
			fmt.Printf("\n")
		}
	}

}