/* 
You are building a medieval village.  You win by having the most Victory Points (VP), which you get by building 
particular buildings.  You build by discarding cards.  The game is over when one player has built one
of each kind of building (and his opponent gets one more turn), or when the deck runs out.

Buildings:
There are 9 different kinds of building: Civic, Defensive, School, Military, Manufacturing, Supply, Market, Farm and Storage.
Each gives you additional powers.  Details are below.  There are 2 of each building in the deck.

Soldiers:
You can also recruit a Soldier card, but only up to the level of your current Military building.  Like a building, you must
discard as many cards as the level of the soldier, unless you have a farm, which reduces the cost of recruiting Soldiers.  
Soldiers can be used to attack the opponent.  By discarding a soldier, you may take a card of equal value from your opponent.  
If they have a defensive building, you must take the defensive building, and you may only take it if it is of equal or 
lesser value.  

Additionally, soldiers may be used for defense.  They subtract their value from the value of the attacking soldier.  If they
have a greater value, both attacking and defending soldiers are discarded.  This is optional, you may choose to save your 
soldier for attack.

Like buildings, soldiers may be upgraded, but not above the level of the military building.

Storage:
There are 4 spots on the board which can be used for storage, but only if you build the Storage building.  You may store as many
cards as the value of the storage building.  Cards put into Storage may only be built, they may not be discarded or trashed.  
When you build a storage building, the storage slots will be filled, but if you build one of those cards, you can use that 
spot to put a card from your hand.  If you upgrade your storage, you may be able to fill empty spots from the deck.  If the 
opponents soldier attacks you and takes your Storage building, he also gets everything you have in storage!

Storage reconsidered:
Level 1 store one card, Level 2: + may build the stored card, Level 3: + may spend the stored card, Level 4: +1 storage spot

Discard vs. Trash
There are two face up piles where cards go after they're used.  When building, cards go into the discard.  The Market buildings 
allow you to draw from the discard pile (you must draw from the top of the pile.  You may not look through the pile).  When 
using the Market buildings you may be able to trash cards to draw cards--cards that go into the trash may never be retrieved.
Soldiers go into the trash when they're used.  

Turn order:
1. Build or Upgrade
2. Attack (optional, if Soldier recruited)
3. Store (optional, if Storage built)
4. Trash (optional, if Market built)
5. Hand limit is 5, draw up to 2, or discard

- or -

Trash all cards in hand, and draw the number trashed.

Building and Upgrading:
To build a level 3 building, you need to discard 3 cards.  Buildings are built using wood, metal, or stone.  
There are buildings that make other buildings made of certain materials cheaper to build.  For instance, an Exchange (made of wood)
costs 3, but if you have a Carpentry, it will only cost you 2.  If you have a Carpentry and a Sawmill, it will only cost you 1!
Buildings can also be upgraded.  If you have a building of one level (say level 1), you can lay the building
of level 2 on top of it (but not the building of level 3 or 4!).  That counts as a build, but doesn't cost you 
anything.

To start: 
Each player gets 5 cards.


*/
package main

import (
	"fmt"
	"math/rand"
	"time"
	"github.com/chrislunt/warwick/card"
	"github.com/chrislunt/warwick/player"
)

var logLevel = 2

func log(level int, message string) {
	if logLevel >= level { 
		fmt.Println(message);
	}
}


func turnToPhase(turn int) (phase int) {
	if turn > 6 {
		phase = 2
	} else if turn > 3 {
		phase = 1 
	} else {
		phase = 0
	}
	return
}


func main() {
	rand.Seed( time.Now().UTC().UnixNano() )

	// double the deck.  This is the canonical reference of all cards.
	var	allCards = append(card.Deck[:], card.Deck[:]...)
	// the stock, which can shrink, is a reference to all cards
	var stock card.Hand
	stock.Cards = make([]*card.Card, len(allCards))
	stock.PullPos = 0 // the position representing the current position to draw from
	/* There are two ways we could randomize, one would be randomize the stock and keep a pointer of where we currently are,
		which has an up-front randomization cost, but all subsequent pulls are cheap.  
	*/
	permutation := rand.Perm(len(allCards))
	for i, v := range permutation {
		stock.Cards[v] = &allCards[i]
	}

	var discardPile card.Hand
	discardPile.Cards = make([]*card.Card, len(allCards))
	discardPile.PullPos = -1

	var trash card.Hand
	trash.Cards = make([]*card.Card, len(allCards))
	// trash is never pulled from, so no pull position

	players := make([]player.Player, 2);
	// initialize the players
	for id := range players {
		players[id].Hand = &card.Hand{}
		players[id].Hand.Limit = 5
		players[id].Hand.Max = 7
		// create the hand with an extra 2 slots beyond the limit, which could happen
		// if you use a soldier and then do an exchange
		players[id].Hand.Cards = make([]*card.Card, players[id].Hand.Max)
		// do the initial draw of 5 cards
		stock.RandomPull(5, players[id].Hand)
		// initize the Tableaus.  The Tableau is a map indexed by a card type constant
		// the map points to a small hand which is the potential stack of cards as someone upgrades
		// there are 10 types of cards, so each slot must be initialized
		players[id].Tableau = &card.Tableau{}
		players[id].Tableau.Stack = make(map[int] *card.Hand)
		players[id].Tableau.Discounts = make([]int, 4)
		players[id].Tableau.VictoryPoints = 0
		players[id].Tableau.BuildBonus = 0

		// the player strategy should be loaded from somewhere.  For now, set it all to 32
		// instead of 1 value per turn, do 3 columns for beginning, middle and end.
		// Value can be set by cost to start with.  Value may be adjusted by changes in cost.
		// value could be affected at time of spend by what may be discarded as well.
		players[id].Strategy = make([][][]int, 3)
		for phase := 0; phase <= 2; phase++ {
			players[id].Strategy[phase] = make([][]int, 10)
			for kind := 0; kind <= 9; kind++ {
				players[id].Strategy[phase][kind] = make([]int, 5)
				for cost := 1; cost <= 4; cost++ {
					players[id].Strategy[phase][kind][cost] = cost * 16 - 1
				}
			}
		}
	}


	turnLimit := 0 // you can use this to cut a game short for dev purposes
	turnCount := 0
	gameOver := false
	// play until the deck runs out
	// or until the first player fills everything in their table (soldier doesn't matter)
	for (stock.PullPos < len(allCards)) && ((turnLimit == 0) || (turnCount < turnLimit)) && !gameOver {
		turnCount++
		phase := turnToPhase(turnCount)

		// for safety
		// if you can't build any of the cards in your hand (because those positions are filled), you can get stuck
		if turnCount > 29 {
			fmt.Println("The game went to 30 turns--ending as a safety")
			gameOver = true
		}
		// turns
		var opponent player.Player
		for id, player := range players {
			if id == 0 {
				opponent = players[1]
			} else {
				opponent = players[0]
			}

			// if we're coming back to this player and they already have 9 cards, it's time to stop
			if player.Tableau.Fill == 9 {
				gameOver = true
				break;
				// there is an error here in that if player 1 goes out first, player 0 doesn't get another play
			}
			// turn order:
			// 1. Build 
			// 2. Attack
			// 3. Store (with Storage)
			// 4. Trash (with Market)
			// 3. Draw up to 5 OR discard down to 5

			// determine card to build, cost
			// determine discards
			// do build
			log(2, fmt.Sprintf("Player %d hand: %s", id, player.Hand))
			log(2, fmt.Sprintf("Player %d Tableau: %s", id, player.Tableau))

			builds := 0

			// we check it each time, since if you build the card, you get to use it immediately
			for builds < (player.Tableau.BuildBonus + 1) {
			    buildPos, cost, upgrade := player.ChooseBuild(phase)

			    var discards []int
			    if buildPos != -1 {
			    	log(1, fmt.Sprintf("Player %d builds %s for %d", id, player.Hand.Cards[buildPos], cost))
				    if cost > 0 {
			    		discards = player.ChooseDiscards(buildPos, cost, phase)
			    		if logLevel > 1 {
							fmt.Println("Player", id, "discards:")
							for _, pos := range discards {
								fmt.Println(player.Hand.Cards[pos])
							}	
						}
					}
					player.Build(buildPos, discards, &discardPile, upgrade)
					log(2, fmt.Sprintf("Player %d has %d cards left", id, player.Hand.Count))
					builds++

			    } else {
			    	break;
			    }
			}

			if (player.Hand.Count == player.Hand.Limit) && (builds == 0) {
			   	// if the player can't build, but they have a full hand, they will get stuck.  Invoke the hand reset rule
		    	player.Hand.Reset()
		    	stock.RandomPull(5, players[id].Hand)
			   	fmt.Println("Player", id, "dumps their hand and redraws") 
			   	// if you recycle your hand, you don't get to do any builds, attacks, exchanges
			   	continue;
			}

			// Attack
			// for now, I'll just attack as soon as I can, but I will try to take the best card
			if player.Tableau.Stack[card.Soldiers] != nil {
				soldierPos := player.Tableau.Stack[card.Soldiers].PullPos
				soldierCost := player.Tableau.Stack[card.Soldiers].Cards[soldierPos].Cost
				// if the opponent has a defensive building, you have to do that
				if opponent.Tableau.Stack[card.Defensive] != nil {
					defensivePos := opponent.Tableau.Stack[card.Defensive].PullPos
					// make sure they can handle the defensive building
					if soldierCost >= opponent.Tableau.Stack[card.Defensive].Cards[defensivePos].Cost {
						// you can take their defensive card
						log(1, fmt.Sprintf("Player %d uses %s and takes opponent's %s", id, player.Tableau.Stack[card.Soldiers].Cards[soldierPos],
							opponent.Tableau.Stack[card.Defensive].Cards[defensivePos]))
						opponent.Tableau.RemoveTop(card.Defensive, player.Hand)
						// then loose your attack card
						player.Tableau.RemoveTop(card.Soldiers, nil) // TODO: remove to trash
					}
				} else {
					// Loop through the Tableau cards and find the best card to take (if you can take one)
					value := -1
					bestKind := -1
					for kind := 0; kind <= 9; kind++ {
						if opponent.Tableau.Stack[kind] != nil {
							stack := opponent.Tableau.Stack[kind]
							kindPos := stack.PullPos
							if soldierCost >= stack.Cards[kindPos].Cost {
								// note, it's how this player values the card, not the opponent
								if (value == -1) || (player.CardValue(stack.Cards[kindPos], phase) > value) {
									value = player.CardValue(stack.Cards[kindPos], phase)
									bestKind = kind
								}
							}
						}
					}
					if bestKind != -1 {
						log(1, fmt.Sprintf("Player %d uses %s and takes opponent's %s", id, player.Tableau.Stack[card.Soldiers].Cards[soldierPos],
							opponent.Tableau.Stack[bestKind].Cards[opponent.Tableau.Stack[bestKind].PullPos]))
						opponent.Tableau.RemoveTop(bestKind, player.Hand)
						// then loose your attack card
						player.Tableau.RemoveTop(card.Soldiers, nil) // TODO: remove to trash
					}
				}
			}

			// ------- STORE --------- //


			// ------- TRASH --------- //
			cardsTrashed := 0
			for (cardsTrashed < player.Tableau.TrashBonus) && (player.Hand.Count > 0) {
				// Strategy would be you won't discard a card over a certain value
				// and don't discard unless you get a draw, or you have too many cards
				trashPos, value := player.LowestValueCard(phase, nil)

				// if you're in the hand limit, don't trash if you have no draw bonus, 
				// or if your lowest value card is still valuable
				// outside of the hand limit, go ahead and trash
				if player.Hand.Count <= player.Hand.Limit { 
					if (player.Tableau.DrawBonus == 0) || (value > 31) { // values are from 0-63
						trashPos = -1
					}
				}
				// if no cards to trash, or no cards of low enough value, get out
				if trashPos == -1 {
					break;
				}
			    log(1, fmt.Sprintf("Player %d trashes %s", id, player.Hand.Cards[trashPos]))
				player.Hand.RemoveCard(trashPos, &trash)
				cardsTrashed += 1
			}
			// you must trash card to get the draw bonus under the current rules
			if (player.Tableau.DrawBonus > 0) && (cardsTrashed > 0) {
				stock.RandomPull(player.Tableau.DrawBonus, players[id].Hand)
			    log(1, fmt.Sprintf("Player %d bonus draws %d", id, player.Tableau.DrawBonus))
			}

			// ------- DRAW --------- //
			// TODO: here we should use "drawFromDiscardPower" when it's greater than one
	        if (player.Tableau.DrawFromDiscardPower >= 1) && (discardPile.PullPos > -1) {
	        	if player.CardValue(discardPile.Cards[discardPile.PullPos], phase) > 31 {
	        		log(1, fmt.Sprintf("Player %d draws %s from the discard", id, discardPile.Cards[discardPile.PullPos]))
	        	}
	        }
		    stock.RandomPull(2, players[id].Hand) // this will only pull up to the hand limit
		    log(2, fmt.Sprintf("Player %d draws up to %d", id, player.Hand.Count))

		    // ------- DISCARD --------- //
		    for player.Hand.Count > player.Hand.Limit {
		    	fmt.Println("=================== Player", id, "has", player.Hand.Count, "cards =================");
		    	trashPos, _ := player.LowestValueCard(phase, nil)
		    	player.Hand.RemoveCard(trashPos, &trash)
		    }

		}
	    fmt.Println("----END OF TURN----")
	}

	// determine the winner
	if logLevel > 0 {
		for id, player := range players {
			 fmt.Println("Player", id, "Tableau:  ", player.Tableau) 
		}
		fmt.Println("Player 0", players[0].Tableau.VictoryPoints, "-", players[1].Tableau.VictoryPoints, "Player 1")
		if players[0].Tableau.VictoryPoints == players[1].Tableau.VictoryPoints {
			fmt.Println("Tie game")
		} else if players[0].Tableau.VictoryPoints < players[1].Tableau.VictoryPoints {
			fmt.Println("Player 1 wins!")
		} else {
			fmt.Println("Player 0 wins!")				
		}
	}

}