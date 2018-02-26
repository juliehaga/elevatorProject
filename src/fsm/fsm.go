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
	elevStateMap.ClearOrder(orderedFloor)
}

//burde vi ha en channel med nye ordre. Dette for trigge fsm


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


func eventButtonPressed(orderedFloor int, button elevio.ButtonType){
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
			if(currentMap[config.My_ID].CurrentFloor == orderedFloor){
				//restart timer()
				//ta høyde for at det også må være riktig retning
			}


 
	}
}


func 
/*
func eventFloorReached(){

}


func costFunc{

}


func disconnectedElevator(){

}
*/

