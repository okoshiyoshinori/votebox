package model

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"time"

	"github.com/okoshiyoshinori/votebox/config"

	"github.com/okoshiyoshinori/votebox/util"
	"github.com/rs/xid"
)

// DB models

type User struct {
	ID          string    `json:"uid" gorm:"primary_key"`
	Nickname    string    `json:"nickname"`
	Avatar      string    `json:"avatar"`
	Description string    `json:"description"`
	Mail        string    `json:"mail"`
	LoginType   string    `json:"logintype"`
	Active      bool      `json:"active"`
	Token       string    `json:"token"`
	Secret      string    `json:"secret"`
	Boxes       []*Box    `json:"boxes"`
	UpdateAt    time.Time `json:"update"`
}

type Box struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	UserID string `json:"uid"`
	Title  string `json:"title"`
	//Active    bool      `json:"active"`
	StartTime time.Time `json:"starttime"`
	EndTime   time.Time `json:"endtime"`
	Type      string    `json:"type"`
	BoolImage bool      `json:"boolimage"`
	Votes     []Vote    `json:"votes"`
	Hid       *HashID   `json:"hid"`
	Result    *UserVote `json:"result"`
	Agency    bool      `json:"agency"`
	User      User      `json:"user"`
	UpdateAt  time.Time `json:"update"`
}

type Vote struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	BoxID     uint      `json:"bid"`
	ItemText  string    `json:"text"`
	ImagePath string    `json:"image"`
	Vote      int       `json:"votes"`
	UpdateAt  time.Time `json:"update"`
}

type UserVote struct {
	Vid uint `json:"id"`
}

type HashID struct {
	ID string `json:"id"`
}

// Form models

type MakeBoxForm struct {
	//	Uid       string         `json:"uid" binding:"required"`
	Title     string         `json:"title" binding:"required"`
	Type      string         `json:"type" binding:"required"`
	BoolImage bool           `json:"boolimage"`
	Agency    bool           `json:"agency"`
	Endtime   time.Time      `json:"endtime"`
	Votes     []VoteItemForm `json:"voteitem"`
}

type VoteItemForm struct {
	Text string `json:"text" binding:"required"`
	Imag string `json:"image"`
}

func (v *VoteItemForm) DecodeImage() (string, error) {
	if v.Imag == "" {
		return "", nil
	}
	guid := xid.New()
	name := guid.String()
	path := config.APConfig.Item + name
	image, err := base64.StdEncoding.DecodeString(v.Imag)
	if err != nil {
		return "", err
	}
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	file.Write(image)

	return name, nil

}

type UserForm struct {
	Uid         string `json:"uid" binding;"required"`
	Nickname    string `json:"nickname" binding:"required"`
	Avatar      string `json:"avatar"`
	Description string `json:"description" binding:"required"`
	Logintype   string `json:"logintype" binding:"required"`
	Mail        string `json:"mail" binding:"required,email"`
}

type CountUp struct {
	Id int `json:"id" binding:"required"`
}

type DeleteUser struct {
	Uid string `json:"uid" binding:"required"`
}

func (b *Box) SetHashID() {
	ht := new(HashID)
	ht.ID = util.Encode(int(b.ID))
	b.Hid = ht
}

type SocketData struct {
	Header  string      `json:header`
	Payload interface{} `json:payload`
}

func MakeData(h string, p interface{}) ([]byte, error) {
	s := new(SocketData)
	s.Header = h
	s.Payload = p
	d, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// other
type SessonInfo struct {
	UserID         interface{}
	NickName       interface{}
	IsSessionAlive interface{}
}

type VoteBoxID struct {
	Bid int `json:"bid"`
	Vid int `json:"vid"`
}

type MaxPage struct {
	Max int `json:"max"`
}

type Observer struct {
	Count int `json:"count"`
}

type CallbackRequest struct {
	Token    string `json:"oauth_token"`
	Verifier string `json:"oauth_verifier"`
}

type Account struct {
	ID              string `json:"screen_name"`
	ScreenName      string `json:"name"`
	ProfileImageURL string `json:"profile_image_url"`
	Email           string `json:"email"`
	Description     string `json:"description"`
}
