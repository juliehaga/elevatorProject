package elevStateMap

import(
	"../config"
	"fmt"
	"sync"
	"os"
	"strconv"
	"io"
)


var LocalMap config.ElevStateMap
var mutex = &sync.Mutex{}


func FindActiveOrders(orderMsgChan chan config.NewPushes, activeOrderTx chan config.ActiveOrders, activeOrderRx chan config.ActiveOrders){
	for {
		select{
			case newButton := <- orderMsgChan:
				for f:= 0; f < config.NUM_FLOORS; f++{
					for b:= config.BT_HallUp; b < config.BT_Cab; b++{
						if newButton.PushesMatrix[f][b] == true {
							activeOrderTx <- config.ActiveOrders{config.ButtonEvent{f, b}, config.My_ID, true, newButton.ElevMap}
							activeOrderRx <- config.ActiveOrders{config.ButtonEvent{f, b}, config.My_ID, true, newButton.ElevMap}
						}
					}
				}	
		}
	}
}


func InitElevStateMap(buttonLampChan chan config.ButtonLamp){
	var _, err = os.Stat("backup.txt")

	if os.IsNotExist(err) {
		fmt.Printf("Creating new backup file\n")
		var file, err = os.Create("backup.txt")
		if err != nil{
			fmt.Printf("Could not create file \n")
		}	
		for f := 0; f < config.NUM_FLOORS; f++{
			LocalMap[config.My_ID].Orders[f][config.BT_Cab] = config.OT_NoOrder
		}
		defer file.Close()
	}else{
		readFromBackup(buttonLampChan)
	}
	
	for e:= 0; e < config.NUM_ELEVS; e++{
		LocalMap[e].CurrentFloor = -1
		LocalMap[e].CurrentDir = config.ED_Down
		LocalMap[e].Connected = false
		LocalMap[e].Door = false
		
		for f := 0; f < config.NUM_FLOORS; f++{
			for b :=0; b < config.BT_Cab; b++{
				LocalMap[e].Orders[f][b] = config.OT_NoOrder
			}
		}
		LocalMap[e].Orders[0][config.BT_HallDown] = -1
		LocalMap[e].Orders[config.NUM_FLOORS-1][config.BT_HallUp] = -1
	}
	LocalMap[config.My_ID].Connected = true
}


func UpdateLocalMap(changedMap config.ElevStateMap) (bool, config.ElevStateMap){
	currentMap := GetLocalMap()
	LocalOrderChangeMade := false

	currentMap[config.My_ID].CurrentFloor = changedMap[config.My_ID].CurrentFloor
	currentMap[config.My_ID].CurrentDir = changedMap[config.My_ID].CurrentDir
	currentMap[config.My_ID].Door = changedMap[config.My_ID].Door
	currentMap[config.My_ID].Connected= changedMap[config.My_ID].Connected

	for f:= 0; f < config.NUM_FLOORS; f++{
		currentMap[config.My_ID].Orders[f][config.BT_Cab] = changedMap[config.My_ID].Orders[f][config.BT_Cab]
		
		for b:= config.BT_HallUp; b < config.BT_Cab; b++{
			if changedMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced && currentMap[config.My_ID].Orders[f][b] == config.OT_NoOrder{
				LocalOrderChangeMade = true
				currentMap[config.My_ID].Orders[f][b] = changedMap[config.My_ID].Orders[f][b]

				for e := 0; e < config.NUM_ELEVS; e++{
					if changedMap[e].Connected == false{
						currentMap[e].Orders[f][b] = changedMap[config.My_ID].Orders[f][b]
					}
				} 
			} else if changedMap[config.My_ID].Orders[f][b] == config.OT_NoOrder && currentMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced {
				LocalOrderChangeMade = true
				currentMap[config.My_ID].Orders[f][b] = changedMap[config.My_ID].Orders[f][b]
				
				for elev := 0; elev < config.NUM_ELEVS; elev++{
					currentMap[elev].Orders[f][b] = config.OT_NoOrder
				}
			}
		}
	}
	SetLocalMap(currentMap)
	writeToBackup()
	return LocalOrderChangeMade, currentMap
}


func UpdateMapFromNetwork(receivedMap config.ElevStateMap, buttonLampChan chan config.ButtonLamp, activeOrderTx chan config.ActiveOrders, id int, orderMsgChan chan config.NewPushes) (bool, config.ElevStateMap){
	currentMap := GetLocalMap()
	updatedMap := GetLocalMap()
	changedMade := false
	var newButton config.ButtonEvent

	//Update status
	for e:= 0; e < config.NUM_ELEVS; e++{
		if e != config.My_ID {
			currentMap[e].CurrentFloor = receivedMap[e].CurrentFloor
			currentMap[e].CurrentDir = receivedMap[e].CurrentDir
			currentMap[e].Door = receivedMap[e].Door
		}
	}

	//Find number of connected elevators that has placed the order
	newPushes := [config.NUM_FLOORS][config.NUM_BUTTONS]bool{}
	ackOrdersMatrix := [config.NUM_FLOORS][config.NUM_BUTTONS]int{}
	connectedElevs := 0

	for e:= 0; e < config.NUM_ELEVS; e++{ 
		if receivedMap[e].Connected == true{
			connectedElevs++
			for f:= 0; f < config.NUM_FLOORS; f++{
				for b:= config.BT_HallUp; b < config.BT_Cab; b++{
					if receivedMap[e].Orders[f][b] == config.OT_OrderPlaced && receivedMap[e].Connected == true{
						ackOrdersMatrix[f][b]++
					}
				}
			}
		}	
	}

	//Find new placed orders in other elevators
	for e:= 0; e < config.NUM_ELEVS; e++{
		for f:= 0; f < config.NUM_FLOORS; f++{
			for b:= config.BT_HallUp; b < config.BT_Cab; b++{
				if (receivedMap[e].Orders[f][b] == config.OT_OrderPlaced && currentMap[e].Orders[f][b] == config.OT_NoOrder){ 
					if ackOrdersMatrix[f][b] < connectedElevs{
						newPushes[f][b] = true
						changedMade = true
						newButton = config.ButtonEvent{f, b}

						updatedMap[e].Orders[f][b] = config.OT_OrderPlaced
						updatedMap[config.My_ID].Orders[f][b]  = config.OT_OrderPlaced

						for elev:= 0; elev < config.NUM_ELEVS; elev++{ 
							if receivedMap[elev].Connected == false{
								updatedMap[elev].Orders[f][b]  = config.OT_OrderPlaced
							}
						}
					}else if connectedElevs < config.NUM_ELEVS {
						newPushes[f][b] = true
						newButton = config.ButtonEvent{f, b}

						updatedMap[e].Orders[f][b] = config.OT_OrderPlaced
						updatedMap[config.My_ID].Orders[f][b]  = config.OT_OrderPlaced

						for elev:= 0; elev < config.NUM_ELEVS; elev++{ 
							if receivedMap[elev].Connected == false{
								updatedMap[elev].Orders[f][b]  = config.OT_OrderPlaced
							}
						}
						orderMsgChan <- config.NewPushes{newPushes, updatedMap}
					}
				}
			}
		}
	}

	//If a new order is found, check if the order is placed in all elevators before acknowledging
	numberOfAckElevs := [config.NUM_FLOORS][config.NUM_BUTTONS]int{}
	if changedMade == true {
		for e:= 0; e < config.NUM_ELEVS; e++{
			if updatedMap[e].Orders[newButton.Floor][newButton.Button] == config.OT_OrderPlaced && receivedMap[e].Connected == true{
				numberOfAckElevs[newButton.Floor][newButton.Button] ++
			}
		}
		if numberOfAckElevs[newButton.Floor][newButton.Button] == connectedElevs{
			orderMsgChan <- config.NewPushes{newPushes, updatedMap}
		}
	}

	//Find executed orders by other elevators
	for f:= 0; f < config.NUM_FLOORS; f++{
		for b:= config.BT_HallUp; b < config.BT_Cab; b++{
			if receivedMap[id].Orders[f][b] == config.OT_NoOrder && currentMap[id].Orders[f][b] == config.OT_OrderPlaced{
				for e := 0; e < config.NUM_ELEVS; e++{
					if (receivedMap[e].CurrentFloor == f && receivedMap[e].Door == true) && receivedMap[e].Connected == true{							
						buttonLampChan <- config.ButtonLamp{f, b, false}
						changedMade = false 
						for elev := 0; elev < config.NUM_ELEVS; elev++{
							updatedMap[elev].Orders[f][b] = config.OT_NoOrder
						}
						activeOrderTx <- config.ActiveOrders{config.ButtonEvent{f, b}, config.My_ID, false, updatedMap}
					}
				}
			}
		}
	}
	SetLocalMap(updatedMap)
	return changedMade, updatedMap
}


func UpdateElevStatusFromNetwork(newStatus config.StatusMsg){
	currentMap := GetLocalMap()
	currentMap[newStatus.ID].CurrentFloor = newStatus.CurrentFloor
	currentMap[newStatus.ID].CurrentDir = newStatus.CurrentDir
	currentMap[newStatus.ID].Door = newStatus.Door
	currentMap[newStatus.ID].Connected = newStatus.Connected
	SetLocalMap(currentMap)
}


func SetConnectedElevator(ID int, value bool){
	currentMap := GetLocalMap()
	currentMap[ID].Connected = value
	SetLocalMap(currentMap)
}


func GetLocalMap() config.ElevStateMap{
	mutex.Lock()
	currentMap := LocalMap
	mutex.Unlock()
	return currentMap
}


func SetLocalMap(changedMap config.ElevStateMap){
	mutex.Lock()
	LocalMap = changedMap
	mutex.Unlock()
}


func writeToBackup(){
	var file, err = os.OpenFile("backup.txt", os.O_RDWR, 0644)
	if err != nil { return }
	defer file.Close()

	for f := 0; f< config.NUM_FLOORS; f++{
		var order int = int(LocalMap[config.My_ID].Orders[f][config.BT_Cab])
		_, err = file.WriteString(strconv.Itoa(order))
		if err != nil { return }
	}
	err = file.Sync()
	if err != nil { return }
}


func readFromBackup(buttonLampChan chan config.ButtonLamp){
	var file, err = os.OpenFile("backup.txt", os.O_RDWR, 0644)
	if err != nil { return }
	defer file.Close()
	currentMap := GetLocalMap()
	var buf = make([]byte, 1024)
	for {
		_, err = file.Read(buf)
		
		if err == io.EOF {
			break
		}
		
		if err != nil && err != io.EOF {
			break
		}
		fmt.Printf("Reading cab orders from backup \n")
		fmt.Printf("floor 1: %v\n", string(buf[0]))
		fmt.Printf("floor 2: %v\n", string(buf[1]))
		fmt.Printf("floor 3: %v\n", string(buf[2]))
		fmt.Printf("floor 4: %v\n", string(buf[3]))
	}
	for floor := 0; floor<config.NUM_FLOORS; floor++{
		order, _ :=strconv.Atoi(string(buf[floor]))
		currentMap[config.My_ID].Orders[floor][config.BT_Cab] = config.OrderType(order)
		if currentMap[config.My_ID].Orders[floor][config.BT_Cab] == config.OT_OrderPlaced{
			buttonLampChan <- config.ButtonLamp{floor, config.BT_Cab, true}
		}
	}
	SetLocalMap(currentMap)
}


func PrintMap(elevMap config.ElevStateMap){
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