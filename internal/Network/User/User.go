package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pion/webrtc/v3"
)

func SetupReceiver() {
	var connected = false
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
			// Handle the failure appropriately
			fmt.Println("Peer Connection has gone to failed")
		}

		if s == webrtc.PeerConnectionStateClosed {
			// Handle the closure appropriately
			fmt.Println("Peer Connection has gone to closed")
		}

		if s == webrtc.PeerConnectionStateConnected {
			connected = true
			fmt.Println("Connection Established")
		}
	})

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
		})
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	receivedOffer, err := getOffer()
	if err != nil {
		fmt.Println("Error in recieved offer")
	}

	fmt.Println("Recieved Offer", receivedOffer)
	// Set the remote description with the received offer
	if err = peerConnection.SetRemoteDescription(*receivedOffer); err != nil {
		fmt.Println("Error setting remote description:", err)
		return
	}

	dataChannel, err := peerConnection.CreateDataChannel("myDataChannel", nil)
	if err != nil {
		fmt.Println("Failed to create Data Channel")
	}

	// Set up a handler to receive messages on the data channel
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("Message from Provider '%s': '%s'\n", dataChannel.Label(), string(msg.Data))
	})

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		fmt.Println("Error creating answer:", err)
		return
	}

	// Set the local description with the answer
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		fmt.Println("Error setting local description:", err)
		return
	}
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Send the answer back to the signaling server
	sendAnswer(answer)

	<-gatherComplete

	if dataChannel.ReadyState() == webrtc.DataChannelStateOpen && connected {
		message := "Hemlo from User!"
		dataChannel.SendText(message)
		fmt.Printf("Sent message: %s\n", message)
	}

	select {}
}

func sendAnswer(answer webrtc.SessionDescription) {
	url := "https://compute-hub-server.onrender.com/sendanswer"
	requestData := struct {
		ConnectID int    `json:"connectID"`
		Answer    string `json:"answer"`
	}{
		ConnectID: 1, // Replace with the appropriate ConnectID
		Answer:    answer.SDP,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Println("Error marshaling JSON data:", err)
		return
	}

	// Create a new HTTP request with a POST method
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client
	client := &http.Client{}

	// Send the HTTP POST request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP POST request:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		//fmt.Println("Answer sent successfully")
	} else {
		fmt.Printf("HTTP POST request failed with status code: %d\n", resp.StatusCode)
	}
}

func getOffer() (*webrtc.SessionDescription, error) {
	url := "https://compute-hub-server.onrender.com/getoffer"

	jsonStr := `{"connectID": 1}`

	// Create a request with the JSON object as the request body
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		fmt.Println("Error creating the request:", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending the request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Request was successful")

		var response struct {
			Offer struct {
				ID      interface{} `json:"ID"`
				SDP     string      `json:"SDP"`
				Connect bool        `json:"connect"`
			} `json:"offer"`
			Success bool `json:"success"`
		}

		// Parse the response JSON
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			fmt.Println("Error decoding JSON response:", err)
			return nil, err
		}

		if response.Success {
			// Create a webrtc.SessionDescription from the SDP string in the response
			sessionDescription := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  response.Offer.SDP,
			}
			//fmt.Print("My formatted Session Description:", sessionDescription)
			if err != nil {
				fmt.Println("Error parsing SDP:", err)
				panic(err)
			}

			// Return the SDP in a webrtc.SessionDescription
			return &sessionDescription, nil
		}
	}

	// Handle the case when the request is not successful
	return nil, fmt.Errorf("request was not successful")
}
