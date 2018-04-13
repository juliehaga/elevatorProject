package network

import (
	"../elevStateMap"
	"../config"
	"./bcast"
	"fmt"
	"time"
)


const interval = 15 * time.Millisecond
const timeout = 1500 * time.Millisecond





func Network(messageRx chan config.Message, messageTx chan config.Message, statusMsgRx chan config.StatusMsg, orderMsgRx chan config.OrderMsg, activeOrderRx chan config.ActiveOrders){

	go bcast.Transmitter(16666, messageTx)
	go bcast.Receiver(16666, messageRx)

	for {
		select{
			case receivedMsg := <-messageRx:
				if receivedMsg.ID != config.My_ID{

					if receivedMsg.MsgType == config.ElevStatus{
						statusMsgRx <- config.StatusMsg{receivedMsg.ID, receivedMsg.ElevMap[receivedMsg.ID].CurrentFloor, receivedMsg.ElevMap[receivedMsg.ID].CurrentDir, receivedMsg.ElevMap[receivedMsg.ID].Door, receivedMsg.ElevMap[receivedMsg.ID].Connected}

					} else if receivedMsg.MsgType == config.Orders {
						orderMsgRx <- config.OrderMsg{receivedMsg.ID, receivedMsg.ElevMap}

					} else if receivedMsg.MsgType == config.ActiveOrder{
						fmt.Printf("Mottar en ordremsg fra %v , om knapp: %v \n", receivedMsg.ID, receivedMsg.Button)

						activeOrderRx <- config.ActiveOrders{receivedMsg.Button, receivedMsg.ID, true, receivedMsg.ElevMap}
					}
				}

		}
	}
	
}


func SendOrders(messageTx chan config.Message, elevMap config.ElevStateMap) {
	elevMapMsg := config.Message{config.My_ID, config.Orders, elevMap, -1, config.ButtonEvent{0, config.BT_HallDown}}
	messageTx <- elevMapMsg
}

func SendElevStatus(messageTx chan config.Message,  elevMap config.ElevStateMap){
	elevMapMsg := config.Message{config.My_ID, config.ElevStatus, elevMap, -1, config.ButtonEvent{0, config.BT_HallDown}}
	messageTx <- elevMapMsg
}


func SendActiveOrder(messageTx chan config.Message,  order config.ActiveOrders){
	elevMap := elevStateMap.GetLocalMap()
	elevMapMsg := config.Message{config.My_ID, config.ActiveOrder, elevMap, -1, order.Button}
	messageTx <- elevMapMsg
}
































