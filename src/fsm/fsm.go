package fsm

import(
	"../elevStateMap"
	"../elevio"
	"../config"
	"time"
)

var state ElevState

type ElevState int
const(
	IDLE            = 0
	MOVING          = 1
	DOOR_OPEN	    = 2
)

const DOOR_TIME 	= 2


func Fsm(motorChan chan elevio.MotorDirection, doorLampChan chan bool, floorChan chan int){
	doorTimer := time.NewTimer(time.Second * DOOR_TIME)
	doorTimer.Stop()
	//go harware(motorChan, doorLampChan)
	a:= 2
	a += 2
	for{
		select{
		case  <- floorChan:
			eventNewFloor(motorChan, doorLampChan, doorTimer)
		//case <- buttonChan:
			//eventNewOrder()
		case <- doorTimer.C:
			eventDoorTimeout(doorLampChan)

		}

	}
}


func eventNewFloor(motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	//sjekker at vi når nye etasjer
	//I moving må vi sjekke om vi skal stoppe 

	//send melding ikke bruk funksjoner
	//elevStateMap.ClearOrder(orderedFloor)
	switch(state){
		case MOVING:
			if(OrderOnFloor(currentMap)){
				motorChan <- elevio.MD_Stop
				doorLampChan <- true
				doorTimer.Reset(time.Second * DOOR_TIME)
				
				//gi beskjed om at ordre er utført -> dette handler buttonlamp for alle heiser
				state = DOOR_OPEN
			}
			
	}
}

func eventDoorTimeout(doorLampChan chan bool){

	switch(state){
		case DOOR_OPEN:
			doorLampChan <- false
			//slå av lyset
			//orderFinished   -> dette burde slukkle alle ly i hele floor
			state = IDLE
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
			//wtf is ordered floor?
			if currentMap[config.My_ID].CurrentFloor == orderedFloor {
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


func OrderOnFloor(elevMap elevStateMap.ElevStateMap) bool{
	switch elevMap[config.My_ID].CurrentDir{

		case elevStateMap.ED_Up:
			//order on current floor and no orders above
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp]==elevStateMap.OT_OrderAccepted || elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab]==elevStateMap.OT_OrderAccepted {
				return true
			} else if !ordersAbove(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown]==elevStateMap.OT_OrderAccepted{
				return true }
			break
		case elevStateMap.ED_Down:
			//order on current floor and no orders below
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

