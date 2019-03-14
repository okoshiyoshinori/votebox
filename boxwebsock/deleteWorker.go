package boxwebsock

import (
	"time"

	"github.com/okoshiyoshinori/votebox/logger"
	"github.com/okoshiyoshinori/votebox/model"
)

var (
	Dworker  DeleteWorker
	Duration = 60 * time.Minute
)

type DeleteWorker struct {
	ticker *time.Ticker
}

func (d *DeleteWorker) getCloseBox() []*model.Box {
	db := model.GetDBConn()
	boxs := []*model.Box{}
	t := time.Now()
	t2 := t.Add(-(Duration))
	db.Where("end_time <= ? AND end_time >= ?", t, t2).Find(&boxs)
	return boxs
}

func (d *DeleteWorker) sendCloseMessage(b []*model.Box) {
	for _, v := range b {
		v.SetHashID()
		RootHub.HubUnRegister <- v.Hid.ID
	}
}

func (d *DeleteWorker) Run() {
	defer d.ticker.Stop()
	for {
		select {
		case <-d.ticker.C:
			logger.Info.Println("websocket hub delete start")
			box := d.getCloseBox()
			if len(box) > 0 {
				d.sendCloseMessage(box)
				logger.Info.Printf("%d boxes", len(box))
			}
			logger.Info.Println("websocket hub delete end")
		}
	}
}

func init() {
	tick := time.NewTicker(Duration)
	Dworker = DeleteWorker{
		ticker: tick,
	}
}
