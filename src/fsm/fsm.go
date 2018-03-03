package fsm

import(
	"../elevStateMap"
	"../elevio"
	"../config"
	"time"
	"math"
)

var state ElevState

type ElevState int
const(
	IDLE            = 0
	MOVING          = 1
	DOOR_OPEN	    = 2
)

const DOOR_TIME 	= 2


func Fsm(motorChan chan elevio.MotorDirection, doorLampChan chan bool, floorChan chan int, buttonLampChan chan elevio.ButtonLamp, mapChangesChan chan elevStateMap.ElevStateMap){
	doorTimer := time.NewTimer(time.Second * DOOR_TIME)
	doorTimer.Stop()
	//go harware(motorChan, doorLampChan)
	a:= 2
	a += 2
	for{
		select{
		case  <- floorChan:
			eventNewFloor(motorChan, doorLampChan, doorTimer, mapChangesChan, buttonLampChan)
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


func eventNewFloor(motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp){
	currentMap := elevStateMap.GetLocalMap()
	switch(state){
		case MOVING:
			if(OrderOnFloor(currentMap)){
				motorChan <- elevio.MD_Stop
				doorLampChan <- true
				doorTimer.Reset(time.Second * DOOR_TIME)

				currentMap[config.My_ID].Door = true;
				orderCompleted(&currentMap, buttonLampChan)
				mapChangesChan <- currentMap
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
	//kanskje vi kan gjenbruke ordersbelow og ordersAbove
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
 	dist := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)));
 	for e := 0; e<config.NUM_ELEVS; e++ {
 		if elevMap[e].CurrentFloor < floor && elevMap[e].CurrentDir == elevStateMap.ED_Up {
 			if elevMap[e].CurrentFloor - floor < dist {
 				return false
 			}

 		}else if elevMap[e].CurrentFloor > floor && elevMap[e].CurrentDir == elevStateMap.ED_Down {
 			if elevMap[e].CurrentFloor < floor && elevMap[e].CurrentDir == elevStateMap.ED_Up {
 				return false
 			}
 		}
 	}
 	for e := 0; e<config.NUM_ELEVS; e++ {
	 	if elevMap[e].CurrentFloor - floor == dist {
	 			return config.My_ID <= e 
	 		}
	}
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

func orderCompleted(elevMap *elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp){
	elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] = elevStateMap.OT_OrderCompleted
	buttonLampChan <- elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_Cab, false}
	switch(elevMap[config.My_ID].CurrentDir){
		case elevStateMap.ED_Up: 
			elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_OrderCompleted
			buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallUp, false}
		case elevStateMap.ED_Down:
			elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_OrderCompleted
			buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
	}
}