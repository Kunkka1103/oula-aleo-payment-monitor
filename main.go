package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"
)

var (
	dsn         string
	//interDsn    string // 新增的中间检查数据库DSN
	robotURL    string
	checkTime   string
	interval    time.Duration
	isCompleted bool // 标记当天是否已完成打款
)

type DingTalkMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

func init() {
	flag.StringVar(&dsn, "dsn", "postgres://user:password@localhost/dbname?sslmode=disable", "PostgreSQL DSN")
	//flag.StringVar(&interDsn, "interDsn", "postgres://user:password@localhost/inter_db?sslmode=disable", "用于中间检查的PostgreSQL DSN")
	flag.StringVar(&robotURL, "robot", "", "钉钉机器人的URL (不需要@群成员)")
	flag.StringVar(&checkTime, "checkTime", "11:00", "每天开始监控的时间 (格式: HH:MM)")
	flag.DurationVar(&interval, "interval", 30*time.Minute, "检查的间隔时间")
}

func main() {
	flag.Parse()

	// 连接数据库
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("连接数据库失败：", err)
	}
	defer db.Close()

	// 连接中间检查数据库
	//interDb, err := sql.Open("postgres", interDsn)
	//if err != nil {
	//	log.Fatal("连接中间检查数据库失败：", err)
	//}
	//defer interDb.Close()

	log.Println("启动定时监控任务...")

	for {
		now := time.Now()
		nextCheckTime, err := getNextCheckTime(checkTime)
		if err != nil {
			log.Fatal("解析检查时间失败：", err)
		}

		// 如果当前时间在 checkTime 之前，等待到指定时间再开始
		if now.Before(nextCheckTime) {
			log.Printf("当前时间 %v，等待到达检查时间：%v", now, nextCheckTime)
			time.Sleep(time.Until(nextCheckTime))
		}

		// 初始化状态
		isCompleted = false
		log.Println("初始化状态，开始今日的监控任务")

		// 启动时立即执行一次检查
		checkAndAlert(db)

		// 启动定时器，按照 interval 定期检查
		ticker := time.NewTicker(interval)
		for range ticker.C {
			if isCompleted {
				log.Println("今日打款已完成，不再继续检查")
				ticker.Stop()
				break
			}
			checkAndAlert(db)
		}

		// 等待到第二天再开始检查
		waitUntilNextDay(nextCheckTime)
	}
}

// 获取当天的检查时间
func getNextCheckTime(checkTime string) (time.Time, error) {
	now := time.Now()
	checkTimeToday, err := time.ParseInLocation("15:04", checkTime, now.Location())
	if err != nil {
		return time.Time{}, err
	}

	// 将解析后的时间设置为今天的日期
	return time.Date(now.Year(), now.Month(), now.Day(), checkTimeToday.Hour(), checkTimeToday.Minute(), 0, 0, now.Location()), nil
}

// 等待直到第二天的 00:00
func waitUntilNextDay(nextCheckTime time.Time) {
	nextDay := time.Date(nextCheckTime.Year(), nextCheckTime.Month(), nextCheckTime.Day()+1, 0, 0, 0, 0, nextCheckTime.Location())
	log.Printf("等待到第二天 %v", nextDay)
	time.Sleep(time.Until(nextDay))
}

func checkAndAlert(db *sql.DB) {
	// 如果当天已完成打款，则不再继续检查
	if isCompleted {
		return
	}

	// 中间检查 - 查询 status = 0 的记录数
	//log.Println("中间检查：查询 status = 0 的记录数...")
	//var jobCount int
	//err := interDb.QueryRow(`SELECT count(*) FROM jobs WHERE status = 0`).Scan(&jobCount)
	//if err != nil {
	//	log.Println("查询 status = 0 的记录数时出错：", err)
	//	return
	//}
	//if jobCount != 0 {
	//	log.Printf("中间检查未通过，当前有 %d 个任务状态为 0，本次检查结束", jobCount)
	//	return
	//}

	// 中间检查 - 查询 confirmed_at
	//log.Println("中间检查：查询最近的 confirmed_at...")
	//var confirmedAt time.Time
	//err = interDb.QueryRow(`SELECT confirmed_at FROM jobs ORDER BY confirmed_at DESC NULLS LAST LIMIT 1`).Scan(&confirmedAt)
	//if err != nil {
	//	log.Println("查询 confirmed_at 时出错：", err)
	//	return
	//}
	//timeSinceConfirmed := time.Since(confirmedAt)
	//log.Printf("距离最近 confirmed_at 已过：%v", timeSinceConfirmed)
	//if timeSinceConfirmed < 5*time.Minute {
	//	log.Printf("最近 confirmed_at %v 在5分钟之内，跳过本次检查", confirmedAt)
	//	return
	//}

	// 原来的检查逻辑
	//log.Println("执行原始的打款检查逻辑...")

	// 加载上海时区
	//shanghaiLocation, err := time.LoadLocation("Asia/Shanghai")
	//if err != nil {
	//	log.Println("加载上海时区失败：", err)
	//	return
	//}

	// 查询最新的 payment_time
	//log.Println("查询最近的 payment_time...")
	//var paymentTime time.Time
	//err = db.QueryRow(`SELECT payment_time FROM token_payment_detail WHERE token_id='ALEO' ORDER BY payment_time DESC NULLS LAST LIMIT 1`).Scan(&paymentTime)
	//if err != nil {
	//	log.Println("查询 payment_time 时出错：", err)
	//	return
	//}

	// 将 payment_time 转换为上海时间仅用于日志打印
	//paymentTimeInShanghai := paymentTime.In(shanghaiLocation)
	//log.Printf("最近打款时间（上海时间）：%v", paymentTimeInShanghai)

	// 检查 payment_time 是否超过半小时
	//timeSincePayment := time.Since(paymentTime)
	//log.Printf("距离最近打款时间已过：%v", timeSincePayment)
	//if timeSincePayment < 30*time.Minute {
	//	log.Printf("最近打款时间%s 在30分钟之内，跳过本次检查", paymentTimeInShanghai)
	//	return
	//}

	// 查询未完成打款的记录数
	log.Println("查询未完成打款的记录...")
	var pendingCount int
	err := db.QueryRow(`
        SELECT count(*)
        FROM token_payment_detail
        WHERE payment_id IN (
            SELECT unnest(payment_ids)
            FROM settlement_daily
            WHERE "date" = CURRENT_DATE-1
            and token_id ='ALEO'
            and payment_status in ('initial','paying')
        ) AND status != 1;
    `).Scan(&pendingCount)
	if err != nil {
		log.Println("查询未完成打款记录时出错：", err)
		return
	}

	log.Printf("未完成打款记录数：%d", pendingCount)

	// 如果有未完成的打款，发送告警
	if pendingCount > 0 {
		log.Println("未完成打款，发送告警")
		sendAlert(fmt.Sprintf("ALEO 今日仍有未完成的打款，未完成打款记录数：%d", pendingCount))
	} else {
		// 如果已完成打款，标记为已完成，停止检查
		log.Println("所有打款已完成，停止检查")
		isCompleted = true
	}
}

func sendAlert(message string) {
	log.Printf("发送消息：%s", message)
	sendToRobot(robotURL, message)
}

func sendToRobot(url, message string) {
	msg := DingTalkMessage{
		MsgType: "text",
	}
	msg.Text.Content = message

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Println("序列化消息时出错：", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println("发送消息时出错：", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("钉钉响应非200状态：%s", resp.Status)
	} else {
		log.Println("消息发送成功")
	}
}
