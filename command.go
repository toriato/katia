package main

import (
	"errors"
	"reflect"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Name        string
	Description string
	Options     interface{}
	Plugin      *Plugin

	OnExecute func(ctx CommandContext) interface{}
}

type CommandContext struct {
	Bot         *Bot
	Command     *Command
	Interaction *discordgo.Interaction
	Data        discordgo.ApplicationCommandInteractionData
	Options     interface{}
}

var (
	ErrCommandConflict = errors.New("명령어의 이름은 중복될 수 없습니다")
)

func (command *Command) parseOptions(fields interface{}) []*discordgo.ApplicationCommandOption {
	options := []*discordgo.ApplicationCommandOption{}

	values := reflect.ValueOf(fields)
	types := reflect.TypeOf(fields)

	for i := 0; i < values.NumField(); i++ {
		field := values.Field(i)
		fieldType := types.Field(i)
		tag := fieldType.Tag

		option := &discordgo.ApplicationCommandOption{
			Name:        tag.Get("name"),
			Description: tag.Get("desc"),
		}

		if tag.Get("required") == "true" {
			option.Required = true
		}

		switch field.Kind() {
		// TODO: ApplicationCommandOptionSubCommandGroup
		// TODO: ApplicationCommandOptionMentionable
		case reflect.Struct:
			option.Type = discordgo.ApplicationCommandOptionSubCommand
			option.Options = command.parseOptions(field.Interface())
		case reflect.String:
			option.Type = discordgo.ApplicationCommandOptionString
		case reflect.Int:
			option.Type = discordgo.ApplicationCommandOptionInteger
		case reflect.Bool:
			option.Type = discordgo.ApplicationCommandOptionBoolean
		case reflect.Ptr:
			switch fieldType.Type {
			case reflect.TypeOf(&discordgo.User{}):
				option.Type = discordgo.ApplicationCommandOptionUser
			case reflect.TypeOf(&discordgo.Channel{}):
				option.Type = discordgo.ApplicationCommandOptionChannel
			case reflect.TypeOf(&discordgo.Role{}):
				option.Type = discordgo.ApplicationCommandOptionRole
			default:
				continue
			}
		default:
			continue
		}

		options = append(options, option)
	}

	return options
}

func (ctx *CommandContext) formatOptions(fields interface{}, options []*discordgo.ApplicationCommandInteractionDataOption) interface{} {
	types := reflect.TypeOf(fields)
	if types.Kind() != reflect.Struct {
		return nil
	}

	values := reflect.New(types)

	for _, option := range options {
		var field reflect.Value

		for i := 0; i < types.NumField(); i++ {
			if types.Field(i).Tag.Get("name") == option.Name {
				field = values.Elem().Field(i)
				break
			}
		}

		var value interface{}

		switch option.Type {
		// TODO: ApplicationCommandOptionSubCommandGroup
		// TODO: ApplicationCommandOptionMentionable
		case discordgo.ApplicationCommandOptionSubCommand:
			value = ctx.formatOptions(field.Interface(), option.Options)
		case discordgo.ApplicationCommandOptionInteger,
			discordgo.ApplicationCommandOptionBoolean,
			discordgo.ApplicationCommandOptionString:
			value = option.Value
		case discordgo.ApplicationCommandOptionUser:
			value = option.UserValue(ctx.Bot.Session)
		case discordgo.ApplicationCommandOptionChannel:
			value = option.ChannelValue(ctx.Bot.Session)
		case discordgo.ApplicationCommandOptionRole:
			value = option.RoleValue(ctx.Bot.Session, ctx.Interaction.GuildID)
		}

		field.Set(reflect.ValueOf(value))
	}

	return values.Elem().Interface()
}
