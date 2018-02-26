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
	FLOOR_REACHED   = 2
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

func eventButtonPressed(orderedFloor int, button elevio.ButtonType){
	currentMap := elevStateMap.GetLocalMap()
	switch(state){
		case IDLE: 
			if currentMap[config.My_ID].CurrentFloor == orderedFloor {
				state = FLOOR_REACHED
			} else {
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
			if currentMap[config.My_ID].CurrentFloor == orderedFloor {
				state = FLOOR_REACHED
			}
			break
		case FLOOR_REACHED:
			//start_timer
			//CLEAR QUEUE?
			//turn off lights?

			elevio.SetMotorDirection(elevio.MD_Stop)
			elevStateMap.ClearOrder(orderedFloor)	
			//start timer	 
	}
}

/*
func eventFloorReached(){

}


func costFunc{

}


func disconnectedElevator(){

}
*/

