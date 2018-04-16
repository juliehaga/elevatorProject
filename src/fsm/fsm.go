package fsm

import(
	"../elevStateMap"
	"../config"
	"time"
	"math"
	"fmt"
)

var state ElevState

type ElevState int
const(
	INIT            = 0
	IDLE            = 1
	MOVING          = 2
	DOOR_OPEN	    = 3
	OUT_OF_ORDER    = 4
)

const DOOR_TIME 	    = 2
const IDLE_TIME 	    = 10


func Fsm(motorChan chan config.MotorDirection, doorLampChan chan bool, floorChan chan int, buttonLampChan chan config.ButtonLamp, mapChangesChan chan config.ElevStateMap, newOrderChan chan config.ButtonEvent, orderCompleteChan chan config.ButtonEvent, activeOrderTx chan config.ActiveOrders){
	state = INIT
	
	doorTimer := time.NewTimer(time.Second * DOOR_TIME)
	doorTimer.Stop()
	idleTimer := time.NewTimer(time.Second * IDLE_TIME)

	for{
		select{
		case  floor := <- floorChan:
			eventNewFloor(orderCompleteChan, motorChan, doorLampChan, doorTimer,mapChangesChan, buttonLampChan, floor, idleTimer, activeOrderTx)
			idleTimer.Reset(time.Second * IDLE_TIME)

		case buttonPushed := <- newOrderChan:
			eventNewAckOrder(orderCompleteChan, buttonLampChan, motorChan, doorLampChan, doorTimer, mapChangesChan, buttonPushed, idleTimer, activeOrderTx)
			idleTimer.Reset(time.Second * IDLE_TIME)

		case <- doorTimer.C:
			eventDoorTimeout(doorLampChan, mapChangesChan, idleTimer, motorChan)
			idleTimer.Reset(time.Second * IDLE_TIME)

		case <- idleTimer.C:
			eventIdleTimeout(idleTimer, motorChan, mapChangesChan)
			idleTimer.Reset(time.Second * IDLE_TIME)
		}
	}
}

func eventIdleTimeout(idleTimer *time.Timer, motorChan chan config.MotorDirection, mapChangesChan chan config.ElevStateMap){
	var motorDir config.MotorDirection
	currentMap := elevStateMap.GetLocalMap()
	
	if state != IDLE {
		fmt.Printf("***********************************************OUT OF ORDER*****************************************\n")
		currentMap[config.My_ID].Connected = false
		if currentMap[config.My_ID].CurrentFloor < config.NUM_FLOORS-1{
			motorDir = config.MD_Down
		} else {
			motorDir = config.MD_Up
		}
		//prøv å kjør i hvilken som helst retning
	}else{
		//velg retning basert på bestillinger
		motorDir, currentMap[config.My_ID].CurrentDir = forceChooseDirection(currentMap)
		if motorDir == config.MD_Stop{
			state = IDLE
		} else{
			motorChan <- motorDir
			state = MOVING
		}
	}
	mapChangesChan <- currentMap
}


func eventNewFloor(orderCompleteChan chan config.ButtonEvent, motorChan chan config.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan config.ElevStateMap, buttonLampChan chan config.ButtonLamp, floor int, idleTimer *time.Timer, activeOrderTx chan config.ActiveOrders){
	currentMap := elevStateMap.GetLocalMap()
	//var det en grunn til at vi skulle oppdatere currentfloor her? 
	currentMap[config.My_ID].Connected = true
	currentMap[config.My_ID].CurrentFloor = floor 
	var motorDir config.MotorDirection

	switch(state){
		case INIT:
			if orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap) == false{
				motorDir, currentMap[config.My_ID].CurrentDir = chooseDirection(currentMap)
				if motorDir == config.MD_Stop{
					state = IDLE
				} else{
					motorChan <- motorDir
					state = MOVING
				}
			} else{
				motorChan <- config.MD_Stop
				doorLampChan <- true
				doorTimer.Reset(time.Second * DOOR_TIME)
				currentMap[config.My_ID].Door = true
				currentMap = orderCompleted(currentMap, buttonLampChan, orderCompleteChan, activeOrderTx)
				currentMap[config.My_ID].IDLE = false
				state = DOOR_OPEN
			}

			
			
		case MOVING:
			if shouldStop(currentMap) {
				motorChan <- config.MD_Stop
					if  orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){
						doorLampChan <- true
						doorTimer.Reset(time.Second * DOOR_TIME)
						currentMap[config.My_ID].Door = true
						currentMap = orderCompleted(currentMap, buttonLampChan, orderCompleteChan, activeOrderTx)
						currentMap[config.My_ID].IDLE = false
						state = DOOR_OPEN
					} else {
						currentMap[config.My_ID].IDLE = true
						state = IDLE
					}	
			}

	}
	mapChangesChan <- currentMap	
}

func eventDoorTimeout(doorLampChan chan bool, mapChangesChan chan config.ElevStateMap, idleTimer *time.Timer, motorChan chan config.MotorDirection){
	currentMap := elevStateMap.GetLocalMap()
	var motorDir config.MotorDirection

	switch(state){
		case DOOR_OPEN:
			doorLampChan <- false
			currentMap[config.My_ID].Door = false
			motorDir, currentMap[config.My_ID].CurrentDir = chooseDirection(currentMap)
			motorChan <- motorDir
			if motorDir != config.MD_Stop {
				state = MOVING
			} else {
				state = IDLE
			}
			//fmt.Printf("*******************************DOOR TIMEOUT***************************\n")
			mapChangesChan <- currentMap	
	}
}


func eventNewAckOrder(orderCompleteChan chan config.ButtonEvent, buttonLampChan chan config.ButtonLamp, motorChan chan config.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan config.ElevStateMap, buttonPushed config.ButtonEvent, idleTimer *time.Timer, activeOrderTx chan config.ActiveOrders){
	currentMap := elevStateMap.GetLocalMap()

	fmt.Printf("eventAckOrder er trigget\n")

	//accept CAB order
	if buttonPushed.Button == config.BT_Cab{ //&& currentMap[config.My_ID].OutOfOrder == false{
		currentMap[config.My_ID].Orders[buttonPushed.Floor][buttonPushed.Button] = config.OT_OrderPlaced
		buttonLampChan <- config.ButtonLamp{buttonPushed.Floor, buttonPushed.Button, true}
	}
		


	
	var motorDir config.MotorDirection
	switch(state){
		case IDLE:
			if orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){// && currentMap[config.My_ID].OutOfOrder == false{


				doorLampChan <- true	
				currentMap[config.My_ID].Door = true
				currentMap = orderCompleted(currentMap, buttonLampChan, orderCompleteChan, activeOrderTx)
				doorTimer.Reset(time.Second * DOOR_TIME)
				//fmt.Printf("----------------------------orderInfloor----------------------------\n")
				state = DOOR_OPEN
			}else{
				motorDir, currentMap[config.My_ID].CurrentDir = chooseDirection(currentMap)
				motorChan <- motorDir
				if motorDir != config.MD_Stop {
					state = MOVING
				} 
				//fmt.Printf("----------------------------choosedir----------------------------\n")

			}
		case DOOR_OPEN:
			if orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){
				doorLampChan <- true	
				doorTimer.Reset(time.Second * DOOR_TIME)
				currentMap[config.My_ID].Door = true
				currentMap = orderCompleted(currentMap, buttonLampChan, orderCompleteChan, activeOrderTx)
			}
	}
	mapChangesChan <- currentMap
}



func shouldStop(elevMap config.ElevStateMap) bool{
	switch(state){
		case MOVING: 
			if elevMap[config.My_ID].CurrentFloor == config.NUM_FLOORS-1 && elevMap[config.My_ID].CurrentDir == config.ED_Up{
				return true
			} else if elevMap[config.My_ID].CurrentFloor == 0 && elevMap[config.My_ID].CurrentDir == config.ED_Down {
				return true
			}
	}
	
	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_Cab]==config.OT_OrderPlaced {
		return true
	}

	switch elevMap[config.My_ID].CurrentDir{
		case config.ED_Up:
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp]==config.OT_OrderPlaced {
				return true
			} else if !ordersAbove(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown]==config.OT_OrderPlaced{
				return true 
			} else if !ordersAbove(elevMap){
				return true
			}

		case config.ED_Down:
		 	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown]==config.OT_OrderPlaced {
		 		return true
		 	} else if !ordersBelow(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp]==config.OT_OrderPlaced {
			 	return true 
			} else if !ordersBelow(elevMap){
				return true
			}
	}
	return false
}

func ordersAbove(elevMap config.ElevStateMap) bool{
	for f := elevMap[config.My_ID].CurrentFloor + 1; f<config.NUM_FLOORS; f++{
		if orderInThisFloor(f, elevMap){
			return true
		}
	}
	return false
}

func ordersBelow(elevMap config.ElevStateMap) bool{
	for f := elevMap[config.My_ID].CurrentFloor - 1; f>=0; f--{
		if orderInThisFloor(f, elevMap){
			return true
		}
	}
	return false
}



func orderCompleted(elevMap config.ElevStateMap, buttonLampChan chan config.ButtonLamp, orderCompleteChan chan config.ButtonEvent, activeOrderTx chan config.ActiveOrders) config.ElevStateMap{
	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_Cab] == config.OT_OrderPlaced{
		elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_Cab] = config.OT_NoOrder
		buttonLampChan <- config.ButtonLamp{elevMap[config.My_ID].CurrentFloor, config.BT_Cab, false}
	}

	switch(elevMap[config.My_ID].CurrentDir){
		case config.ED_Up: 
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp] == config.OT_OrderPlaced{
				elevMap = ClearOrder(elevMap, config.ButtonEvent{elevMap[config.My_ID].CurrentFloor, config.BT_HallUp}, buttonLampChan, activeOrderTx)

			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown] == config.OT_OrderPlaced {
				elevMap = ClearOrder(elevMap, config.ButtonEvent{elevMap[config.My_ID].CurrentFloor, config.BT_HallDown}, buttonLampChan, activeOrderTx)		
			}
			
		case config.ED_Down:
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown] == config.OT_OrderPlaced{
				elevMap = ClearOrder(elevMap, config.ButtonEvent{elevMap[config.My_ID].CurrentFloor, config.BT_HallDown}, buttonLampChan, activeOrderTx)				
			
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp] == config.OT_OrderPlaced{
				elevMap = ClearOrder(elevMap, config.ButtonEvent{elevMap[config.My_ID].CurrentFloor, config.BT_HallUp}, buttonLampChan, activeOrderTx)
			}
			
	}
	return elevMap
}




func orderInThisFloor( floor int, elevMap config.ElevStateMap) bool{
	ackElevs := 0 
	if (floor != -1){
		for b := config.BT_HallUp; b <= config.BT_Cab; b++ {
			if b == config.BT_Cab && elevMap[config.My_ID].Orders[floor][b] == config.OT_OrderPlaced{
				return true
			}else{
				for e:= 0; e < config.NUM_ELEVS; e++{
					if elevMap[e].Orders[floor][b] == config.OT_OrderPlaced{
						ackElevs ++;

					}
				}
				if ackElevs == config.NUM_ELEVS{
					return true
				}else{
					ackElevs = 0
				}

			}

		}
	
	}
	return false
}


func chooseDirection(elevMap config.ElevStateMap) (config.MotorDirection, config.ElevDir){
	switch elevMap[config.My_ID].CurrentDir{
		case config.ED_Up: 
			if ordersAbove(elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if  orderInThisFloor(f, elevMap) && (nearestElevator(elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced){
						return config.MD_Up, config.ED_Up
					}
				}
			} else if ordersBelow(elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, elevMap) && (nearestElevator(elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced){
						return config.MD_Down, config.ED_Down
					}
				}
			}	
		case config.ED_Down:
			if ordersBelow(elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, elevMap) && (nearestElevator(elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced){
						return config.MD_Down, config.ED_Down
					}
				}
			} else if ordersAbove(elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if orderInThisFloor(f, elevMap) && (nearestElevator(elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced ){
						return config.MD_Up, config.ED_Up
					}
				}
			}
	}
	return config.MD_Stop, elevMap[config.My_ID].CurrentDir
}



func nearestElevator(elevMap config.ElevStateMap, floor int) bool{

 	myDist := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)))
 	if elevMap[config.My_ID].CurrentFloor < floor { 
	 	for e := 0; e<config.NUM_ELEVS; e++ {
	 		if elevMap[e].Connected == true{ //&& elevMap[e].OutOfOrder == false{	
			 	if e != config.My_ID{	
			 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
			 		if distElev < myDist{
			 			if elevMap[e].CurrentFloor < floor && (elevMap[e].CurrentDir == config.ED_Up){
			 				return false
			 			} else if elevMap[e].CurrentFloor > floor && (elevMap[e].CurrentDir == config.ED_Down) {
			 				return false
			 			} else if  elevMap[e].CurrentFloor == floor{
			 				return false
			 			}
			 		} else if myDist == distElev {
			 			if elevMap[e].CurrentFloor < floor {
			 				if elevMap[config.My_ID].CurrentDir == elevMap[e].CurrentDir {
			 					if e > config.My_ID{
			 						return false
			 					}
			 				} else if elevMap[e].CurrentDir == config.ED_Up{
			 					return false

			 				}
			 			} else if elevMap[e].CurrentFloor > floor{
			 				if elevMap[config.My_ID].CurrentDir != elevMap[e].CurrentDir{
			 					if e > config.My_ID{
			 						return false
			 					}
			 				} else if elevMap[e].CurrentDir == config.ED_Down{
			 					return false

			 				}
			 			}
			 		} 
			 	}
			} 	
		 }
 	} else if elevMap[config.My_ID].CurrentFloor > floor {
	 	for e := 0; e<config.NUM_ELEVS; e++ {
	 		if elevMap[e].Connected == true{ //&& elevMap[e].OutOfOrder == false{	
			 	if e != config.My_ID{	
			 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
			 		if distElev < myDist{
			 			if elevMap[e].CurrentFloor >= floor && (elevMap[e].CurrentDir ==config.ED_Down){
			 				return false
			 			} else if elevMap[e].CurrentFloor <= floor && (elevMap[e].CurrentDir == config.ED_Up) {
		 					return false
		 				} else if  elevMap[e].CurrentFloor == floor{
		 					return false
			 			} 

			 		}else if myDist == distElev{
		 				if elevMap[e].CurrentFloor > floor {
		 					if elevMap[config.My_ID].CurrentDir == elevMap[e].CurrentDir {
		 						if e > config.My_ID{
		 							return false
		 						}
		 					} else if elevMap[e].CurrentDir == config.ED_Down{
		 						return false
		 					}
		 				} else if elevMap[e].CurrentFloor < floor{
		 					if elevMap[config.My_ID].CurrentDir != elevMap[e].CurrentDir{
		 						if e > config.My_ID{
		 							return false
		 						}
		 					} else if elevMap[e].CurrentDir == config.ED_Up{
		 						return false
		 					}
		 				}
		 			}
		 		}
		 	}
		}	
	} else if elevMap[config.My_ID].CurrentFloor == floor{
		if e > config.My_ID{
		 	return false
		}
	} 
 	return true
}


//Trenger ikke denne funksjonen om vi ikke skal ha motorTimeout
func forceChooseDirection(elevMap config.ElevStateMap) (config.MotorDirection, config.ElevDir){


	if orderInThisFloor(elevMap[config.My_ID].CurrentFloor, elevMap){
		return config.MD_Stop, elevMap[config.My_ID].CurrentDir
	
	} else if elevMap[config.My_ID].CurrentFloor >= config.NUM_FLOORS/2 {
		if ordersAbove(elevMap){
			elevMap[config.My_ID].CurrentDir = config.ED_Up
			return config.MD_Up, config.ED_Up
		} else if ordersBelow(elevMap){
			elevMap[config.My_ID].CurrentDir = config.ED_Down
			return config.MD_Down, config.ED_Down
		}
	} else if elevMap[config.My_ID].CurrentFloor < config.NUM_FLOORS/2{
		if ordersBelow(elevMap) {
			elevMap[config.My_ID].CurrentDir = config.ED_Down
			return config.MD_Down, config.ED_Down
		} else if ordersAbove(elevMap){
			elevMap[config.My_ID].CurrentDir = config.ED_Up
			return config.MD_Up, config.ED_Up
		}
	}
	return config.MD_Stop, elevMap[config.My_ID].CurrentDir

	//Det finnes ingen ordre!!
} 


//Kanskje vi må sende med map her, mulig det blir feil å bruke den i completedOrder om ikke???
func ClearOrder(elevMap config.ElevStateMap, button config.ButtonEvent, buttonLampChan chan config.ButtonLamp, activeOrderTx chan config.ActiveOrders) config.ElevStateMap{
	for elev := 0; elev < config.NUM_ELEVS; elev++{				
		elevMap[elev].Orders[button.Floor][button.Button] = config.OT_NoOrder
	}	
	buttonLampChan <-  config.ButtonLamp{button.Floor, button.Button, false}
	activeOrderTx <- config.ActiveOrders{button, config.My_ID, false, elevMap}
	fmt.Printf("clearer ordre ettter utført bestilling %v\n", button)
	return elevMap
}