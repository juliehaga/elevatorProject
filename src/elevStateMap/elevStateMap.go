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




func writeToBackup(){
	// open file using READ & WRITE permission
	var file, err = os.OpenFile("backup.txt", os.O_RDWR, 0644)
	if err != nil { return }
	defer file.Close()

	// write some text line-by-line to file

	for f := 0; f< config.NUM_FLOORS; f++{
		var order int = int(LocalMap[config.My_ID].Orders[f][config.BT_Cab])
		_, err = file.WriteString(strconv.Itoa(order))
		if err != nil { return }
	}
	// save changes
	err = file.Sync()
	if err != nil { return }
}

func readFromBackup(buttonLampChan chan config.ButtonLamp){
	// re-open file
	var file, err = os.OpenFile("backup.txt", os.O_RDWR, 0644)
	if err != nil { return }
	defer file.Close()
	currentMap := GetLocalMap()

	// read file, line by line

	var buf = make([]byte, 1024)
	for {
		_, err = file.Read(buf)
		
		// break if finally arrived at end of file
		if err == io.EOF {
			break
		}
		
		// break if error occured
		if err != nil && err != io.EOF {
			break
		}
		//order := strconv.Atoi((string(buf)))

		fmt.Printf("order 1: %v\n", string(buf[0]))
		fmt.Printf("order 1: %v\n", string(buf[1]))
		fmt.Printf("order 1: %v\n", string(buf[2]))
		fmt.Printf("order 1: %v\n", string(buf[3]))


		//fmt.Printf("New line %v \n", order)
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

func InitElevStateMap(buttonLampChan chan config.ButtonLamp){
	var _, err = os.Stat("backup.txt")

	if os.IsNotExist(err) {
		fmt.Printf("Creating new backup file\n")
		var file, err = os.Create("backup.txt")
		if err != nil{
			fmt.Printf("Could not create file \n")
			//return
		}	
		for f := 0; f < config.NUM_FLOORS; f++{
			LocalMap[config.My_ID].Orders[f][config.BT_Cab] = config.OT_NoOrder
		}
		defer file.Close()
	}else{
		readFromBackup(buttonLampChan)

	}


	//initialize map	
	for e:= 0; e < config.NUM_ELEVS; e++{
		LocalMap[e].IDLE = true
		LocalMap[e].CurrentDir = config.ED_Down
		LocalMap[e].Connected = false
		LocalMap[e].Door = false
		LocalMap[e].CurrentFloor = -1
		LocalMap[e].OutOfOrder = false
			
		for f := 0; f < config.NUM_FLOORS; f++{
			for b :=0; b < config.BT_Cab; b++{
				LocalMap[e].Orders[f][b] = config.OT_NoOrder
			
			}
		}

		LocalMap[e].Orders[0][config.BT_HallDown] = -1
		LocalMap[e].Orders[config.NUM_FLOORS-1][config.BT_HallUp] = -1
	}
	//LocalMap[config.My_ID].CurrentFloor = elevio.GetFloor()
	LocalMap[config.My_ID].Connected = true
	//PrintMap(LocalMap)
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


func UpdateLocalMap(changedMap config.ElevStateMap) (bool, config.ElevStateMap){
	currentMap := GetLocalMap()
	LocalOrderChangeMade := false


//Hvilke av disse trenger vi faktisk???
	currentMap[config.My_ID].CurrentFloor = changedMap[config.My_ID].CurrentFloor
	currentMap[config.My_ID].CurrentDir = changedMap[config.My_ID].CurrentDir
	currentMap[config.My_ID].Door = changedMap[config.My_ID].Door
	currentMap[config.My_ID].IDLE = changedMap[config.My_ID].IDLE
	currentMap[config.My_ID].OutOfOrder = changedMap[config.My_ID].OutOfOrder


	for f:= 0; f < config.NUM_FLOORS; f++{
		//CAB-orders kan skrives rett over fordi de sendes ikke
		currentMap[config.My_ID].Orders[f][config.BT_Cab] = changedMap[config.My_ID].Orders[f][config.BT_Cab]

		for b:= config.BT_HallUp; b < config.BT_Cab; b++{
			if changedMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced && currentMap[config.My_ID].Orders[f][b] == config.OT_NoOrder{
				//lagt inn en ordre, dersom local -> send
				LocalOrderChangeMade = true
				fmt.Printf("jeg legger til en lokal ordre\n")
				currentMap[config.My_ID].Orders[f][b] = changedMap[config.My_ID].Orders[f][b]

				for e := 0; e < config.NUM_ELEVS; e++{
					if changedMap[e].Connected == false{
						currentMap[e].Orders[f][b] = changedMap[config.My_ID].Orders[f][b]
					}
				} 

			} else if changedMap[config.My_ID].Orders[f][b] == config.OT_NoOrder && currentMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced {
				LocalOrderChangeMade = true
				currentMap[config.My_ID].Orders[f][b] = changedMap[config.My_ID].Orders[f][b]
				fmt.Printf("SENDING CLEAR ORDER\n")
				for elev := 0; elev < config.NUM_ELEVS; elev++{
					currentMap[elev].Orders[f][b] = config.OT_NoOrder
				}
			}
			//sjekk floor og door_open
			//send en ordre_complete_msg
		}
	}
	SetLocalMap(currentMap)
	writeToBackup()
	return LocalOrderChangeMade, currentMap
}


func UpdateMapFromNetwork(recievedMap config.ElevStateMap, buttonLampChan chan config.ButtonLamp, activeOrderTx chan config.ActiveOrders, id int, orderMsgChan chan config.NewButtons) (bool, config.ElevStateMap){

	currentMap := GetLocalMap()
	changedMade := false

	newMap := GetLocalMap()

	

	//Update status
	for e:= 0; e < config.NUM_ELEVS; e++{
		if e != config.My_ID {
		currentMap[e].CurrentFloor = recievedMap[e].CurrentFloor
		currentMap[e].CurrentDir = recievedMap[e].CurrentDir
		currentMap[e].Door = recievedMap[e].Door
		currentMap[e].OutOfOrder = recievedMap[e].OutOfOrder
		currentMap[e].IDLE = recievedMap[e].IDLE
		}
	}



	newButtons := [config.NUM_FLOORS][config.NUM_BUTTONS]bool{}
	orders := [config.NUM_FLOORS][config.NUM_BUTTONS]int{}
	connectedElevs := 0


	//ser hvor mange som har lagt inn bestillingne i mapet vi mottar
	for e:= 0; e < config.NUM_ELEVS; e++{ 
		if recievedMap[e].Connected == true{
			connectedElevs++

			for f:= 0; f < config.NUM_FLOORS; f++{
				for b:= config.BT_HallUp; b < config.BT_Cab; b++{
					if recievedMap[e].Orders[f][b] == config.OT_OrderPlaced && recievedMap[e].Connected == true{
						orders[f][b]++
					}
				}
			}
		}	
	}

	var newButton config.ButtonEvent


	for e:= 0; e < config.NUM_ELEVS; e++{
		for f:= 0; f < config.NUM_FLOORS; f++{
			for b:= config.BT_HallUp; b < config.BT_Cab; b++{

				if (recievedMap[e].Orders[f][b] == config.OT_OrderPlaced && currentMap[e].Orders[f][b] == config.OT_NoOrder){ //&& recievedMap[e].Connected == true{
					fmt.Printf("orders[f][b] %v, connectedElevs %v \n", orders[f][b], connectedElevs)

					if orders[f][b] < connectedElevs{
						//har funnet en ny ordre
						newButtons[f][b] = true
						changedMade = true
						fmt.Printf("Fant en 1-er")
						newButton = config.ButtonEvent{f, b}

						//kopierer ordren og legger den til i min ordrematrise
						newMap[e].Orders[f][b] = config.OT_OrderPlaced
						newMap[config.My_ID].Orders[f][b]  = config.OT_OrderPlaced


						//Vi har funnet en ordre. Sjekker om noen heiser er disconnected. Da må vi legge inn bestillingen der. 
						for elev:= 0; elev < config.NUM_ELEVS; elev++{ 
							if recievedMap[elev].Connected == false{
								newMap[elev].Orders[f][b]  = config.OT_OrderPlaced
							}
						}
						//fmt.Printf("fant nytt knappe trykk %v %v, lagt til i egen matrise\n", f, b)
					}else if connectedElevs < config.NUM_ELEVS {
						//vi har en disconnected heis
						newButtons[f][b] = true
						newButton = config.ButtonEvent{f, b}

						//kopierer ordren og legger den til i min ordrematrise
						newMap[e].Orders[f][b] = config.OT_OrderPlaced
						newMap[config.My_ID].Orders[f][b]  = config.OT_OrderPlaced

						//Vi har funnet en ordre. Sjekker om noen heiser er disconnected. Da må vi legge inn bestillingen der. 
						for elev:= 0; elev < config.NUM_ELEVS; elev++{ 
							if recievedMap[elev].Connected == false{
								newMap[elev].Orders[f][b]  = config.OT_OrderPlaced
							}
						}
						orderMsgChan <- config.NewButtons{newButtons, newMap}



					}

				}

			}
		}
	}

	numberOfAckElevs := [config.NUM_FLOORS][config.NUM_BUTTONS]int{}

	//leiter etter 3 -enere i det oppdaterte mapet
	if changedMade == true {
		for e:= 0; e < config.NUM_ELEVS; e++{
			if newMap[e].Orders[newButton.Floor][newButton.Button] == config.OT_OrderPlaced && recievedMap[e].Connected == true{
				fmt.Printf("inkrementerer ACK elevs på %v\n", e)
				numberOfAckElevs[newButton.Floor][newButton.Button] ++
			}
		}

		if numberOfAckElevs[newButton.Floor][newButton.Button] == connectedElevs{
			fmt.Printf("aksepterer ordren\n")
			//dersom vi finner 3 1-ere uten at det er funnet en ny 1-er vil newButtons være 00000
			orderMsgChan <- config.NewButtons{newButtons, newMap}
			//dersom du har lagt inn 3 enere - ønsker du ikke å sende map-etss
		}
	}



	

	for f:= 0; f < config.NUM_FLOORS; f++{
		for b:= config.BT_HallUp; b < config.BT_Cab; b++{

			if recievedMap[id].Orders[f][b] ==config.OT_NoOrder && currentMap[id].Orders[f][b] == config.OT_OrderPlaced{
				for e := 0; e < config.NUM_ELEVS; e++{
					if (recievedMap[e].CurrentFloor == f && recievedMap[e].Door == true) && recievedMap[e].Connected == true{							
						buttonLampChan <- config.ButtonLamp{f, b, false}
						fmt.Printf("fjerner ordre fra nettverket %v %v \n", f, b)
						changedMade = false
						//clear orders from network 
						
						for elev := 0; elev < config.NUM_ELEVS; elev++{
							newMap[elev].Orders[f][b] = config.OT_NoOrder
						}
						activeOrderTx <- config.ActiveOrders{config.ButtonEvent{f, b}, config.My_ID, false, newMap}
					}
				}
			}
		}
	}

	SetLocalMap(newMap)
	return changedMade, newMap
}


func FindActiveOrders(orderMsgChan chan config.NewButtons, activeOrderTx chan config.ActiveOrders, activeOrderRx chan config.ActiveOrders){

	for {
		select{
		//Burde bare gjøre sjekken når man faktisk mottar en ordre. 
		case newButton := <- orderMsgChan:

			for f:= 0; f < config.NUM_FLOORS; f++{
				for b:= config.BT_HallUp; b < config.BT_Cab; b++{
					if newButton.PushesMatrix[f][b] == true {
						fmt.Printf("jeg har 3 enere, sender ut ordre melding %v \n", config.ButtonEvent{f, b})
						activeOrderTx <- config.ActiveOrders{config.ButtonEvent{f, b}, config.My_ID, true, newButton.ElevMap}
						activeOrderRx <- config.ActiveOrders{config.ButtonEvent{f, b}, config.My_ID, true, newButton.ElevMap}
					}
				}
			
			}	
		//time.Sleep(10* time.Millisecond)
		}

	}
}


















func UpdateElevStatusFromNetwork(newStatus config.StatusMsg){
	currentMap := GetLocalMap()
	currentMap[newStatus.ID].CurrentFloor = newStatus.CurrentFloor
	currentMap[newStatus.ID].CurrentDir = newStatus.CurrentDir
	currentMap[newStatus.ID].Door = newStatus.Door
	currentMap[newStatus.ID].OutOfOrder = newStatus.OutOfOrder
	currentMap[newStatus.ID].IDLE = newStatus.IDLE
	SetLocalMap(currentMap)
}

func SetConnectedElevator(ID int, value bool){
	currentMap := GetLocalMap()
	currentMap[ID].Connected = value
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