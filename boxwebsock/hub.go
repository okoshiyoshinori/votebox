package boxwebsock

import (
	"strconv"
	"time"

	"github.com/okoshiyoshinori/votebox/logger"
	"github.com/okoshiyoshinori/votebox/model"
)

var RootHub = NewHubs()

type Hub struct {
	Clients map[*Client]bool
}

type Hubs struct {
	Hubs             map[string]*Hub
	HubRegister      chan map[string]*Hub
	HubUnRegister    chan string
	ClientRegister   chan map[string]*Client
	ClientUnRegister chan map[string]*Client
	BroadCast        chan map[string][]byte
}

func NewHub() *Hub {
	return &Hub{
		Clients: make(map[*Client]bool),
	}
}

func NewHubs() *Hubs {
	return &Hubs{
		Hubs:             make(map[string]*Hub),
		HubRegister:      make(chan map[string]*Hub),
		HubUnRegister:    make(chan string),
		ClientRegister:   make(chan map[string]*Client),
		ClientUnRegister: make(chan map[string]*Client),
		BroadCast:        make(chan map[string][]byte),
	}
}

func init() {
	db := model.GetDBConn()
	boxs := []*model.Box{}
	/*
		db.Where("active = ?", true).Find(&boxs)
	*/
	db.Where("end_time >= ?", time.Now()).Find(&boxs)
	if len(boxs) <= 0 {
		return
	}
	for _, v := range boxs {
		v.SetHashID()
		RootHub.Hubs[v.Hid.ID] = NewHub()
	}
}

/*
func (h *Hub) Run() {
	log.Println("job run start")
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case <-h.Stop:
			for client := range h.Clients {
				client.Stop <- []byte{1}
			}
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.BroadCast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
*/

func (hs *Hubs) RegistHub(d map[string]*Hub) {
	for k, v := range d {
		hs.Hubs[k] = v
	}
}

func (hs *Hubs) UnregistHub(d string) {
	_, ok := hs.Hubs[d]
	if !ok {
		return
	}
	for client := range hs.Hubs[d].Clients {
		client.Stop <- []byte{}
	}
	go func() {
		defer delete(hs.Hubs, d)
		for {
			if len(hs.Hubs[d].Clients) == 0 {
				logger.Info.Println("box websocket delete")
				return
			}

		}
	}()
}

func (hs *Hubs) RegistClient(d map[string]*Client) {
	var k string
	var v *Client
	for k, v = range d {
		_, ok := hs.Hubs[k]
		if !ok {
			return
		}
		hs.Hubs[k].Clients[v] = true
	}
	var count int
	str, err := RedisClient.Get(k).Result()
	if err != nil {
		count = 0
	} else {
		count, _ = strconv.Atoi(str)
	}
	count++
	RedisClient.Set(k, count, 0)
	o := new(model.Observer)
	o.Count = count
	jsonByteData, err := model.MakeData("Observer", &o)
	if err != nil {
		return
	}
	_, ok := hs.Hubs[k]
	if !ok {
		return
	}
	for client := range hs.Hubs[k].Clients {
		select {
		case client.Send <- jsonByteData:
		}
	}
}

func (hs *Hubs) UnregistClient(d map[string]*Client) {
	var k string
	var v *Client
	for k, v = range d {
		_, ok := hs.Hubs[k]
		if !ok {
			return
		}
		delete(hs.Hubs[k].Clients, v)
	}
	var count int
	str, err := RedisClient.Get(k).Result()
	if err != nil {
		count = 0
	} else {
		count, _ = strconv.Atoi(str)
	}
	count--
	if count < 0 {
		count = 0
	}
	RedisClient.Set(k, count, 0)
	o := new(model.Observer)
	o.Count = count
	jsonByteData, err := model.MakeData("Observer", &o)
	if err != nil {
		return
	}
	_, ok := hs.Hubs[k]
	if !ok {
		return
	}
	for client := range hs.Hubs[k].Clients {
		select {
		case client.Send <- jsonByteData:
		}
	}
}

func (hs *Hubs) SendMessage(d map[string][]byte) {
	var box string
	var message []byte
	for k, v := range d {
		box, message = k, v
	}
	for client := range hs.Hubs[box].Clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(hs.Hubs[box].Clients, client)
		}
	}
}

func (hs *Hubs) Run() {
	for {
		select {
		case hub := <-hs.HubRegister:
			hs.RegistHub(hub)
		case hub := <-hs.HubUnRegister:
			hs.UnregistHub(hub)
		case client := <-hs.ClientRegister:
			hs.RegistClient(client)
		case client := <-hs.ClientUnRegister:
			hs.UnregistClient(client)
		case broadcast := <-hs.BroadCast:
			hs.SendMessage(broadcast)
		}
	}
}
