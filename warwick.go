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
 Storage reconsidered:  Level 1 store one card, Level 2: + may build the stored card, Level 3: + may spend the stored card, Level 4: +1 storage spot
 re2considered: 1: store card on table, 2: store 2nd card, 3: can move card back into hand, 4: fill any open storage spots at the time you build this card
 re3considered: with each level, you open a spot that must be immediately filled from the draw, discard or hand
 re4considered: 1: open a spot that may be immediately filled from the draw, discard or hand, 2: refill that spot, 3: add another, 4: refill both

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


func store(storePower int, stock *card.Hand, discardPile *card.Hand, player *player.Player, phase int) {
	var topSpot int
	switch {
	case storePower == 1 || storePower == 2:
		// the player may choose from hand, discard or stock to fill the storage
		// if the spot is open, you may refill it
		topSpot = 0
	case storePower == 3 || storePower == 4:
		// a second storage spot opens, fill from stock, discard or hand
		// for a 4, refill either open storage spots
		topSpot = 1
	}
	for spot := 0; spot <= topSpot; spot++ {
		if (*player).Tableau.Storage[spot] == nil {
			storeCard := (*player).ChooseStore(stock, discardPile, phase)
			log(1, fmt.Sprintf("Stored in storage %d: %s", spot, storeCard))
			(*player).Tableau.Storage[spot] = storeCard
		}
	}
}


func buildStock() (stock card.Hand, stockSize int) {
	rand.Seed( time.Now().UTC().UnixNano() )

	// double the deck.  This is the canonical reference of all cards.
	var	allCards = append(card.Deck[:], card.Deck[:]...)
	stockSize = len(allCards)

	// the stock, which can shrink, is a reference to all cards
	stock.Cards = make([]*card.Card, stockSize)
	stock.PullPos = stockSize - 1 // the position representing the current position to draw from

	/* There are two ways we could randomize, one would be randomize the stock and keep a pointer of where we currently are,
		which has an up-front randomization cost, but all subsequent pulls are cheap.  
	*/
	// TODO make this a parameter
	testStockId := -1
	var permutation []int
	if testStockId != -1 {
		/* rather than having to specify the whole deck, I allow you to only specify the top of the deck */
		fillSize := stockSize - len(card.TestStock[testStockId])
		fillOut := rand.Perm(fillSize)
		// for easier reading I specify the TestStock in reverse order, so get it ready to go on top
		s := card.TestStock[testStockId]
		for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
	        s[i], s[j] = s[j], s[i]
    	}
		permutation = append(fillOut[0:fillSize], card.TestStock[testStockId]...);
	} else {
		permutation = rand.Perm(stockSize)
	}
	for i, v := range permutation {
		stock.Cards[i] = &allCards[v]
	}
	return
}


func main() {
	stock, stockSize := buildStock()

	var discardPile card.Hand
	discardPile.Cards = make([]*card.Card, stockSize)
	discardPile.PullPos = -1

	var trash card.Hand
	trash.Cards = make([]*card.Card, stockSize)
	// trash is never pulled from, so no pull position

	players := make([]player.Player, 2);

	// set up rules about where you can get cards from for different actions
	legalBuildFrom := map[int] bool{
		player.FromHand: 	true,
		player.FromStorage: true,
		player.FromStock: 	false,
		player.FromDiscard: false,
	}

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
		// there are 10 types of cards, plus 2 storage spots so each slot must be initialized
		players[id].Tableau = &card.Tableau{}
		players[id].Tableau.Stack = make(map[int] *card.Hand)
		players[id].Tableau.Discounts = make([]int, 4)
		players[id].Tableau.BuildBonus = 0
		players[id].Tableau.AttackBonus = 0
		players[id].Tableau.Storage = make([] *card.Card, 2)
		players[id].Human = false

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
	// TODO: this should be an input parameter
	players[0].Human = true

	turnLimit := 0 // you can use this to cut a game short for dev purposes
	turnCount := 0
	gameOver := false
	// play until the deck runs out
	// or until the first player fills everything in their table (soldier doesn't matter)
	for (stock.PullPos > -1) && ((turnLimit == 0) || (turnCount < turnLimit)) && !gameOver {
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
		for id, currentPlayer := range players {
			if id == 0 {
				opponent = players[1]
			} else {
				opponent = players[0]
			}

			// if we're coming back to this player and they already have 9 cards, it's time to stop
			if currentPlayer.Tableau.Fill == 9 {
				gameOver = true
				break;
				// there is an error here in that if player 1 goes out first, player 0 doesn't get another play
			}
			// turn order:
			// 1. Build 
			// 2. Attack
			// 3. Trash (with Market)
			// 4. Draw up to 5 OR discard down to 5

			// determine card to build, cost
			// determine discards
			// do build
//			log(2, fmt.Sprintf("Player %d hand: %s", id, currentPlayer.Hand))
			log(2, fmt.Sprintf("Player %d Tableau: %s", id, currentPlayer.Tableau))

			builds := 0

			// we check it each time, since if you build the card, you get to use it immediately
			for builds < (currentPlayer.Tableau.BuildBonus + 1) {
				buildPos, cost, upgrade := currentPlayer.PlayerChooses(legalBuildFrom, phase)
			    var discards []player.Pos
			    if buildPos.From != player.NoCard {
			    	log(1, fmt.Sprintf("Player %d builds %s for %d", id, currentPlayer.CardByPos(buildPos), cost))
				    if cost > 0 {
			    		discards = currentPlayer.ChooseDiscards(buildPos, cost, phase)
			    		if logLevel > 1 {
							fmt.Println("Player", id, "discards:")
							for _, pos := range discards {
								fmt.Println(currentPlayer.CardByPos(pos))
							}	
						}
					}
					kind := currentPlayer.CardByPos(buildPos).Kind
				    cardValue := currentPlayer.CardByPos(buildPos).Cost
					currentPlayer.Build(buildPos, discards, &discardPile, upgrade)
					// if it's storage, you get a chance to place a card
					if kind == card.Storage {
						store(cardValue, &stock, &discardPile, &currentPlayer, phase);
					}

					log(2, fmt.Sprintf("currentPlayer %d has %d cards left", id, currentPlayer.Hand.Count))
					builds++

			    } else {
			    	break;
			    }
			}

	    	// When they don't build, and they have cards, check if they'd like to trash and redraw
			if builds == 0 {
		    	preResetCount := currentPlayer.Hand.Count 
		    	if (currentPlayer.Human && preResetCount > 0 && currentPlayer.HumanWantsRedraw()) || (currentPlayer.Hand.Count == currentPlayer.Hand.Limit) {
			   		// if the computer player can't build, but they have a full hand, they will get stuck.  Invoke the hand reset rule
			    	currentPlayer.Hand.Reset()
    				stock.RandomPull(preResetCount, players[id].Hand)
				   	fmt.Println("Player", id, "dumps their hand and redraws") 
				   	// if you recycle your hand, you don't get to do any builds, attacks, exchanges
    				continue;
				}
			}

			// ------ Attack --------- //
			steal := currentPlayer.ChooseAttack(opponent, phase) // steal is a card kind
			if steal != -1 {
				log(1, fmt.Sprintf("Player %d uses %s and takes opponent's %s", id, currentPlayer.TopCard(card.Soldiers), opponent.TopCard(steal)))
				opponent.Tableau.RemoveTop(steal, currentPlayer.Hand) 
				// then loose your attack card
				currentPlayer.Tableau.RemoveTop(card.Soldiers, &trash) // TODO: remove to trash, test if it works
			}

			// TODO: Human chooses cards to trash
			// ------- TRASH --------- //
			cardsTrashed := 0
			// TrashBonus measures the amount of cards you can trash in order to draw a new one
			if currentPlayer.Tableau.TrashBonus > 0 && currentPlayer.Hand.Count > 0 {
				trashPoses := currentPlayer.ChooseTrash(phase)
				cardsTrashed = currentPlayer.TrashCards(trashPoses, &trash)
			}
			// you must trash card to get the draw bonus under the current rules
			if (currentPlayer.Tableau.DrawBonus > 0) && (cardsTrashed > 0) {
				stock.RandomPull(currentPlayer.Tableau.DrawBonus, players[id].Hand)
			    log(1, fmt.Sprintf("Player %d bonus draws %d", id, currentPlayer.Tableau.DrawBonus))
			}

			// ------- DRAW --------- //
			// see how many open spots there are in the hand.  This may not run at all
			currentPlayer.Draw(&discardPile, &stock, phase)

		    // ------- DISCARD --------- //
		    // TODO: allow player to choose discard
		    for currentPlayer.Hand.Count > currentPlayer.Hand.Limit {
		    	fmt.Println("=================== Player", id, "has", currentPlayer.Hand.Count, "cards =================");
		    	trashPos, _ := currentPlayer.LowestValueCard(phase, nil)
		    	currentPlayer.Hand.RemoveCard(trashPos.Index, &trash)
		    }

		}
	    fmt.Println("----END OF TURN----")
	}

	// determine the winner
	if logLevel > 0 {
		vp := make([]int, 2)
		for id, currentPlayer := range players {
			 fmt.Println("Player", id, "Tableau:  ", currentPlayer.Tableau) 
			 vp[id] = currentPlayer.VictoryPoints()
		}

		fmt.Println("Player 0", vp[0], "-", vp[1], "Player 1")
		if vp[0] == vp[1] {
			fmt.Println("Tie game")
		} else if vp[0] < vp[1] {
			fmt.Println("Player 1 wins!")
		} else {
			fmt.Println("Player 0 wins!")				
		}
	}

}