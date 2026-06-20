package remind

import (
	"encoding/csv"
	"errors"
	"log/slog"
	"os"
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
		t := time.Date(now.Year(), now.Month(), now.Day(), s.data.Hour(), s.data.Minute(), 0, 0, loc)
		return t.Before(s.lastTime)
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
		} else {
			s.nextTime = t
			break
		}
	}
}

func (s *remind) rember() {
	RemberChan <- Message{s.whereId, s.message}
	s.push()
}

var allReminds = []remind{}
var now time.Time
var loc *time.Location
var RemberChan chan Message

type Message struct {
	ChannelID string
	Text      string
}

func Init() (ok bool, missed []Message) {
	ok = false
	now = time.Now()
	loc = now.Location()
	f, err := os.Open("commands/remind/remind.csv")
	if errors.Is(err, os.ErrNotExist) {
		f, err := os.Create("commands/remind/remind.csv")
		if err != nil {
			slog.Error("Não foi possivel criar remind.csv", "error", err.Error())
			ok = false
			return
		}
		f.WriteString("when,where,data,message")
		f.Close()
		return
	} else if err != nil {
		slog.Error("Não foi possivel abrir remind.csv", "error", err.Error())
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
		switch line[0] {
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
			continue
		}
		remindAdd(line[1], r, data, miss, line[3])
	}
	if len(allReminds) == 0 {
		slog.Info("Não há nenhum lembrete.")
		return
	}

	ok = true
	RemberChan = make(chan Message)
	missed = remindIsLost()
	go loop()
	slog.Info("Todos os reminds serão lembrados.", "count", len(allReminds))

	return
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
	for idx := range allReminds {
		remind := &allReminds[idx]
		if now.After(remind.nextTime) {
			remind.rember()
		}
	}
}

func loop() {
	now = time.Now()
	nextRemindTime := allReminds[0].nextTime

	for _, remind := range allReminds {
		if nextRemindTime.After(remind.nextTime) {
			nextRemindTime = remind.nextTime
		}
	}

	for {
		time.Sleep(time.Until(nextRemindTime))
		now = time.Now()
		remindAll()
	}
}

func remindAdd(where string, when repetition, data time.Time, miss bool, message string) {
	var r remind
	r.whereId = where
	r.when = when
	r.remindIfMiss = miss
	r.data = data
	r.lastTime = now
	r.nextTime = r.next()
	r.message = message
	allReminds = append(allReminds, r)
}
