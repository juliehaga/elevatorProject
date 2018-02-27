package fsm

import(
	"../elevStateMap"
	"../elevio"
	"../config"
)

var state ElevState


type ElevState int
const(
	IDLE            = 0
	MOVING          = 1
	DOOR_OPEN	    = 2
)




func Fsm(){
	a:= 2
	a += 2

}

func eventOrderAccepted(){
	//order accepted on this elevator

	//velg retning
		//state=MOVING
}

func eventNewFloor(){
	//sjekker at vi når nye etasjer
	//I moving må vi sjekke om vi skal stoppe 

	//send melding ikke bruk funksjoner
	elevio.SetMotorDirection(elevio.MD_Stop)
	//elevStateMap.ClearOrder(orderedFloor)
	switch(state){
		case MOVING:
			
	}
}

//burde vi ha en channel med nye ordre. Dette for trigge fsm

/*
func eventDoorTimeout{


			if !timerActive{
				timerActive = true
				timer = time.Now()
			}
			else {doorTimer()}
			//start timer	
}

	func doorTimer(){
	if timer -timer.now() >= 2000{
		state = IDLE
		timerActive = false
	}
}
*/

func eventNewOrder(orderedFloor int, button elevio.ButtonType){
	currentMap := elevStateMap.GetLocalMap()
	switch(state){
		case IDLE: 
			if currentMap[config.My_ID].CurrentFloor == orderedFloor {
				//start timeout
				state = DOOR_OPEN

				//her bare tar vi bestillingen, må varsle? 

			} else {
				//send msg to hardware about dir
				if orderedFloor > currentMap[config.My_ID].CurrentFloor{
					elevio.SetMotorDirection(elevio.MD_Up)	
				}else{
					elevio.SetMotorDirection(elevio.MD_Down)
				}

				state = MOVING
			}
			break
			//må vi breake case?
		case MOVING:
			break
		case DOOR_OPEN:
			
	}
}


func orderedFloorReached(elevMap elevStateMap.ElevStateMap) bool{
	switch elevMap[config.My_ID].CurrentDir{

		case elevStateMap.ED_Up:
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp]==elevStateMap.OT_OrderAccepted || elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab]==elevStateMap.OT_OrderAccepted {
				return true
			} else if !ordersAbove(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown]==elevStateMap.OT_OrderAccepted{
				return true }
			break
		case elevStateMap.ED_Down:
		 	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown]==elevStateMap.OT_OrderAccepted || elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab]==elevStateMap.OT_OrderAccepted {
		 		return true
			} else if !ordersBelow(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp]==elevStateMap.OT_OrderAccepted {
			 	return true }
			break
		default:
	}
	return false
}

func ordersAbove(elevMap elevStateMap.ElevStateMap) bool{
	for i := elevMap[config.My_ID].CurrentFloor + 1; i<config.NUM_FLOORS; i++{
		for j := elevio.BT_HallUp; j<= elevio.BT_Cab; j++{ 
			return elevMap[config.My_ID].Orders[i][j] == elevStateMap.OT_OrderAccepted
		}
	}
	return false
}

func ordersBelow(elevMap elevStateMap.ElevStateMap) bool{
	for i := elevMap[config.My_ID].CurrentFloor - 1; i>=0; i--{
		for j := elevio.BT_HallUp; j<= elevio.BT_Cab; j++{ 
			return elevMap[config.My_ID].Orders[i][j] == elevStateMap.OT_OrderAccepted
		}
	}
	return false
}
/*
func eventFloorReached(){

}


func costFunc{

}


func disconnectedElevator(){

}
*/

