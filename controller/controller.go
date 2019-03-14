package controller

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/okoshiyoshinori/votebox/votetoken"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/okoshiyoshinori/votebox/boxwebsock"
	"github.com/okoshiyoshinori/votebox/config"
	"github.com/okoshiyoshinori/votebox/logger"
	"github.com/okoshiyoshinori/votebox/model"
	"github.com/okoshiyoshinori/votebox/textimage"
	"github.com/okoshiyoshinori/votebox/twitter"
	"github.com/okoshiyoshinori/votebox/util"
)

func GetBoxController(c *gin.Context) {
	p, err := strconv.Atoi(c.Query("p"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	pernum := config.APConfig.PerNum //per page count
	limit := pernum
	offset := (p - 1) * limit
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	boxs := make([]*model.Box, 0, limit)
	db := model.GetDBConn()
	db.Limit(limit).Order("start_time desc").Offset(offset).Find(&boxs)
	if len(boxs) <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"Error": "box is not found"})
		return
	}
	for _, v := range boxs {
		v.SetHashID()
	}
	c.JSON(http.StatusOK, &boxs)
}

func GetBoxCount(c *gin.Context) {
	db := model.GetDBConn()
	obj := new(model.MaxPage)
	var count interface{}
	db.Table("boxes").Count(&count)
	d, err := strconv.Atoi(string(count.([]byte)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	temp := float64(d) / float64(config.APConfig.PerNum)
	obj.Max = int(math.Ceil(temp))
	c.JSON(200, &obj)
}

func GetBoxByIdController(c *gin.Context) {
	id, err := util.Decode(c.Param("bid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest"})
		return
	}
	/*
		id, err := strconv.Atoi(c.Param("bid"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
			return
		}
	*/
	var Box model.Box
	db := model.GetDBConn()
	if db.Where("id = ?", id).Find(&Box).RecordNotFound() {
		c.JSON(http.StatusNotFound, gin.H{"Error": "該当の投票箱はありません。"})
		return
	}
	db.Model(&Box).Related(&Box.User).Related(&Box.Votes)
	Box.SetHashID()
	Box.User.Secret = ""
	Box.User.Token = ""
	session := sessions.Default(c)
	se := session.Get("voteid")
	if se != nil {
		cstring := se.([]byte)
		var vids = []*model.VoteBoxID{}
		json.Unmarshal(cstring, &vids)
		for _, v := range Box.Votes {
			for _, x := range vids {
				if v.ID == uint(x.Vid) {
					d := new(model.UserVote)
					d.Vid = v.ID
					Box.Result = d
					break
				}
			}
		}
	}
	c.JSON(http.StatusOK, &Box)
}

func GetUserByIdController(c *gin.Context) {
	//id := c.Param("uid")
	id, err := votetoken.TokenAuthentication(c)
	if err != nil {
		c.JSON(401, gin.H{"Error": "認証に失敗しました。"})
	}
	user := model.User{}
	db := model.GetDBConn()
	if db.Select("id,nickname,avatar,mail,description,login_type,active,update_at").Where("id= ?", id).Find(&user).RecordNotFound() {
		c.JSON(http.StatusNotFound, gin.H{"Error": "user with username " + id + " not found"})
		return
	}
	db.Model(&user).Order("start_time desc").Related(&user.Boxes)
	for _, v := range user.Boxes {
		v.SetHashID()
	}
	c.JSON(http.StatusOK, &user)

}

/*
func GetBoxUserByIdController(c *gin.Context) {
	id := c.Param("uid")
	db := model.GetDBConn()
	user := new(model.User)
	if db.Where("id = ?", id).Find(&user).RecordNotFound() {
		c.JSON(http.StatusNotFound, gin.H{"Error": "user with username " + id + " not found"})
		return
	}
	db.Model(&user).Order("start_time desc").Related(&user.Boxes)
	c.JSON(http.StatusOK, &user)
}
*/

func PostBoxController(c *gin.Context) {
	form := model.MakeBoxForm{}
	if err := c.Bind(&form); err != nil {
		logger.Error.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest"})
		return
	}
	db := model.GetDBConn()
	tx := db.Begin()
	if tx.Error != nil {
		logger.Error.Println(tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	/*
		endtime, err := time.Parse("2016-01-01 23:23:19", form.Endtime)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
			return
		}
	*/
	id, err := votetoken.TokenAuthentication(c)
	if err != nil {
		c.JSON(401, gin.H{"Error": "認証に失敗しました。"})
	}
	box := model.Box{
		UserID: id,
		Title:  form.Title,
		//Active:    true,
		StartTime: time.Now(),
		EndTime:   form.Endtime,
		Type:      form.Type,
		BoolImage: form.BoolImage,
		UpdateAt:  time.Now(),
		Agency:    form.Agency,
	}

	if err := tx.Create(&box).Error; err != nil {
		tx.Rollback()
		logger.Error.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	for _, v := range form.Votes {
		path, err := v.DecodeImage()
		if err != nil {
			logger.Error.Println(err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
			return
		}
		d := new(model.Vote)
		d.BoxID = box.ID
		d.ItemText = v.Text
		d.ImagePath = path
		d.Vote = 0
		d.UpdateAt = time.Now()
		if err := tx.Create(d).Error; err != nil {
			tx.Rollback()
			logger.Error.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
			return
		}
	}
	tx.Commit()

	box.SetHashID()

	//hub := boxwebsock.NewHub()
	//	m := make(map[string]*boxwebsock.Hub)
	//	m[box.Hid.ID] = hub
	//boxwebsock.RootHub.HubRegister <- m
	boxwebsock.RedisClient.Publish("regist", box.Hid.ID)
	me := make(map[string]string)
	me[box.Hid.ID] = box.Title
	textimage.Sub.Question <- me
	//tweet
	/*
		user := &model.User{}
		db.Where("id = ?", box.UserID).Find(&user)
		client := twitter.InitClient()
		twitter.PostTweet(client, &box)
	*/
	c.JSON(http.StatusOK, gin.H{"status": "更新成功"})
}

func PostUserController(c *gin.Context) {
	d := model.UserForm{}
	if err := c.Bind(&d); err != nil {
		logger.Error.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	id, err := votetoken.TokenAuthentication(c)
	if err != nil {
		c.JSON(401, gin.H{"Error": "認証に失敗しました。"})
	}
	db := model.GetDBConn()
	if db.Where("id = ? AND mail = ? AND login_type = ?", id, d.Mail, d.Logintype).RecordNotFound() {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
		return
	}
	tx := db.Begin()
	if tx.Error != nil {
		logger.Error.Println(tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}

	user := new(model.User)
	user.ID = id
	user.Nickname = d.Nickname
	user.Avatar = d.Avatar
	user.Description = d.Description
	user.Active = true
	user.Mail = d.Mail
	user.LoginType = d.Logintype
	user.UpdateAt = time.Now()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		logger.Error.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"status": "更新成功"})
}

func PutVoteController(c *gin.Context) {
	session := sessions.Default(c)
	se := session.Get("voteid")
	var vids []*model.VoteBoxID
	if se == nil {
		vids = []*model.VoteBoxID{}
	} else {
		cstring := se.([]byte)
		json.Unmarshal(cstring, &vids)
	}
	f := new(model.CountUp)
	if err := c.Bind(&f); err != nil {
		logger.Error.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}

	db := model.GetDBConn()
	v := new(model.Vote)
	if db.Where("id = ?", f.Id).Find(&v).RecordNotFound() {
		c.JSON(http.StatusNotFound, gin.H{"status": "選択項目がありません"})
		return
	}
	b := new(model.Box)
	if db.Find(&b, v.BoxID).RecordNotFound() {
		c.JSON(http.StatusNotFound, gin.H{"status": "選択項目がありません"})
		return
	}

	now := time.Now()

	if now.Equal(b.EndTime) || now.After(b.EndTime) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "投票箱は締め切られました"})
		return
	}

	new_v := new(model.Vote)
	new_v.Vote = v.Vote + 1
	new_v.UpdateAt = time.Now()
	tx := db.Begin()
	if tx.Error != nil {
		logger.Error.Println(tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	if err := tx.Model(&v).Updates(&new_v).Error; err != nil {
		tx.Rollback()
		logger.Error.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	tx.Commit()

	// cookie登録
	cs := model.VoteBoxID{
		Bid: int(v.BoxID),
		Vid: int(v.ID),
	}
	vids = append(vids, &cs)
	data, err := json.Marshal(vids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	session.Set("voteid", data)
	if err := session.Save(); err != nil {
		logger.Error.Println(err)
	}

	// 投票通知
	db.Model(&b).Related(&b.Votes)
	b.SetHashID()
	jsonByteData, err := model.MakeData("result", b.Votes)
	boxwebsock.RedisClient.Publish("result", jsonByteData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	/*
		s := make(map[string][]byte)
		s[b.Hid.ID] = jsonByteData
		boxwebsock.RootHub.BroadCast <- s
	*/
	// c.SetCookie("voteid", string(data), 60*60*24*60, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"status": "更新成功"})
}

func PutUserController(c *gin.Context) {
	f := new(model.UserForm)
	if err := c.Bind(&f); err != nil {
		logger.Error.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest"})
		return
	}
	id, err := votetoken.TokenAuthentication(c)
	if err != nil {
		c.JSON(401, gin.H{"Error": "認証に失敗しました。"})
	}
	f.Uid = id
	db := model.GetDBConn()
	user := new(model.User)
	if db.Where("id = ?", f.Uid).Find(&user).RecordNotFound() {
		c.JSON(http.StatusNotFound, gin.H{"status": "該当ユーザーが存在しません"})
		return
	}
	new_user := new(model.User)
	new_user.Nickname = f.Nickname
	new_user.Avatar = f.Avatar
	new_user.Description = f.Description
	new_user.LoginType = f.Logintype
	new_user.Mail = f.Mail
	new_user.UpdateAt = time.Now()
	tx := db.Begin()
	if tx.Error != nil {
		logger.Error.Println(tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	if err := tx.Model(&user).Updates(&new_user).Error; err != nil {
		tx.Rollback()
		logger.Error.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, &user)
}

/*
func DeleteUserController(c *gin.Context) {
	f := new(model.DeleteUser)
	if err := c.Bind(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Internal Error"})
		log.Println(err)
		return
	}

	user := new(model.User)
	db := model.GetDBConn()
	if db.Where("id = ?", f.Uid).Find(&user).RecordNotFound() {
		c.JSON(http.StatusNotFound, gin.H{"status": "該当ユーザーが存在しません"})
		return
	}
	if user.Active == false {
		c.JSON(http.StatusNotFound, gin.H{"status": "該当ユーザーが存在しません"})
		return
	}
	tx := db.Begin()
	if tx.Error != nil {
		tx.Rollback()
		log.Println(tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	new_user := new(model.User)
	new_user.Active = false
	new_user.UpdateAt = time.Now()
	tx.Model(&user).Update("active", false)
	tx.Model(&user).Update("updateAt", time.Now())
		if err := tx.Model(&user).Updates(&new_user).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
			return
		}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"status": "削除完了しました"})
}
*/

func DeleteUserController(c *gin.Context) {
	//userID := c.Param("uid")
	userID, err := votetoken.TokenAuthentication(c)
	if err != nil {
		c.JSON(401, gin.H{"Error": "認証に失敗しました。"})
	}
	db := model.GetDBConn()
	boxs := []*model.Box{}
	vote := []model.Vote{}
	votes := []model.Vote{}
	db.Where("user_id = ?", userID).Find(&boxs)
	for _, v := range boxs {
		v.SetHashID()
		db.Where("box_id= ?", v.ID).Find(&vote)
		//	db.Where("box_id = ?", v.ID).Delete(model.Vote{})
		votes = append(votes, vote...)
	}
	tx := db.Begin()
	for _, v := range votes {
		if err := tx.Delete(model.Vote{}, "id = ?", v.ID).Error; err != nil {
			tx.Rollback()
			logger.Error.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
			return
		}
	}
	for _, v := range boxs {
		if err := tx.Delete(model.Box{}, "id = ?", v.ID).Error; err != nil {
			tx.Rollback()
			logger.Error.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
			return
		}
	}
	if err := tx.Delete(model.User{}, "id = ?", userID).Error; err != nil {
		tx.Rollback()
		logger.Error.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	tx.Commit()
	for _, v := range boxs {
		if err := os.Remove(config.APConfig.Ogb + v.Hid.ID + ".jpeg"); err != nil {
			logger.Error.Println(err)
		}
	}
	for _, v := range votes {
		if err := os.Remove(config.APConfig.Item + v.ImagePath); err != nil {
			continue
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "退会が完了しました"})

}

func DeleteBoxController(c *gin.Context) {
	boxID, err := util.Decode(c.Param("bid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest"})
		return
	}
	/*
		boxID, err := strconv.Atoi(c.Param("bid"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "無効な引数です"})
			return
		}
	*/
	//vote delete
	db := model.GetDBConn()
	vox := []*model.Vote{}
	db.Where("box_id = ?", boxID).Find(&vox)
	db.Where("box_id = ?", boxID).Delete(model.Vote{})
	//box delete
	db.Where("id = ?", boxID).Delete(model.Box{})
	// websocket 一斉接続解除
	boxwebsock.RedisClient.Publish("delete", c.Param("bid"))
	// boxwebsock.RootHub.HubUnRegister <- c.Param("bid")
	//画像削除
	for _, v := range vox {
		if err := os.Remove(config.APConfig.Item + v.ImagePath); err != nil {
			continue
		}
	}
	if err := os.Remove(config.APConfig.Ogb + c.Param("bid") + ".jpeg"); err != nil {
		logger.Error.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "削除完了しました"})
}

func WsServeController(c *gin.Context) {
	box := c.Param("bid")
	db := model.GetDBConn()
	bid, err := util.Decode(box)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest"})
		return
	}
	if db.Where("id = ?", bid).Find(&model.Box{}).RecordNotFound() {
		c.JSON(http.StatusNotFound, gin.H{"status": "ボックスがありません"})
		return
	}
	if db.Where("id = ? and end_time >= ?", bid, time.Now()).Find(&model.Box{}).RecordNotFound() {
		c.JSON(http.StatusBadRequest, gin.H{"status": "投票箱が締め切られています"})
		return
	}
	conn, err := boxwebsock.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	client := &boxwebsock.Client{
		Hubs: boxwebsock.RootHub,
		Conn: conn,
		Box:  box,
		Send: make(chan []byte),
		Stop: make(chan []byte),
	}
	s := make(map[string]*boxwebsock.Client)
	s[box] = client
	client.Hubs.ClientRegister <- s
	go client.ReadPump()
	go client.WritePump()

}

func TwitterOauth(c *gin.Context) {
	session := sessions.Default(c)
	configOauth := twitter.GetConnect()
	rt, err := configOauth.RequestTemporaryCredentials(nil, config.APConfig.CallbackUrl, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "リクエストが不正です"})
		return
	}
	session.Set("request_token", rt.Token)
	session.Set("request_token_secret", rt.Secret)
	if err := session.Save(); err != nil {
		logger.Error.Println(err)
	}
	url := configOauth.AuthorizationURL(rt, nil)
	c.JSON(http.StatusOK, &url)
}

func CallbackController(c *gin.Context) {
	session := sessions.Default(c)
	request := model.CallbackRequest{}
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "リクエストが不正です"})
		return
	}
	at, err := twitter.GetAccessToken(
		&oauth.Credentials{
			Token:  session.Get("request_token").(string),
			Secret: session.Get("request_token_secret").(string),
		},
		request.Verifier,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "予期せぬエラーが発生しました"})
		return
	}
	account := model.Account{}
	if err = twitter.GetMe(at, &account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "予期せぬエラーが発生しました"})
		return
	}
	// session.Set("oauth_token", at.Token)
	// session.Set("oauth_token_secret", at.Secret)
	/*
		if err := session.Save(); err != nil {
			logger.Error.Println(err)
		}
	*/
	beforeUser := model.User{}
	var isFound bool = false

	db := model.GetDBConn()
	if !db.Where("id = ? and mail = ?", account.ID, account.Email).Find(&beforeUser).RecordNotFound() {
		isFound = true
	}

	user := &model.User{
		ID:          account.ID,
		Nickname:    account.ScreenName,
		Description: account.Description,
		Mail:        account.Email,
		Avatar:      account.ProfileImageURL,
		Token:       at.Token,
		Secret:      at.Secret,
		LoginType:   "twitter",
		Active:      true,
		UpdateAt:    time.Now(),
	}
	tx := db.Begin()
	if tx.Error != nil {
		logger.Error.Println(tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
		return
	}
	if isFound {
		if err := tx.Model(&beforeUser).Update(&user).Error; err != nil {
			logger.Error.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
			return
		}
	} else {
		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			logger.Error.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Error"})
			return
		}
	}
	tx.Commit()
	c.JSON(http.StatusOK, votetoken.CreateToken(account.ID))
}
