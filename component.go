package main

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

type Component struct {
	Name      string
	Plugin    *Plugin
	OnExecute func(ctx ComponentContext) interface{}
}

type ComponentContext struct {
	Bot         *Bot
	Component   *Component
	Interaction *discordgo.Interaction
	Data        discordgo.MessageComponentInteractionData
}

var (
	ErrComponentConflict = errors.New("명령어의 이름은 중복될 수 없습니다")
)
