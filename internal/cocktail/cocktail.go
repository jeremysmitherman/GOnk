package cocktail

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gitlab.com/cantinadev/thecocktaildbclient/cocktail"
	"gitlab.com/cantinadev/thecocktaildbclient/fetcher"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func Handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	ctdb := fetcher.New(os.Getenv("TCDB_API_KEY"), &http.Client{})
	if strings.HasPrefix(m.Content, "!drank") {
		tokens := strings.Split(m.Content, " ")
		if tokens[1] == "with" {
			ingredients := tokens[2:len(tokens)]
			drinks, err := ctdb.SearchByIngredients(ingredients)
			if len(drinks) == 0 {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No drinks found with %s, <:kek:720702170563084288>", strings.Join(ingredients, " and ")))
			}

			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "🔥 Aww, you done broke it 🔥")
			}

			var fields []*discordgo.MessageEmbedField
			var curField *discordgo.MessageEmbedField = nil
			for i, d := range drinks {
				if i > 30 { break }
				if i % 5 == 0 {
					curField = &discordgo.MessageEmbedField{
						Name:   fmt.Sprintf("Dranks %d-%d", i + 1, i + 11),
						Inline: true,
					}
					fields = append(fields, curField)
				}
				curField.Value += fmt.Sprintf("%s\n", d.Name)
			}

			embed := &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("Drinks with %s", strings.Join(ingredients, " and ")),
				Description: "Type !drank <drink name> for details on a specific drink",
				Timestamp:   time.Now().Format(time.RFC3339),
				Color:       0x33ff33,
				Fields:      fields,
			}

			_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
			if err != nil {
				log.Fatalln(err)
			}
		} else if tokens[1] == "random" {
			drink, err := ctdb.GetRandom()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "🔥 Aww, you done broke it 🔥")
				return
			}

			s.ChannelMessageSendEmbed(m.ChannelID, getDrinkEmbed(drink))
		} else {
			search := tokens[1:len(tokens)]
			drinks, _ := ctdb.SearchByName(strings.Join(search, " "))
			s.ChannelMessageSendEmbed(m.ChannelID, getDrinkEmbed(drinks[0]))
		}
	}
}

func getDrinkEmbed(drink cocktail.Drink) *discordgo.MessageEmbed {
	ingredients := ""
	for _, i := range drink.Ingredients {
		ingredients += i.Amount + " " + i.Name + "\n"
	}

	return &discordgo.MessageEmbed{
		Title:       drink.Name,
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0x33ff33,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{
			URL:      drink.Image + "/preview",
			Width:    50,
			Height:   50,
		},
		Fields:      []*discordgo.MessageEmbedField{
			{
				Name:   "Ingredients",
				Value:  ingredients,
				Inline: true,
			},{
				Name:   "Instructions",
				Value:  drink.Instructions,
				Inline: false,
			},
		},
	}
}
