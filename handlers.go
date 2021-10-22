package katia

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func handleReady(bot *Bot, e *discordgo.Ready) {
	if plugins, err := bot.ResolvePluginGraph(); err != nil {
		bot.Logger.WithField("plugins", plugins).Fatal(err)
	} else {
		for _, plugin := range plugins {
			if plugin.OnEnable != nil {
				if err := plugin.OnEnable(bot, plugin); err != nil {
					log.Fatal(err)
				}
			}

			plugin.Logger.Infof("plugin '%s' enabled", plugin.Name)
		}
	}
}

func handleInteractionCreate(bot *Bot, e *discordgo.InteractionCreate) {
	var result interface{}

	log := bot.Logger.WithField("interaction", e.Interaction)

	switch e.Type {
	case discordgo.InteractionApplicationCommand:
		i, r := handleInteractionCommand(bot, e)
		result = r

		log = log.WithFields(logrus.Fields{
			"type":     "command",
			"instance": i,
		})
	case discordgo.InteractionMessageComponent:
		i, r := handleInteractionComponent(bot, e)
		result = r
		log = log.WithFields(logrus.Fields{
			"type":     "component",
			"instance": i,
		})
	}

	if result == nil {
		log.Warnf("%s executed missing interaction", e.Member.User.ID)
		return
	}

	log.Infof("%s executed interaction", e.Member.User.ID)

	res := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{},
	}

	switch r := result.(type) {
	case Error:
		res.Data.Flags = 1 << 6
		res.Data.Embeds = []*discordgo.MessageEmbed{{
			Description: r.Error(),
			Color:       0xFF0000,
		}}

	case error:
		id := uuid.New().String()

		log.WithField("id", id).Errorf("%+v", r)
		res.Data.Flags = 1 << 6
		res.Data.Embeds = []*discordgo.MessageEmbed{{
			Description: "내부 오류가 발생했습니다",
			Footer: &discordgo.MessageEmbedFooter{
				Text: id,
			},
			Color: 0xFF0000,
		}}

	case bool:
		res.Type = discordgo.InteractionResponseDeferredMessageUpdate

	case string:
		res.Data.Content = r

	case *discordgo.MessageEmbed:
		res.Data.Embeds = []*discordgo.MessageEmbed{r}

	case []*discordgo.MessageEmbed:
		res.Data.Embeds = r

	case *discordgo.InteractionResponse:
		res = r

	case *discordgo.InteractionResponseData:
		res.Data = r

	default:
		bot.Logger.
			WithField("type", fmt.Sprintf("%v", r)).
			WithField("value", fmt.Sprintf("%#v", r)).
			Warnf("interaction returns unsupported type", r)
		res.Data.Content = `¯\_(ツ)_/¯`
	}

	bot.Session.InteractionRespond(e.Interaction, res)
}

func handleInteractionCommand(bot *Bot, e *discordgo.InteractionCreate) (interface{}, interface{}) {
	data := e.ApplicationCommandData()
	command := bot.Command(data.Name)
	if command == nil {
		return nil, nil
	}

	ctx := CommandContext{
		Bot:         bot,
		Command:     command,
		Data:        data,
		Interaction: e.Interaction,
	}

	ctx.Logger = command.Plugin.Logger.WithField("context", ctx)

	if command.Options != nil {
		ctx.Options = ctx.formatOptions(command.Options, data.Options)
	}

	return command, command.OnExecute(ctx)
}

func handleInteractionComponent(bot *Bot, e *discordgo.InteractionCreate) (interface{}, interface{}) {
	data := e.MessageComponentData()
	component := bot.Component(data.CustomID)
	if component == nil {
		return nil, nil
	}

	ctx := ComponentContext{
		Bot:         bot,
		Component:   component,
		Data:        data,
		Interaction: e.Interaction,
	}

	ctx.Logger = component.Plugin.Logger.WithField("context", ctx)

	return component, component.OnExecute(ctx)
}
