package config

import "strconv"

const(
	NUM_ELEVS		= 3
	NUM_FLOORS   	= 4
	NUM_BUTTONS		= 3	
)

var My_ID int
var My_PORT int 

func InitGlobalSettings(id string, port string){
	My_ID, _ = strconv.Atoi(id)
	My_PORT, _ = strconv.Atoi(port)
}


type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type ButtonLamp struct {
	Floor  int
	Button ButtonType
	Set bool
}

type ActiveOrders struct {
	Button ButtonEvent
	ID int
	ActiveOrder bool
	ElevMap ElevStateMap
}

type NewPushes struct {
	PushesMatrix [NUM_FLOORS][NUM_BUTTONS]bool
	ElevMap ElevStateMap
}

type ElevInfo struct{
	CurrentFloor int
	CurrentDir ElevDir
	Connected bool
	Door bool
	Orders [NUM_FLOORS][NUM_BUTTONS] OrderType
}


type ElevStateMap [NUM_ELEVS]ElevInfo

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

type Message struct {
	Stamp int64
	ID int
	MsgType int
	ElevMap   ElevStateMap
	Receiver_ID int 
	Button ButtonEvent
}


type StatusMsg struct {
	ID int
	CurrentFloor int
	CurrentDir ElevDir
	Door bool
	Connected bool
}

type OrderMsg struct{
	ID int
	ElevMap ElevStateMap
}

type ActiveOrderMsg struct{
	ID int
	Order ButtonEvent
}

type OrderCompleteMsg struct{
	button ButtonType
}


type ButtonType int

type MotorDirection int

type OrderType int

type ElevDir int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)


const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

const (
	OT_NoOrder             = 0
	OT_OrderPlaced         = 1
)

const (
	ED_Up   	   ElevDir = 1
	ED_Down                = -1
)

const (
	ElevStatus             = 0
	Orders                 = 1
	Ack 	               = 2
	ActiveOrder    		   = 3
)



