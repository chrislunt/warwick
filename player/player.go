package player

import (
	"fmt"
	"github.com/chrislunt/warwick/card"
)

type Player struct {
	Hand *card.Hand
	Tableau *card.Tableau
	Strategy [][][]int // the inputs are the turn, the card kind, and the card cost
}


// a return of (-1, -1) means no legal build
func (player Player) ChooseBuild(phase int) (buildPos int, cost int, upgrade bool) {
	buildPos = -1
	cost = -1
	upgrade = false

	// loop through the cards finding buildable cards
	buildable := make([]*card.Card, player.Hand.Max + 2) // the plus 2 is for the potential 2 storage spaces
	searchArray := make([]*card.Card, player.Hand.Max + 2)
	copy(searchArray, player.Hand.Cards)

	// I debated about creating an abstraction of the additional searchable spaces, but decided to 
	// limit it to the storage card as it is now defined.  So I just add the Storage card on to the end of 
	// a search array
	if player.Tableau.Stack[card.Storage] != nil {
		storageMax := 1
		if player.TopCard(card.Storage).Cost >= 3 { // this means you have two slots
			storageMax = 2
		}
		for storagePos := 0; storagePos < storageMax; storagePos++ {
			if player.Tableau.Storage[storagePos] != nil {
				// add the card that's in storage into the search array
				searchArray[player.Hand.Max + storagePos] = player.Tableau.Storage[storagePos]
			}
		}
	}

	for i, thiscard := range searchArray {
		if thiscard == nil {
			continue
		}
		// You can't build a soldier card higher than your military card
		if thiscard.Kind == card.Soldiers {
			if (player.Tableau.Stack[card.Military] == nil) {
				continue
			}
			if player.TopCard(card.Military).Cost < thiscard.Cost {
				continue
			}
		}

		// make sure the card isn't of a kind already built
		// or that if it is, it's only one higher than the current card
		if player.Tableau.Stack[thiscard.Kind] != nil {
			if player.TopCard(thiscard.Kind).Cost != thiscard.Cost - 1 {
				continue
			}
		} else {
			// and that you can afford the card
			discountedCost := thiscard.Cost + player.Tableau.Discounts[thiscard.Material]
			// -1 to count because you must account for the card itself
			if discountedCost > (player.Hand.Count - 1) {
				continue
			}
		}

		buildable[i] = thiscard
	}

	value := 0 // 0 to 63
	for id, thiscard := range buildable {
		if thiscard != nil {
			// compare the value of this card to the current high
			if (buildPos == -1) || (player.CardValue(thiscard, phase) > value) {
				// this is now our new high
				buildPos = id
				cost = thiscard.Cost + player.Tableau.Discounts[thiscard.Material]
				upgrade = false
				if player.Tableau.Stack[thiscard.Kind] != nil {
					pullPos := player.Tableau.Stack[thiscard.Kind].PullPos
					if player.Tableau.Stack[thiscard.Kind].Cards[pullPos].Cost == thiscard.Cost - 1 {
						upgrade = true
						cost = 0
					}
				}
				value = player.CardValue(thiscard, phase)
				// modify the value of the card based on the cost
			}
		}
	}

	return
}


func (player Player) ChooseDiscards(protected int, cost int, phase int) (discards []int) {
// TODO: discard from storage: change the expression of discards to include the discards from storage as well
	// Consider if you'd rather use your stored cards.  You may value them differently, especially if you have
	// an upgrade for your storage that may refill the spot.  To not get too complicated, let's just compare
	// on the basis of the raw value, and if any of the stored cards are less than the card in the hand,
	// we'll remove those instead.

	discards = make([]int, cost) 
	// use cost to track your index position in the discards
	excludeList := make([]bool, player.Hand.Max + 2) // the +2 are for "stored cards"
	excludeList[protected] = true
	for cost > 0 {
		id, _ := player.LowestValueCard(phase, excludeList)
		excludeList[id] = true
		cost--
		discards[cost] = id
	}
	return
}


func (player Player) LowestValueCard(phase int, excludeList []bool) (lowPos int, value int) {
	lowPos = -1
	value = 0 // 0 to 63
	if excludeList == nil {
		excludeList = make([]bool, player.Hand.Max + 2) // the +2 are for "stored cards"
	}
	for id, thiscard := range player.Hand.Cards {
		if (thiscard == nil) || (excludeList[id]) {
			continue
		}
		// compare the value of this card to the current low
		if (lowPos == -1) || (player.CardValue(thiscard, phase) < value) {
			// this is now our new low
			lowPos = id
			value = player.CardValue(thiscard, phase)
		}
	}
	return
}


func (player Player) CardValue(thiscard *card.Card, phase int) (value int) {
	// is the card already in the Tableau, or less than a value in the Tableau?
	// the hard part would be figuring the odds that a card may be taken by an 
	// opposing soldier
	modifier := 0
	if player.Tableau.Stack[thiscard.Kind] != nil {
	    posCost := player.TopCard(thiscard.Kind).Cost
		if posCost == thiscard.Cost {
			value = 1 // we'll make it one instead of zero, in case a soldier takes it
			return
		} else if posCost > thiscard.Cost {
			value = 0 // technically it may not be zero if a soldier takes it
			return
		} else if posCost < (thiscard.Cost - 1) {
			// in this case, our card isn't playable yet, but may be in the future
			modifier = -10
		}
	} 
	// if the card is still playable get the base value of the card, which depends on the player's strategy
	value = player.Strategy[phase][thiscard.Kind][thiscard.Cost]
	value += modifier
	return
}


func (player *Player) Build(buildPos int, discards []int, discardPile *card.Hand, upgrade bool) {
	// TODO: discards needs to change to include the storage
	buildCard := (*player).CardByPos(buildPos)
	kind := buildCard.Kind
	if (*player).Tableau.Stack[kind] == nil { // initialize
		(*player).Tableau.Stack[kind] = new(card.Hand)
		(*player).Tableau.Stack[kind].Cards = make([]*card.Card, 5)
		(*player).Tableau.Stack[kind].PullPos = 0
	}
	// if we're upgrading an existing card, keep track of where it was
	var pullPos int
	if upgrade {
		pullPos = (*player).Tableau.Stack[kind].PullPos // this is the current card in power, if it's an upgrade
	}

	// the position in the Tableau could have a card of different values, so put it in the spot for that value
	TableauPos := buildCard.Cost
	(*player).Tableau.Stack[kind].Cards[TableauPos] = buildCard
	(*player).Tableau.Stack[kind].PullPos = TableauPos

	// update the discount power the player has, and the victory points
	for i := 0; i < 4; i++ {
		if upgrade {
			// if they already have a power, subtract the old power before you add the new one
			(*player).Tableau.Discounts[i] -= (*player).TopCard(kind).CostModifier[i] //(*player).Tableau.Stack[kind].Cards[pullPos].CostModifier[i]
		}
		(*player).Tableau.Discounts[i] += buildCard.CostModifier[i]
	}

	// add victory points
	if upgrade {
		(*player).Tableau.VictoryPoints -= (*player).Tableau.Stack[kind].Cards[pullPos].VictoryPoints
		if (*player).Tableau.Stack[kind].Cards[pullPos].VictoryPoints > 0 { 
			fmt.Println("Remove", (*player).Tableau.Stack[kind].Cards[pullPos].VictoryPoints, "victoryPoints")
		}
	}

	(*player).Tableau.VictoryPoints += buildCard.VictoryPoints
	if buildCard.VictoryPoints > 0 { 
		fmt.Println("Add", buildCard.VictoryPoints, "victoryPoints")
	}

	// if appropriate, cache the bonus info at the Tableau level
	// NB the value is added or subtracted, because multiple cards could be in effect
	if upgrade {
		(*player).Tableau.BuildBonus -= (*player).Tableau.Stack[kind].Cards[pullPos].BuildBonus
		(*player).Tableau.DrawBonus -= (*player).Tableau.Stack[kind].Cards[pullPos].DrawBonus
		(*player).Tableau.TrashBonus -= (*player).Tableau.Stack[kind].Cards[pullPos].TrashBonus
		(*player).Tableau.DrawFromDiscardPower -= (*player).Tableau.Stack[kind].Cards[pullPos].DrawFromDiscardPower
		(*player).Tableau.AttackBonus -= (*player).Tableau.Stack[kind].Cards[pullPos].AttackBonus
	}
	(*player).Tableau.BuildBonus += buildCard.BuildBonus
	(*player).Tableau.DrawBonus += buildCard.DrawBonus
	(*player).Tableau.TrashBonus += buildCard.TrashBonus
	(*player).Tableau.DrawFromDiscardPower += buildCard.DrawFromDiscardPower
	(*player).Tableau.AttackBonus += buildCard.AttackBonus

	// if it's not a soldier, and it's not an upgrade, add to the Tableau fill count
	if (buildCard.Kind != card.Soldiers) && !upgrade {
		(*player).Tableau.Fill++
	}

	// remove the built card
	if buildPos >= (*player).Hand.Max { // this is when moving the card from storage
		(*player).Tableau.RemoveFromStorage(buildPos - (*player).Hand.Max, nil)
	} else {
		(*player).Hand.RemoveCard(buildPos, nil)
	}

	for _, discardPos := range discards {
		(*player).Hand.RemoveCard(discardPos, discardPile)
	}
}


func (player Player) TopCard(kind int) (top *card.Card) {
	if player.Tableau.Stack[kind] == nil {
		return nil
	}
	return player.Tableau.Stack[kind].Cards[player.Tableau.Stack[kind].PullPos]
}


func (player Player) CardByPos(pos int) (returnCard *card.Card) {
   	if pos >= player.Hand.Max { // this is true when building from storage
   		returnCard = player.Tableau.Storage[pos - player.Hand.Max]
   	} else {
		returnCard = player.Hand.Cards[pos]
   	}
   	return
}


