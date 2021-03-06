package elevio


import( 
	"../config"
	"../elevStateMap"
	"time"
	"sync"
	"net"
	"fmt"
)



const _pollRate = 100 * time.Millisecond

var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex
var _conn net.Conn


func Elevio(motorChan chan config.MotorDirection, doorLampChan chan bool, newOrderChan chan config.ButtonEvent, floorChan chan int, buttonLampChan chan config.ButtonLamp, newLocalOrderChan chan config.ButtonEvent, mapChangesChan chan config.ElevStateMap) {
	go PollButtons(newLocalOrderChan)
    go PollFloorSensor(floorChan)

	for {
		select {
		case dir := <- motorChan:
			SetMotorDirection(dir)

		case light := <-doorLampChan:
			SetDoorOpenLamp(light)

		case lamp := <- buttonLampChan:
			SetButtonLamp(lamp)

		case orderButton := <- newLocalOrderChan: 
			acceptOrder := false
			currentMap := elevStateMap.GetLocalMap()
			if orderButton.Button == config.BT_Cab && currentMap[config.My_ID].Connected == true{
				newOrderChan <- orderButton
			} else if orderButton.Button != config.BT_Cab{
				for e:= 0; e < config.NUM_ELEVS; e++{
					if (currentMap[e].Connected && e != config.My_ID) && currentMap[config.My_ID].Connected == true{
						acceptOrder = true
					}
				}
				if acceptOrder == true{
					currentMap[config.My_ID].Orders[orderButton.Floor][orderButton.Button] = config.OT_OrderPlaced
				} else {
					fmt.Printf("No way to assure redundancy\n")
				} 
				mapChangesChan <- currentMap
			}
		}	
	}
}


func InitDriver(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
	SetDoorOpenLamp(false)
	currentMap := elevStateMap.GetLocalMap()
	SetMotorDirection(config.MD_Down)
	currentMap[config.My_ID].CurrentDir = config.ED_Down

	for getFloor() == -1{}
	SetMotorDirection(config.MD_Stop)

	currentMap[config.My_ID].CurrentFloor = getFloor()
	elevStateMap.SetLocalMap(currentMap)
	clearAllButtonLamps()
}


func InitOrdersFromNetwork(networkMap config.ElevStateMap){
	numberOfAckElevs := 0
	currentMap := elevStateMap.GetLocalMap()

	for f := 0; f < config.NUM_FLOORS; f++{
		for b:= config.BT_HallUp; b < config.BT_Cab; b++{
			for e := 0; e < config.NUM_ELEVS; e++{
				currentMap[e].Orders[f][b] = networkMap[e].Orders[f][b]
				if currentMap[e].Orders[f][b] == config.OT_OrderPlaced{
					numberOfAckElevs  ++;
				}
			}
			if numberOfAckElevs  == config.NUM_ELEVS{
				SetButtonLamp(config.ButtonLamp{f, b, true})
				numberOfAckElevs = 0
			}else{
				numberOfAckElevs = 0
				SetButtonLamp(config.ButtonLamp{f, b, false})
			}
		}
	}
	elevStateMap.SetLocalMap(currentMap)
}


func clearAllButtonLamps(){
	for f:= 0; f < _numFloors; f++ {
		for b:= config.ButtonType(0); b < 3; b++ {
			SetButtonLamp(config.ButtonLamp{f, b, false})	
		}
	}
}


func SetMotorDirection(dir config.MotorDirection) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{1, byte(dir), 0, 0})
}


func SetButtonLamp(lamp config.ButtonLamp) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{2, byte(lamp.Button), byte(lamp.Floor), toByte(lamp.Set)})
}


func SetFloorIndicator(floor int) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{3, byte(floor), 0, 0})
}


func SetDoorOpenLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{4, toByte(value), 0, 0})
}


func SetStopLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{5, toByte(value), 0, 0})
}


func PollButtons(receiver chan<- config.ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := config.ButtonType(0); b < 3; b++ {
				v := getButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- config.ButtonEvent{f, config.ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}


func PollFloorSensor(receiver chan<- int) {	
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := getFloor()
		if v < _numFloors && v >= 0 {
			SetFloorIndicator(v)
		}
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}


func getButton(button config.ButtonType, floor int) bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{6, byte(button), byte(floor), 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}


func getFloor() int {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}


func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}


func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
