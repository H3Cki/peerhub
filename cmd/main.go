package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/H3Cki/sdphub"

	"github.com/gorilla/websocket"
)

type GetAnswerersResponse struct {
	Answerers []AnswererInfo
}

type AnswererInfo struct {
	Name        string
	Protected   bool
	Address     string
	Description string
	LastMessage time.Time
}

func main() {
	hub := sdphub.NewHub()

	http.HandleFunc("/answerers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		resp := GetAnswerersResponse{
			Answerers: []AnswererInfo{},
		}

		for _, a := range hub.Answerers() {
			resp.Answerers = append(resp.Answerers, AnswererInfo{
				Name:        a.Name,
				Protected:   a.AccessKey != "",
				Address:     a.Address,
				Description: a.Description,
				LastMessage: a.LastMessage,
			})
		}

		bytes, _ := json.Marshal(resp)
		w.Write(bytes)
	})

	http.HandleFunc("/signal", func(w http.ResponseWriter, r *http.Request) {
		u := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		conn, err := u.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer conn.Close()

		for {
			if err := handleMessage(conn, hub); err != nil {
				fmt.Println(err)
				return
			}
		}
	})

	http.ListenAndServe("localhost:8000", nil)
}

func handleMessage(conn *websocket.Conn, hub *sdphub.Hub) error {
	msg := sdphub.Message{}
	if err := conn.ReadJSON(&msg); err != nil {
		return err
	}

	ctx := sdphub.MessageContext{
		Conn:        conn,
		ConvID:      msg.ConvID,
		Addr:        conn.RemoteAddr().String(),
		MessageTime: time.Now(),
	}

	fmt.Println(msg.Type, msg.ConvID)

	switch msg.Type {
	case sdphub.MessageTypeCreateAnswerer:
		data := sdphub.CreateAnswererRequest{}
		if err := unmarshalData(msg.Data, &data); err != nil {
			sdphub.SendError(ctx, err)
			return nil
		}

		if err := hub.CreateAnswerer(ctx, data); err != nil {
			fmt.Println(err)
		}
	case sdphub.MessageTypeFindAnswerer:
		data := sdphub.CreateOffererRequest{}
		if err := unmarshalData(msg.Data, &data); err != nil {
			sdphub.SendError(ctx, err)
			return nil
		}

		if err := hub.SendOffer(ctx, data); err != nil {
			fmt.Println(err)
		}
	case sdphub.MessageTypeCreateOfferer:
		data := sdphub.CreateOffererRequest{}
		if err := unmarshalData(msg.Data, &data); err != nil {
			sdphub.SendError(ctx, err)
			return nil
		}

		if err := hub.CreateOfferer(ctx, data); err != nil {
			fmt.Println(err)
		}
	case sdphub.MessageTypeAcceptAgreement:
		data := sdphub.AcceptAgreementRequest{}
		if err := unmarshalData(msg.Data, &data); err != nil {
			sdphub.SendError(ctx, err)
			return nil
		}

		if err := hub.AcceptAgreement(ctx, data); err != nil {
			fmt.Println(err)
		}
	case sdphub.MessageTypeRejectAgreement:
		data := sdphub.RejectAgreementRequest{}
		if err := unmarshalData(msg.Data, &data); err != nil {
			sdphub.SendError(ctx, err)
			return nil
		}

		if err := hub.RejectAgreement(ctx, data); err != nil {
			fmt.Println(err)
		}
	default:
		sdphub.SendError(ctx, errors.New("unsupported message type"))
	}

	return nil
}

func unmarshalData(data any, v any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}
