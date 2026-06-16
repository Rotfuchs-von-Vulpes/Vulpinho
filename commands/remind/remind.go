package remind

import (
	"encoding/csv"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type repetition int

const (
	everySecond repetition = iota
	everyMinute
	everyHour
	everyDay
	everyWeek
	everyMonth
	everyYear
	specific
)

type timeType int

const (
	t_instant timeType = iota
	t_minute
	t_hour
	t_day
	t_month
)

type Date struct {
}

type Time struct {
}

type remind struct {
	whereId       string
	when          repetition
	message       string
	initTimestamp int64
	lastTimestamp int64
	lastTime      time.Time
	nextTime      time.Time
	data          time.Time
	remindIfMiss  bool
	repeatAmmount uint
	repeatMax     uint
}

func (s *remind) next() time.Time {
	switch s.when {
	case specific:
		return s.data
	case everySecond:
		return s.lastTime.Add(time.Second)
	case everyMinute:
		return s.lastTime.Add(time.Minute)
	case everyHour:
		return s.lastTime.Add(time.Hour)
	case everyDay:
		t := time.Date(now.Year(), now.Month(), now.Day(), s.data.Hour(), s.data.Minute(), 0, 0, loc)
		if t.After(s.lastTime) {
			return t
		} else {
			return t.AddDate(0, 0, 1)
		}
	case everyMonth:
		t := time.Date(now.Year(), now.Month(), s.data.Day(), s.data.Hour(), s.data.Minute(), 0, 0, loc)
		if t.After(s.lastTime) {
			return t
		} else {
			return t.AddDate(0, 1, 0)
		}
	case everyYear:
		t := time.Date(now.Year(), s.data.Month(), s.data.Day(), s.data.Hour(), s.data.Minute(), 0, 0, loc)
		if t.After(s.lastTime) {
			return t
		} else {
			return t.AddDate(1, 0, 0)
		}
	}
	return time.Time{}
}

func (s *remind) miss() bool {
	switch s.when {
	case everyDay:
		t1 := time.Date(now.Year(), now.Month(), now.Day(), s.data.Hour(), s.data.Minute(), 0, 0, loc)
		return t1.After(s.lastTime) && t1.AddDate(0, 0, 1).Before(s.lastTime)
	case everyMonth:
		t1 := time.Date(now.Year(), now.Month(), s.data.Day(), s.data.Hour(), s.data.Minute(), 0, 0, loc)
		return t1.After(s.lastTime) && t1.AddDate(0, 0, 1).Before(s.lastTime)
	case everyYear:
		t1 := time.Date(now.Year(), s.data.Month(), s.data.Day(), s.data.Hour(), s.data.Minute(), 0, 0, loc)
		return t1.After(s.lastTime) && t1.AddDate(0, 0, 1).Before(s.lastTime)
	}
	return false
}

func (s *remind) push() {
	for {
		t := s.next()
		if now.After(t) {
			s.lastTime = t
			if s.repeatMax > 0 {
				s.repeatAmmount += 1
			}
		} else {
			s.nextTime = t
			break
		}
	}
}

var allReminds = []remind{}
var now time.Time
var loc *time.Location
var RemberChan chan Message

type Message struct {
	ChannelID string
	Message   string
}

func Init() (ok bool, missed []Message) {
	now = time.Now()
	loc = now.Location()
	f, err := os.Open("commands/remind/remind.csv")
	if err != nil {
		f, err := os.Create("commands/remind/remind.csv")
		if err != nil {
			slog.Error("Não foi possivel criar remind.csv", "error", err.Error())
			ok = false
			return
		}
		f.WriteString("when,where,data,ammount,max,message")
		f.Close()
		ok = false
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	data, err := r.ReadAll()
	for i, line := range data {
		if i == 0 {
			continue
		}
		var r repetition
		miss := true
		var data time.Time
		ammount, err := strconv.ParseUint(line[3], 10, 64)
		if err != nil {
			slog.Error("Cant read ammount field", "error", err)
			continue
		}
		max, err := strconv.ParseUint(line[4], 10, 64)
		if err != nil {
			slog.Error("Cant read max field", "error", err)
			continue
		}
		switch line[0] {
		case "everySecond":
			r = everySecond
			miss = false
			max = 10
		case "everyMinute":
			r = everyMinute
			miss = false
			max = 10
		case "everyHour":
			data, err = time.Parse("04", line[2])
			if err != nil {
				continue
			}
			r = everyHour
			miss = false
			max = 5
		case "everyDay":
			data, err = time.Parse("15:04", line[2])
			if err != nil {
				continue
			}
			r = everyDay
		case "everyMonth":
			data, err = time.Parse("02 15:04", line[2])
			if err != nil {
				continue
			}
			r = everyMonth
		case "everyYear":
			data, err = time.Parse("01 02 15:04", line[2])
			if err != nil {
				continue
			}
			r = everyYear
		default:
			return
		}
		remindAdd(line[1], r, data, uint(ammount), uint(max), miss, line[5])
	}
	ok = true
	RemberChan = make(chan Message)
	missed = remindIsLost()
	go loop()
	return
}

func removeRemind() {
	remove := func(idx int) {
		allReminds[idx] = allReminds[len(allReminds)-1]
		allReminds = allReminds[:len(allReminds)-1]
	}
	for idx, remind := range allReminds {
		if remind.repeatMax == 0 {
			continue
		}
		if remind.repeatAmmount >= remind.repeatMax {
			remove(idx)
		}
	}
}

func rember(r *remind) {
	RemberChan <- Message{r.whereId, r.message}
	r.push()
}

func remindIsLost() (final []Message) {
	for _, remind := range allReminds {
		if !remind.remindIfMiss {
			continue
		}
		if remind.miss() {
			final = append(final, Message{remind.whereId, remind.message})
		}
	}
	return
}

func remindAll() {
	for idx, remind := range allReminds {
		if now.After(remind.nextTime) {
			rember(&allReminds[idx])
		}
	}
}

func loop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		now = time.Now()
		remindAll()
		removeRemind()
	}
}

func remindAdd(where string, when repetition, data time.Time, ammount, max uint, miss bool, message string) {
	var r remind
	r.whereId = where
	r.when = when
	r.repeatAmmount = ammount
	r.repeatMax = max
	r.remindIfMiss = miss
	r.data = data
	r.lastTime = now
	r.nextTime = r.next()
	r.message = message
	allReminds = append(allReminds, r)
}
