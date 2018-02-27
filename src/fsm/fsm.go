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


func Fsm(motorChan chan elevio.MotorDirection, doorLampChan chan bool, floorChan chan int, buttonLampChan chan elevio.ButtonLamp){
	doorTimer := time.NewTimer(time.Second * DOOR_TIME)
	doorTimer.Stop()
	//go harware(motorChan, doorLampChan)
	a:= 2
	a += 2
	for{
		select{
		case  <- floorChan:
			eventNewFloor(motorChan, doorLampChan, doorTimer)
		//case <- ackOrderChan:
			//denne må trigges av ???
			//eventNewAckOrder()
		case <- doorTimer.C:
			eventDoorTimeout(doorLampChan)

//		case <- buttonChan: 
//			eventLocalOrder
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
		case DOOR_OPEN:
			
	}
}

func eventDoorTimeout(doorLampChan chan bool){

	switch(state){
		case DOOR_OPEN:
			doorLampChan <- false
			//slå av buttonlamp
			//orderFinished   -> dette burde slukkle alle lys i hele floor
			state = IDLE
	}
}


func eventNewAckOrder(buttonLampChan chan elevio.ButtonLamp, motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	//antar at lyset allerede er tent, dette må gjøres et annet sted

	switch(state){
		case IDLE:
			if orderInThisFloor(currentMap) {
				
				//Gi beskjed om at ordre utført
				doorLampChan <- true
				doorTimer.Reset(time.Second * DOOR_TIME)
				state = DOOR_OPEN
			}else{
				motorDir := chooseDirection(currentMap)
				if motorDir != elevio.MD_Stop {
					motorChan <- motorDir
					state = MOVING
				}
			}

			
		case MOVING:

			break
		case DOOR_OPEN:
			//if i etasje -> restart doorOpen
			
			
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

func chooseDirection(elevMap elevStateMap.ElevStateMap) elevio.MotorDirection{

	switch elevMap[config.My_ID].CurrentDir{

		case elevStateMap.ED_Up:
		
			for f := elevMap[config.My_ID].CurrentFloor+1; f < config.NUM_FLOORS; f++{
				if elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderAccepted || nearestElevator(elevMap, f) {
					return elevio.MD_Up
				}
			}

			for f := elevMap[config.My_ID].CurrentFloor-1; f>= 0; f-- {
				if elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderAccepted || nearestElevator(elevMap, f) {
					return elevio.MD_Down
				}
			}	
			break
		case elevStateMap.ED_Down:

			for f := elevMap[config.My_ID].CurrentFloor-1; f>= 0; f-- {
				if elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderAccepted || nearestElevator(elevMap, f) {
					return elevio.MD_Down
				}
			}	

			for f := elevMap[config.My_ID].CurrentFloor+1; f < config.NUM_FLOORS; f++{
				if elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderAccepted || nearestElevator(elevMap, f) {
					return elevio.MD_Up
				}
			}

			
			break
			
	}
	return elevio.MD_Stop
}


func nearestElevator(elevMap elevStateMap.ElevStateMap, floor int) bool{
 	return true
}

func orderInThisFloor( elevMap elevStateMap.ElevStateMap) bool{
	for b := elevio.BT_HallUp; b <= elevio.BT_Cab; b++ {
		if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][b] == elevStateMap.OT_OrderAccepted {
			return true
		}
	}
	return false
}