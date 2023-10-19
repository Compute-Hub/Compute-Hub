package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pion/webrtc/v3"
)

func SetupProvider() {
	var answerReceived = make(chan string)

	var config webrtc.Configuration = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			fmt.Println("Peer Connection has gone to failed exiting")
		}

		if s == webrtc.PeerConnectionStateClosed {
			fmt.Println("Peer Connection has gone to closed exiting")
			os.Exit(0)
		}
	})

	dataChannel, err := peerConnection.CreateDataChannel("myDataChannel", nil)
	if err != nil {
		fmt.Println("Failed to create Data Channel")
	}

	dataChannel.OnOpen(func() {
		fmt.Println("Data Channel Created and ready to use")
	})

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		if msg.IsString {
			fmt.Printf("Received message: %s\n", string(msg.Data))
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		fmt.Println("Error Setting Local description")
		panic(err)
	}

	go func() {
		answer := sendOffer(offer)
		answerReceived <- answer
	}()

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	remoteAnswer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  <-answerReceived,
	}

	if err := peerConnection.SetRemoteDescription(remoteAnswer); err != nil {
		fmt.Println("Error setting remote description:", err)
		return
	}

	<-gatherComplete

	//fmt.Println("Connection Eastablished!")

	if dataChannel.ReadyState() == webrtc.DataChannelStateOpen {
		message := "Hemlo from Provider!"
		dataChannel.SendText(message)
		fmt.Printf("Sent message: %s\n", message)
	}

	select {}

}

type Offer struct {
	ID  int16
	SDP string
}

func sendOffer(offer webrtc.SessionDescription) string {
	url := "https://compute-hub-server.onrender.com/offer"

	of := Offer{
		ID:  1,
		SDP: offer.SDP,
	}

	offerbytes, err := json.Marshal(of)
	if err != nil {
		fmt.Println("Error Marshaling Json!", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(offerbytes))
	if err != nil {
		fmt.Println("Error creating the request:", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending the request:", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Request was successful")
		var response struct {
			Reciever string `json:"receiver"`
		}

		// Parse the response JSON
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			fmt.Println("Error decoding JSON response:", err)
		}

		// Access the "answer" key
		if response.Reciever != "" {
			// The "answer" key exists, and you can access it with response.Answer
			//fmt.Println("Received answer:", response.Reciever)
			return response.Reciever
		} else {
			fmt.Println("No 'answer' key found in the response.")
		}
	} else {
		fmt.Printf("Request failed with status code: %d\n", resp.StatusCode)
	}
	return ""
}
