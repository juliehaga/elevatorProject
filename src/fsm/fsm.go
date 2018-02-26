package fsm

include(
	"../elevStateMap"
	"../elevio"
)

var state ElevState


type ElevState int
const(
	IDLE            = 0
	MOVING          = 1
	FLOOR_REACHED   = 2
)




func fsm{
	switch case

}

func eventOrderAccepted(){
	//order accepted on this elevator

	//velg retning
		//state=MOVING
}

func eventButtonPressed(orderedFloor int, button ButtonType){
	currentMap := GetLocalMap()
	switch(state){
		case IDLE: 
			if currentMap.CurrentFloor == orderedFloor {
				state = FLOOR_REACHED}
			else {
				if orderedFloor > currentMap.CurrentFloor{
					elevio.SetMotorDirection(MD_Up)	
				}else{
					elevio.SetMotorDirection(MD_Down)
				}
				state = MOVING
				
			}
			break
			//må vi breake case?
		case MOVING:
			//sjekk at vi når riktig etasje
			
			
			//gets a new order when
			 
			
		
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

