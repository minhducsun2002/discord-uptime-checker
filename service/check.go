package service

import (
	"discord-uptime-checker/structures"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	BotLabel     = "bot"
	ChannelLabel = "channel"
)

type CheckService struct {
	session  *discordgo.Session
	config   structures.Config
	registry *prometheus.Registry

	upGauge         *prometheus.GaugeVec
	latencyGauge    *prometheus.GaugeVec
	lastUpdateGauge *prometheus.GaugeVec

	lock sync.RWMutex

	responses  map[uint64]int
	responders map[uint64]byte

	started bool
}

func NewCheckService(session *discordgo.Session, config structures.Config, registry *prometheus.Registry) *CheckService {
	uptime := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bot_up",
			Help: "Whether bot is up",
		},
		[]string{BotLabel, ChannelLabel},
	)

	latency := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bot_latency",
			Help: "Time from the last request message to its respective response message in milliseconds",
		},
		[]string{BotLabel, ChannelLabel},
	)

	lastUpdate := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bot_last_update",
			Help: "Timestamp of last update",
		},
		[]string{BotLabel, ChannelLabel},
	)

	responders := make(map[uint64]byte)
	for _, target := range config {
		responders[target.Bot] = 1
	}

	return &CheckService{
		session:  session,
		config:   config,
		registry: registry,
		started:  false,

		lock: sync.RWMutex{},

		responses:  map[uint64]int{},
		responders: responders,

		upGauge:         uptime,
		latencyGauge:    latency,
		lastUpdateGauge: lastUpdate,
	}
}

func (c *CheckService) Start() {
	if c.started {
		return
	}

	c.registry.MustRegister(c.upGauge, c.latencyGauge, c.lastUpdateGauge)
	log.Printf("Checking %v target(s)", len(c.config))

	c.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		authorId, _ := strconv.ParseUint(m.Author.ID, 10, 64)
		if _, ok := c.responders[authorId]; !ok {
			return
		}

		if m.MessageReference == nil {
			return
		}

		refMessageId := m.MessageReference.MessageID
		c.lock.RLock()
		unlock := true

		refMessageId64, _ := strconv.ParseUint(refMessageId, 10, 64)

		if index, ok := c.responses[refMessageId64]; ok {
			{
				unlock = false
				c.lock.RUnlock()
			}

			c.lock.Lock()
			defer c.lock.Unlock()

			target := c.config[index]
			if target.Bot != authorId {
				return
			}

			delete(c.responses, refMessageId64)

			bot := strconv.FormatUint(target.Bot, 10)
			channel := strconv.FormatUint(target.Channel, 10)

			// log.Printf("Got success response from %v in %v (request message %v)", bot, channel, refMessageId64)

			labels := prometheus.Labels{BotLabel: bot, ChannelLabel: channel}

			t1, _ := discordgo.SnowflakeTimestamp(refMessageId)
			t2, _ := discordgo.SnowflakeTimestamp(m.ID)

			c.upGauge.With(labels).Set(1)
			c.latencyGauge.With(labels).Set(t2.Sub(t1).Seconds())
			c.lastUpdateGauge.With(labels).SetToCurrentTime()

			c.goClean(channel, refMessageId)
			c.goClean(channel, m.ID)
		}

		if unlock {
			c.lock.RUnlock()
		}
	})

	go c.run()
}

func (c *CheckService) run() {
	for idx, config := range c.config {
		go c.loop(&config, idx)
	}
}

func (c *CheckService) loop(target *structures.Target, index int) {
	bot := fmt.Sprintf("%d", target.Bot)
	content := fmt.Sprintf("<@%v> %s", bot, target.Keyword)
	channel := strconv.FormatUint(target.Channel, 10)

	timeout := time.Duration(target.Timeout) * time.Second
	labels := prometheus.Labels{BotLabel: bot, ChannelLabel: channel}

	for range time.NewTicker(7 * time.Second).C {
		message, err := c.session.ChannelMessageSend(channel, content)
		if err != nil {
			log.Printf("error checking for %v in channel %v: %v", bot, channel, err)
			continue
		}

		messageId, _ := strconv.ParseUint(message.ID, 10, 64)
		c.responses[messageId] = index

		go func() {
			time.Sleep(timeout)
			c.lock.Lock()
			defer c.lock.Unlock()

			if _, ok := c.responses[messageId]; ok {
				log.Printf("Timed out getting response from %v in %v after %v", bot, channel, timeout)

				delete(c.responses, messageId)

				c.upGauge.With(labels).Set(0)
				c.latencyGauge.With(labels).Set(-1)
				c.lastUpdateGauge.With(labels).SetToCurrentTime()

				c.goClean(channel, message.ID)
			}
		}()
	}
}

func (c *CheckService) goClean(channelId, messageId string) {
	_ = c.session.ChannelMessageDelete(channelId, messageId)
}
