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

var Sent [config.NUM_ELEVS]int64
var Received [config.NUM_ELEVS]int64





func Network(messageRx chan config.Message, messageTx chan config.Message, statusMsgRx chan config.StatusMsg, orderMsgRx chan config.OrderMsg, activeOrderRx chan config.ActiveOrders,  ackChan chan config.AckMsg){

	go bcast.Transmitter(16666, messageTx)
	go bcast.Receiver(16666, messageRx)

	for {
		select{
			case receivedMsg := <-messageRx:

				if receivedMsg.ID != config.My_ID{
					if receivedMsg.Stamp == Received[receivedMsg.ID]{
						//har ikke lest denne meldinger


						if receivedMsg.MsgType == config.ElevStatus{
							statusMsgRx <- config.StatusMsg{receivedMsg.ID, receivedMsg.ElevMap[receivedMsg.ID].CurrentFloor, receivedMsg.ElevMap[receivedMsg.ID].CurrentDir, receivedMsg.ElevMap[receivedMsg.ID].Door, receivedMsg.ElevMap[receivedMsg.ID].Connected}
						} else if receivedMsg.MsgType == config.Orders {
							orderMsgRx <- config.OrderMsg{receivedMsg.ID, receivedMsg.ElevMap}
						} else if receivedMsg.MsgType == config.ActiveOrder{
							fmt.Printf("Mottar en ordremsg fra %v , om knapp: %v \n", receivedMsg.ID, receivedMsg.Button)
							activeOrderRx <- config.ActiveOrders{receivedMsg.Button, receivedMsg.ID, true, receivedMsg.ElevMap}
						} 
						Received[receivedMsg.ID]++

					}
				}
		}
	}
	
}


func SendOrders(messageTx chan config.Message, elevMap config.ElevStateMap, ackChan chan config.AckMsg) {
	elevMapMsg := config.Message{-1,config.My_ID, config.Orders, elevMap, -1, config.ButtonEvent{0, config.BT_HallDown}}
	waitForAck(elevMapMsg, messageTx, ackChan)
}

func SendElevStatus(messageTx chan config.Message,  elevMap config.ElevStateMap, ackChan chan config.AckMsg){
	elevMapMsg := config.Message{-1, config.My_ID, config.ElevStatus, elevMap, -1, config.ButtonEvent{0, config.BT_HallDown}}
	waitForAck(elevMapMsg, messageTx, ackChan)
}


func SendActiveOrder(messageTx chan config.Message,  order config.ActiveOrders, ackChan chan config.AckMsg){
	elevMap := elevStateMap.GetLocalMap()
	elevMapMsg := config.Message{-1, config.My_ID, config.ActiveOrder, elevMap, -1, order.Button}
	waitForAck(elevMapMsg, messageTx,ackChan)
}

func waitForAck(message config.Message, messageTx chan config.Message, ackChan chan config.AckMsg){
	for e:= 0; e < config.NUM_ELEVS; e++{
		if e != config.My_ID{
			message.Reciever_ID = e	
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
































