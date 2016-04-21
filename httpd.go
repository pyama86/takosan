package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"log"
	"math"
	"time"
)

type Httpd struct {
	Host string
	Port int
}

type Param struct {
	Channel    string   `form:"channel" binding:"required"`
	Message    string   `form:"message"`
	Name       string   `form:"name"`
	Icon       string   `form:"icon"`
	Fallback   string   `form:"fallback"`
	Color      string   `form:"color"`
	Pretext    string   `form:"pretext"`
	AuthorName string   `form:"author_name"`
	AuthorLink string   `form:"author_link"`
	AuthorIcon string   `form:"author_icon"`
	Title      string   `form:"title"`
	TitleLink  string   `form:"title_link"`
	Text       string   `form:"text"`
	FieldTitle []string `form:"field_title[]"`
	FieldValue []string `form:"field_value[]"`
	FieldShort []bool   `form:"field_short[]"`
	ImageURL   string   `form:"image_url"`
	Manual     bool     `form:"manual"`
	PostAt     string   `form:"post_at"`
}

func NewHttpd(host string, port int) *Httpd {
	return &Httpd{
		Host: host,
		Port: port,
	}
}

func (h *Httpd) Run() {
	m := martini.Classic()
	m.Get("/", func() string { return "Hello, I'm Takosan!!1" })
	m.Post("/notice", binding.Bind(Param{}), messageHandler)
	m.Post("/privmsg", binding.Bind(Param{}), messageHandler)
	m.RunOnAddr(fmt.Sprintf("%s:%d", h.Host, h.Port))
}

func messageHandler(p Param) (int, string) {
	ch := make(chan error, 1)

	// The format is compat with JavaScript's date.toISOString().
	// It expects the time is represented in UTC.
	dateFormat := "2006-01-02T15:04:05.000Z"
	postTime, err := time.Parse(dateFormat, p.PostAt)

	if err != nil {
		return sendSync(p, ch)
	} else {
		return sendAsync(p, postTime, ch)
	}
}

func sendSync(p Param, ch chan error) (int, string) {
	go MessageBus.Publish(NewMessage(p, ch), 0)
	err := <-ch

	if err != nil {
		message := fmt.Sprintf("Failed to send message to %s: %s\n", p.Channel, err)
		log.Printf(fmt.Sprintf("[error] %s", message))
		return 400, message
	} else {
		return 200, fmt.Sprintf("Message sent successfully to %s", p.Channel)
	}
}

func sendAsync(p Param, postTime time.Time, ch chan error) (int, string) {
	delay := int64(math.Max(float64(postTime.Unix()-time.Now().UTC().Unix()), 0))

	go MessageBus.Publish(NewMessage(p, ch), delay)

	return 200, fmt.Sprintf("Message enqueued and will be sent after %d seconds", delay)
}
