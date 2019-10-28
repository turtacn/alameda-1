package metrics

import (
	"fmt"
	Keycodes "github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes"
	ApiEvents "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/events"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultKeycodeEnabled       = true
	DefaultKeycodeSpecs         = "0 0 * * * *"
	DefaultKeycodeEventInterval = "90,60,30,15,7,6,5,4,3,2,1,0,-1,-2,-3,-4,-5,-6,-7"
	DefaultKeycodeEventLevel    = "90:Info,15:Warn,0:Error"
)

type KeycodeMetrics struct {
	AlertMetrics

	eventLevel  map[int]ApiEvents.EventLevel
	eventPosted map[int]bool
	keycodeMgt  *Keycodes.KeycodeMgt
	days        int
	expired     bool
}

func NewKeycodeMetrics(notifier *Notifier) *KeycodeMetrics {
	keycode := KeycodeMetrics{}
	keycode.name = "keycode"
	keycode.notifier = notifier
	keycode.eventLevel = make(map[int]ApiEvents.EventLevel, 0)
	keycode.eventPosted = make(map[int]bool, 0)
	keycode.keycodeMgt = Keycodes.NewKeycodeMgt()
	keycode.days = 0
	keycode.expired = false
	keycode.GenerateCriteria()
	return &keycode
}

func (c *KeycodeMetrics) Validate() {
	if c.MeetCriteria() == true {
		if c.eventPosted[c.days] == false {
			c.eventPosted[c.days] = true
			c.PostEvent()
		}
	}
}

func (c *KeycodeMetrics) GenerateCriteria() {
	eventMap := map[int]ApiEvents.EventLevel{}
	for _, level := range strings.Split(c.notifier.EventLevel, ",") {
		day, _ := strconv.Atoi(strings.Split(level, ":")[0])
		value := strings.Split(level, ":")[1]

		switch value {
		case "Info":
			eventMap[day] = ApiEvents.EventLevel_EVENT_LEVEL_INFO
		case "Warn":
			eventMap[day] = ApiEvents.EventLevel_EVENT_LEVEL_WARNING
		case "Error":
			eventMap[day] = ApiEvents.EventLevel_EVENT_LEVEL_ERROR
		}
	}

	nowDay := 0
	for _, dayStr := range strings.Split(c.notifier.EventInterval, ",") {
		day, _ := strconv.Atoi(dayStr)
		if _, ok := eventMap[day]; ok {
			nowDay = day
		}
		c.eventLevel[day] = eventMap[nowDay]
		c.eventPosted[day] = false
	}
}

func (c *KeycodeMetrics) MeetCriteria() bool {
	currentTimestamp := time.Now().Unix()

	// Refresh keycode cache before getting keycode information
	c.keycodeMgt.Refresh(true)

	keycodes, summary, err := c.keycodeMgt.GetAllKeycodes()
	if err != nil {
		scope.Error("failed to check criteria when validate license")
		return false
	}

	// If no keycode is applied, we do not need to check the remaining days
	if len(keycodes) == 0 {
		c.days = 0
		c.expired = false
		c.clearEventPosted()
		return false
	}

	// Check if license is already expired
	isExpired := false
	if currentTimestamp >= summary.ExpireTimestamp {
		isExpired = true
	}

	// If expired status is changed from "true" to "false", we need to clear eventPosted map
	if c.expired != isExpired {
		c.expired = isExpired
		if isExpired == false {
			c.clearEventPosted()
		}
	}

	currentTime := time.Unix(currentTimestamp, 0)
	expireTime := time.Unix(summary.ExpireTimestamp, 0)

	diff := expireTime.Sub(currentTime)
	c.days = int(math.Floor(diff.Hours() / 24))

	if _, ok := c.eventLevel[c.days]; ok {
		return true
	}

	return false
}

func (c *KeycodeMetrics) PostEvent() error {
	eventLevel := c.eventLevel[c.days]

	days := 0
	if c.days >= 0 {
		days = c.days
	} else {
		days = 0
	}

	message := fmt.Sprintf("Expired in %d day(s).", days)

	err := Keycodes.PostEvent(eventLevel, message)
	if err != nil {
		scope.Error(err.Error())
		scope.Error("failed to post event when validate keycode")
		return err
	}

	return nil
}

func (c *KeycodeMetrics) clearEventPosted() {
	for k := range c.eventPosted {
		c.eventPosted[k] = false
	}
}
