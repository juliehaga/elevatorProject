package network

import (
	"../elevStateMap"
	"../config"
	"./bcast"
)

const EMPTY = -1

var Sent [config.NUM_ELEVS]int64
var Received [config.NUM_ELEVS]int64


func Network(messageRx chan config.Message, messageTx chan config.Message, statusMsgRx chan config.StatusMsg, orderMsgRx chan config.OrderMsg, activeOrderRx chan config.ActiveOrders){

	go bcast.Transmitter(16666, messageTx)
	go bcast.Receiver(16666, messageRx)

	for {
		select{
			case receivedMsg := <-messageRx:
				if receivedMsg.ID != config.My_ID{
					if receivedMsg.Stamp == Received[receivedMsg.ID]{

						if receivedMsg.MsgType == config.ElevStatus{
							statusMsgRx <- config.StatusMsg{receivedMsg.ID, receivedMsg.ElevMap[receivedMsg.ID].CurrentFloor, receivedMsg.ElevMap[receivedMsg.ID].CurrentDir, receivedMsg.ElevMap[receivedMsg.ID].Door, receivedMsg.ElevMap[receivedMsg.ID].Connected}
						} else if receivedMsg.MsgType == config.Orders {
							orderMsgRx <- config.OrderMsg{receivedMsg.ID, receivedMsg.ElevMap}
						} else if receivedMsg.MsgType == config.ActiveOrder{
							activeOrderRx <- config.ActiveOrders{receivedMsg.Button, receivedMsg.ID, true, receivedMsg.ElevMap}
						} 
						Received[receivedMsg.ID]++
					}
				}
		}
	}
}


func SendOrders(messageTx chan config.Message, elevMap config.ElevStateMap) {
	elevMapMsg := config.Message{EMPTY,config.My_ID, config.Orders, elevMap, EMPTY, config.ButtonEvent{0, config.BT_HallDown}}
	sendNetworkMsg(elevMapMsg, messageTx)
}

func SendElevStatus(messageTx chan config.Message,  elevMap config.ElevStateMap){
	elevMapMsg := config.Message{EMPTY, config.My_ID, config.ElevStatus, elevMap, EMPTY, config.ButtonEvent{0, config.BT_HallDown}}
	sendNetworkMsg(elevMapMsg, messageTx)
}

func SendActiveOrder(messageTx chan config.Message,  order config.ActiveOrders){
	elevMap := elevStateMap.GetLocalMap()
	elevMapMsg := config.Message{EMPTY, config.My_ID, config.ActiveOrder, elevMap, EMPTY, order.Button}
	sendNetworkMsg(elevMapMsg, messageTx)
}

func sendNetworkMsg(message config.Message, messageTx chan config.Message){
	for e:= 0; e < config.NUM_ELEVS; e++{
		if e != config.My_ID{
			message.Receiver_ID = e	
			message.Stamp = Sent[e]	

			if message.ElevMap[e].Connected == true{	
				for i := 0; i < 5; i++{
					messageTx <- message	
				}
				Sent[e]++
			}
		}
	}
}
































