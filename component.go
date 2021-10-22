package katia

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type Component struct {
	Name      string                                 `json:"name"`
	Plugin    *Plugin                                `json:"-"`
	OnExecute func(ctx ComponentContext) interface{} `json:"-"`
}

type ComponentContext struct {
	Bot         *Bot                                      `json:"-"`
	Component   *Component                                `json:"component"`
	Interaction *discordgo.Interaction                    `json:"interaction"`
	Logger      *logrus.Entry                             `json:"-"`
	Data        discordgo.MessageComponentInteractionData `json:"-"`
}

var (
	ErrComponentConflict = errors.New("명령어의 이름은 중복될 수 없습니다")
)
