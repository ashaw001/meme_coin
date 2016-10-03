package handlers

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

func BetToPayout(bet int, payoutMultiplier float64) int {
	payout := int(math.Floor(float64(bet) * payoutMultiplier))
	return payout
}

func gambleProcess(content string, author *User, db *sqlx.DB) string {
	message := ""
	args := strings.Split(content, " ")
	if len(args) == 4 {
		bet, err := strconv.Atoi(args[1])
		if err != nil {
			message = "amount is too large or not a number, try again."
			return message
		}
		if bet <= 0 {
			message = "amount has to be more than 0"
			return message
		}
		game := args[2]
		gameInput := args[3]

		if bet > author.CurMoney {
			message = "not enough funds to complete transaction, total: " + strconv.Itoa(author.CurMoney) + " needed:" + strconv.Itoa(bet)
			return message
		}

		// Pick a number game
		if game == "number" {
			numberErrMessage := "!gamble <amount> number <numberToGuess>:<highestNumberInRange>. So !gamble 100 number 10:100 will run a pick a number game between 1 and 100 and the payout will be x100, because you have a 1  in 100 chance to win."
			gameInputs := strings.Split(gameInput, ":")

			if len(gameInputs) != 2 {
				return numberErrMessage
			}
			pickedNumber, err := strconv.Atoi(gameInputs[0])
			if err != nil || pickedNumber < 1 {
				return numberErrMessage
			}
			rangeNumber, err := strconv.Atoi(gameInputs[1])
			if err != nil || rangeNumber < pickedNumber {
				return numberErrMessage
			}
			if rangeNumber <= 1 {
				message = "your highestNumberInRange needs to be greater than 1"
				return message
			}

			answer := rand.Intn(rangeNumber) + 1
			message := "The result was " + strconv.Itoa(answer)
			if answer == pickedNumber {
				payout := BetToPayout(bet, float64(rangeNumber-1))
				MoneyAdd(author, payout, "gamble", db)
				message = message + ". Congrats, " + author.Username + " won " + strconv.Itoa(payout) + " memes."
				fmt.Println(message)
				return message
			} else {
				MoneyDeduct(author, bet, "gamble", db)
				message = message + ". Bummer, " + author.Username + " lost " + strconv.Itoa(bet) + " memes. :("
				fmt.Println(message)
				return message
			}
		}

		// Coin flip game
		if game == "coin" || game == "flip" {
			if gameInput == "heads" || gameInput == "tails" {
				answers := []string{"heads", "tails"}
				answer := answers[rand.Intn(len(answers))]
				message := "The result was " + answer

				if answer == gameInput {
					// 1x payout
					payout := BetToPayout(bet, 1.0)
					MoneyAdd(author, payout, "gamble", db)
					message = message + ". Congrats, " + author.Username + " won " + strconv.Itoa(payout) + " memes."
					fmt.Println(message)
					return message
				} else {
					MoneyDeduct(author, bet, "gamble", db)
					message = message + ". Bummer, " + author.Username + " lost " + strconv.Itoa(bet) + " memes. :("
					fmt.Println(message)
					return message
				}
			} else {
				message = "pick heads or tails bud. `!gamble <amount> coin heads|tails`"
				return message
			}
		}
	} else if args[0] == "!gamble" {
		message = `
			Gamble command is used as follows: '!gamble <amount> <game> <gameInput>
			 '!gamble <amount> coin|flip heads|tails' payout is 1x
			 '!gamble <amount> number <numberToGuess>:<highestNumberInRange>' payout is whatever the <highestNumberInRange> is.`
		return message
	}
	return message
}

func Gamble(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	author := UserGet(m.Author, db)
	message := gambleProcess(m.Content, &author, db)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
