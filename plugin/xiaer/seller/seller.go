package seller

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/database/mongo"
	"github.com/FloatTech/ZeroBot-Plugin/database/redis"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/xiaer"

	logger "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const id = "seller"
const name = "买卖"
const TotalCount = 10
const FalseEmoji = "❌"
const TrueEmoji = "✔️"

type Operation struct {
	Name   string
	Cost   float64 // 花费
	Salary float64 // 薪水
	Worth  float64 // 随机身价
	Count  int     // 消耗点数
}

var OperationData = map[string]Operation{
	"安抚":  {"安抚", 0, 0, 0, 1},
	"臭骂":  {"一顿臭骂", 0, 0, 0, 1},
	"板砖":  {"拿板砖拍晕", 30, 0, 0, 1},
	"挖煤":  {"去当家教或去黑煤窑挖煤", 150, 50, 0, 5},
	"卖唱":  {"去歌厅卖唱", 30, 80, 0, 5},
	"保姆":  {"去当小保姆", 30, 100, 0, 5},
	"摊贩":  {"去当小摊贩", 40, 100, 0, 5},
	"活埋":  {"挖了坑活埋", 0, 0, 50, 10},
	"饿":   {"饿了3天3夜", 0, 0, 100, 10},
	"钢管舞": {"跳一段钢管舞", 100, 0, 0, 10},
	"充电":  {"去学习充电", 0, 42, 0, 5},
}

func init() {
	engine := control.Register(id, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "打工\n" +
			"- 打工仔签到 打工仔市场\n" +
			"- 打工仔列表 操作列表\n" +
			"- 购买打工仔@一个人\n" +
			"- 出售打工仔@一个人\n" +
			"- 安排安抚@一个人\n",
	})
	engine.OnFullMatch("打工仔签到").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			sellSign(ctx)
		})
	engine.OnFullMatch("打工仔市场").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			slaveMarket(ctx)
		})
	engine.OnFullMatch("打工仔列表").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			slaveList(ctx)
		})
	engine.OnFullMatch("操作列表").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			operationList(ctx)
		})
	engine.OnPrefix("购买打工仔").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			buySlaves(ctx)
		})
	engine.OnPrefix("出售打工仔").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			sellSlaves(ctx)
		})
	engine.OnPrefix("安排").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			txt := ctx.State["args"].(string)
			if txt != "" {
				operateSlaves(ctx, txt)
			}
		})
}

func slaveList(ctx *zero.Ctx) {
	currentUser := loadUserInfo(ctx.Event.GroupID, ctx.Event.Sender.ID)
	result := "打工仔列表 : \n"
	for _, slaverId := range currentUser.Slaver {
		targetNickname := xiaer.CardNameInGroup(ctx, slaverId)
		result += fmt.Sprintf(" %v \n\n", targetNickname)
	}

	result += fmt.Sprintf("我的资产: %.0f, 身价：%.0f, 操作点: %d", currentUser.Money, currentUser.Worth, currentUser.OperationCount)

	ctx.SendChain(message.Text(
		result,
	))
}

func slaveMarket(ctx *zero.Ctx) {
	ctx1, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := mongo.Collection("game.seller.user")
	cur, err := coll.Find(ctx1, bson.M{"groupCode": ctx.Event.GroupID})
	if err != nil {
		panic(err)
	}
	result := "打工仔市场 : \n\n 打工仔  身价  老板  可交易\n"

	defer cur.Close(ctx1)

	nowTime := time.Now()
	for cur.Next(ctx1) {
		var user User
		err = cur.Decode(&user)
		if err != nil {
			panic(err)
		}
		targetNickname := xiaer.CardNameInGroup(ctx, user.Uin)
		if user.Master > 0 {
			ownerNickName := xiaer.CardNameInGroup(ctx, user.Master)
			result += fmt.Sprintf(" %v || %.0f || %s || %s\n", targetNickname, user.Worth, ownerNickName, getTrueFalse(user.GuardTime.Before(nowTime)))
		} else {
			result += fmt.Sprintf(" %v || %.0f ||    || %s\n", targetNickname, user.Worth, getTrueFalse(user.GuardTime.Before(nowTime)))
		}
	}

	ctx.SendChain(message.Text(
		result,
	))
}

func getTrueFalse(after bool) string {
	if after {
		return TrueEmoji
	} else {
		return FalseEmoji
	}
}

func operationList(ctx *zero.Ctx) {
	result := "操作列表 : \n"

	for key, operation := range OperationData {
		result += fmt.Sprintf("%s: 花费%.0f 薪资%.0f/h 身价±%.0f 点数%d\n", key, operation.Cost, operation.Salary, operation.Salary, operation.Count)
	}

	ctx.SendChain(message.Text(result))
}

func sellSign(ctx *zero.Ctx) {
	day := time.Now()
	today := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())

	lock, err := redis.TryLock(fmt.Sprintf("SELLER::LOCK::%v", ctx.Event.GroupID), time.Second*5, time.Second*15)
	if err != nil {
		return
	}
	defer lock.Unlock()
	last := loadUserInfo(ctx.Event.GroupID, ctx.Event.Sender.ID)
	// 检查当前时间和最后登录时间对比
	if last == nil {
		last = &User{GroupCode: ctx.Event.GroupID, Uin: ctx.Event.Sender.ID, Worth: 500, Money: 2000, Count: TotalCount, Slaver: make([]int64, 0)}
	}

	// 正常签到 签到加500个金币
	if last.LastLoginTime.Before(today) {
		if last.Worth <= 3000 {
			last.Worth += 300
		} else if last.Worth <= 5000 {
			last.Worth += 360
		} else if last.Worth <= 8000 {
			last.Worth += 400
		} else if last.Worth <= 10000 {
			last.Worth += 500
		} else {
			last.Worth += 560
		}

		last.Money += 1000
		last.LastLoginTime = day
		last.Count = TotalCount - len(last.Slaver)
	} else {
		ctx.SendChain(message.Text("您今天已经签到过"))
		return
	}

	updateUser(ctx.Event.GroupID, ctx.Event.Sender.ID, last)

	ctx.SendChain(message.Text(
		fmt.Sprintf(
			"签到成功 : \n"+
				" 金钱 : %.0f \n"+
				" 身价 : %.0f \n"+
				" 可购买次数 : %v \n"+
				" 打工仔数量: %d \n",
			last.Money, last.Worth, last.Count, len(last.Slaver),
		),
	))
}

func operateSlaves(ctx *zero.Ctx, content string) {
	_, at := xiaer.GetFirstAt(ctx)
	var op = Operation{}
	for key, operation := range OperationData {
		if strings.Contains(content, key) {
			op = operation
		}
	}
	if (Operation{} == op) {
		ctx.SendChain(message.Text("您需要发送 安排安抚@一个人 才能操作给自己打工的打工仔"))
		return
	}

	if at == 0 {
		ctx.SendChain(message.Text("您需要发送 安排操作@一个人 才能操作给自己打工的打工仔"))
		return
	}
	lock, err := redis.TryLock(fmt.Sprintf("SELLER::LOCK::%v", ctx.Event.GroupID), time.Second*5, time.Second*15)
	if err != nil {
		return
	}
	defer lock.Unlock()

	if ctx.Event.Sender.ID == at {
		ctx.SendChain(message.Text("无法操作自己"))
		return
	}
	seller := loadUserInfo(ctx.Event.GroupID, ctx.Event.Sender.ID)

	if seller == nil {
		ctx.SendChain(message.Text("请先完成第一次签到"))
		return
	}
	if seller.Count <= 0 {
		ctx.SendChain(message.Text("您当前购买次数不足"))
		return
	}
	slaver := loadUserInfo(ctx.Event.GroupID, at)
	if slaver == nil {
		ctx.SendChain(message.Text("目标用户暂未参与该游戏"))
		return
	}
	if slaver.Master != ctx.Event.Sender.ID {
		ctx.SendChain(message.Text("当前打工仔不是你的打工仔"))
		return
	}

	if seller.OperationCount < op.Count {
		ctx.SendChain(message.Text(fmt.Sprintf("当前操作点数不足需要%d, 拥有%d", op.Count, seller.OperationCount)))
		return
	}

	seller.OperationCount -= op.Count

	if seller.Money >= op.Cost {
		if op.Cost > 0 {
			seller.Money -= op.Cost
		}
	} else {
		ctx.SendChain(message.Text(fmt.Sprintf("当前操作资产不足%.0f, 拥有%.0f", op.Cost, seller.Money)))
		return
	}

	if op.Salary > 0 {
		slaver.Salary = op.Salary
		slaver.WorkStartTime = time.Now()
	} else {
		if !slaver.WorkStartTime.IsZero() {
			slaver.WorkStartTime = time.Time{}
		}
	}

	if op.Worth > 0 {
		isAdd := rand.Int()%2 == 0
		if isAdd {
			slaver.Worth -= op.Worth
		} else {
			slaver.Worth += op.Worth
		}
	}

	updateUser(ctx.Event.GroupID, slaver.Uin, slaver)
	updateUser(ctx.Event.GroupID, seller.Uin, seller)

	atMemberNickName := xiaer.CardNameInGroup(ctx, at)
	sprintf := fmt.Sprintf(
		"操作成功 : \n"+
			" 安排【%s】 【%s】\n"+
			" 操作花费 : %.0f \n"+
			" 可操作点数: %d \n",
		atMemberNickName, op.Name, op.Cost, seller.OperationCount,
	)
	logger.Error(sprintf)
	ctx.SendChain(message.Text(
		sprintf,
	))
}

func buySlaves(ctx *zero.Ctx) {
	has, at := xiaer.GetFirstAt(ctx)
	if !has || at == 0 {
		ctx.SendChain(message.Text("您需要发送 购买打工仔并@一个人 才能购买他人作为打工仔"))
		return
	}
	lock, err := redis.TryLock(fmt.Sprintf("SELLER::LOCK::%v", ctx.Event.GroupID), time.Second*5, time.Second*15)
	if err != nil {
		return
	}
	defer lock.Unlock()

	if ctx.Event.Sender.ID == at {
		ctx.SendChain(message.Text("无法购买自己"))
		return
	}
	seller := loadUserInfo(ctx.Event.GroupID, ctx.Event.Sender.ID)
	if seller == nil {
		ctx.SendChain(message.Text("请先完成第一次签到"))
		return
	}
	if seller.Count <= 0 {
		ctx.SendChain(message.Text("您当前购买次数不足"))
		return
	}
	if seller.Master == at {
		ctx.SendChain(message.Text("无法购买自己的老板"))
		return
	}
	slaver := loadUserInfo(ctx.Event.GroupID, at)
	if slaver == nil {
		ctx.SendChain(message.Text("目标用户暂未参与该游戏"))
		return
	}
	if slaver.Master == seller.Uin {
		ctx.SendChain(message.Text("当前打工仔已经被您购买"))
		return
	}
	if slaver.GuardTime.After(time.Now()) {
		ctx.SendChain(message.Text("当前打工仔处于购买30分钟保护期"))
		return
	}
	if seller.Money < slaver.Worth {
		ctx.SendChain(message.Text(fmt.Sprintf("您当前资产小于打工仔身价: %.0f < %.0f", seller.Money, slaver.Worth)))
		return
	}
	logger.Info("开始购买")
	consumption := slaver.Worth + 50
	seller.Money -= consumption // 购买者扣除打工仔身价 + 50金币手续费

	var salaryMoney float64 = 0
	if slaver.Master > 0 {
		masterUser := loadUserInfo(ctx.Event.GroupID, slaver.Master)
		if masterUser != nil {
			masterUser.Money += slaver.Worth
			if !slaver.WorkStartTime.IsZero() {
				salaryMoney = slaver.Salary * math.Floor(time.Now().Sub(slaver.WorkStartTime).Hours())
				masterUser.Money += salaryMoney
			}
			newSlaverIds := make([]int64, 0)
			for _, slaverId := range masterUser.Slaver {
				if slaver.Uin != slaverId {
					newSlaverIds = append(newSlaverIds, slaverId)
				}
			}
			masterUser.Slaver = newSlaverIds
			// 更新原来的Master的金钱
			updateUser(ctx.Event.GroupID, masterUser.Uin, masterUser)
		}
	}
	slaver.Master = seller.Uin
	slaver.WorkStartTime = time.Time{}
	slaver.TormentCount = 10
	slaver.GuardTime = time.Now().Add(30 * time.Minute)
	slaver.Worth += 50
	seller.LastBuyTime = time.Now()
	if seller.OperationCount <= 40 {
		seller.OperationCount += 10
	}

	seller.Count -= 1
	seller.Slaver = append(seller.Slaver, slaver.Uin)
	updateUser(ctx.Event.GroupID, seller.Uin, seller)
	updateUser(ctx.Event.GroupID, slaver.Uin, slaver)

	logger.Error(consumption, slaver.Worth, slaver.Worth+salaryMoney, seller.Count, len(seller.Slaver))
	sprintf := fmt.Sprintf(
		"购买成功 : \n"+
			" 购买花费 : %.0f \n"+
			" 打工仔身价 : %.0f \n"+
			" 原老板获得 : %.0f \n"+
			" 可购买次数 : %v \n"+
			" 打工仔数量: %d \n",
		consumption, slaver.Worth, slaver.Worth+salaryMoney, seller.Count, len(seller.Slaver),
	)
	logger.Error(sprintf)
	ctx.SendChain(message.Text(sprintf))
}

func sellSlaves(ctx *zero.Ctx) {
	at := int64(0)
	hasAt, fAt := xiaer.GetFirstAt(ctx)
	if hasAt {
		at = fAt
	}

	if at == 0 {
		ctx.SendChain(message.Text("您需要发送 出售打工仔并@一个人 才能出售他人作为打工仔"))
		return
	}
	lock, err := redis.TryLock(fmt.Sprintf("SELLER::LOCK::%v", ctx.Event.GroupID), time.Second*5, time.Second*15)
	if err != nil {
		return
	}
	defer lock.Unlock()

	if ctx.Event.Sender.ID == at {
		ctx.SendChain(message.Text("无法出售自己"))
		return
	}

	seller := loadUserInfo(ctx.Event.GroupID, ctx.Event.Sender.ID)
	if seller == nil {
		ctx.SendChain(message.Text("请先完成第一次签到"))
		return
	}

	isExistIndex := -1
	for i, slaverId := range seller.Slaver {
		if slaverId == at {
			isExistIndex = i
		}
	}
	if isExistIndex <= -1 {
		ctx.SendChain(message.Text("无法出售非自己的打工仔"))
		return
	}

	slaver := loadUserInfo(ctx.Event.GroupID, at)

	if slaver == nil {
		ctx.SendChain(message.Text("目标用户暂未参与该游戏"))
		return
	}

	if slaver.Master != seller.Uin {
		ctx.SendChain(message.Text("无法出售非自己的打工仔"))
		return
	}
	if slaver.GuardTime.After(time.Now()) {
		ctx.SendChain(message.Text("当前打工仔处于购买30分钟保护期"))
		return
	}
	logger.Info("开始出售")
	var salaryMoney = slaver.Worth * 0.8

	if !slaver.WorkStartTime.IsZero() {
		salaryMoney += slaver.Salary * time.Now().Sub(slaver.WorkStartTime).Hours()
	}
	seller.Money += salaryMoney

	newSlaverIds := make([]int64, 0)
	for _, slaverId := range seller.Slaver {
		if slaver.Uin != slaverId {
			newSlaverIds = append(newSlaverIds, slaverId)
		}
	}
	seller.Slaver = newSlaverIds

	slaver.Master = 0
	slaver.WorkStartTime = time.Time{}
	slaver.TormentCount = 10
	slaver.GuardTime = time.Time{}

	slaver.Worth -= 10
	if seller.OperationCount <= 40 {
		seller.OperationCount += 10
	}

	updateUser(ctx.Event.GroupID, seller.Uin, seller)
	updateUser(ctx.Event.GroupID, slaver.Uin, slaver)

	sprintf := fmt.Sprintf(
		"出售成功 : \n"+
			" 出售获得 : %.0f \n"+
			" 打工仔身价 : %.0f \n"+
			" 可购买次数 : %v \n"+
			" 打工仔数量: %d \n",
		salaryMoney, slaver.Worth, seller.Count, len(seller.Slaver),
	)
	logger.Error(sprintf)
	ctx.SendChain(message.Text(sprintf))
}

func loadUserInfo(groupCode int64, uin int64) *User {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := mongo.Collection("game.seller.user")
	cur, err := coll.Find(ctx, bson.M{"groupCode": groupCode, "uin": uin})
	if err != nil {
		panic(err)
	}
	defer cur.Close(ctx)
	if cur.Next(ctx) {
		var user User
		err = cur.Decode(&user)
		if err != nil {
			panic(err)
		}
		return &user
	} else {
		return nil
	}
}

func updateUser(code int64, uin int64, user *User) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := mongo.Collection("game.seller.user")
	_, err := coll.UpdateOne(
		ctx,
		bson.M{"uin": uin, "groupCode": code},
		bson.M{"$set": bson.M{
			"lastLoginTime":  user.LastLoginTime,
			"worth":          user.Worth,
			"count":          user.Count,
			"lastBuyTime":    user.LastBuyTime,
			"money":          user.Money,
			"master":         user.Master,
			"slaver":         user.Slaver,
			"guardTime":      user.GuardTime,
			"tormentCount":   user.TormentCount,
			"workStartTime":  user.WorkStartTime,
			"salary":         user.Salary,
			"operationCount": user.OperationCount,
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		logger.Error("更新数据错误")
		panic(err)
	}
}

type User struct {
	GroupCode      int64     // 群号
	Uin            int64     // QQ号
	LastLoginTime  time.Time // 最后登录时间
	Worth          float64   // 身价
	Count          int       // 可购买次数
	LastBuyTime    time.Time // 最后购买时间
	Money          float64   // 金币
	Master         int64     // MasterId
	Slaver         []int64   // 打工仔Id
	GuardTime      time.Time // 购买保护时间
	TormentCount   int       // 可折磨次数，根据Master
	WorkStartTime  time.Time // 开始工作时间
	Salary         float64   // 薪资
	OperationCount int
}

type SignTime struct {
	GroupCode  int64
	Uin        int64
	SignSeries int
	LastDay    string
}

type Points struct {
	GroupCode int64
	Uin       int64
	Point     int
}
