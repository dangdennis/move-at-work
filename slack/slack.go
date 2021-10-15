package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var secrets struct {
	SlackBotToken      string
	SlackSigningSecret string
}

var workouts = map[string]string{
	"Monday":    "https://www.youtube.com/watch?v=hJFt4SuUFQE",
	"Tuesday":   "https://www.youtube.com/watch?v=DyXLiI8PsCQ",
	"Wednesday": "https://www.youtube.com/watch?v=CzZ8mxAeaKo",
	"Thursday":  "https://www.youtube.com/watch?v=8xX2Fq-DoB8",
	"Friday":    "https://www.youtube.com/watch?v=ynyCVCp5OMc",
}

var api = slack.New(secrets.SlackBotToken)
var signingSecret = secrets.SlackSigningSecret

//encore:api public raw path=/slack/events
func Bot(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sv, err := slack.NewSecretsVerifier(req.Header, signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := sv.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := sv.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	}

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			now := time.Now()
			msg := fmt.Sprintf("Today's routine. Practice every 90mins. %s", workouts[now.Weekday().String()])
			api.PostMessage(ev.Channel, slack.MsgOptionText(msg, false))
		}

		w.Header().Set("Content-Type", "text")
		w.Write([]byte(innerEvent.Type))
	}

}
