package player

import (
	"fmt"
	"strings"
	"github.com/chrislunt/warwick/card"
	"os"
	"os/exec"
)

type Player struct {
	Hand *card.Hand
	Tableau *card.Tableau
	Strategy [][][]int // the inputs are the turn, the card kind, and the card cost
	Human bool
	State string // when playing with a human, this give you a place to store the current state to share with the player
}

// These represent the places a player could choose cards from
const NoCard = 0
const FromHand = 1
const FromStorage = 2
const FromStock = 3
const FromDiscard = 4

var legalStoreFrom = map[int] bool{
	FromHand: 	true,
	FromStorage: false,
	FromStock: 	true,
	FromDiscard: true,
}

var legalDiscardFrom = map[int] bool{
	FromHand: 	true,
	FromStorage: true,
	FromStock: 	false,
	FromDiscard: false,
}

var legalTrashFrom = map[int] bool{
	FromHand:	true,
	FromStorage: true,
	FromStock:	false,
	FromDiscard: false,
}

// This represents the place you can get a card from
type Pos struct {
	From int
	Index int
}

// this is a function type used to test is a card is valid for an action
type cardTest func(Pos, card.Card, Player) (bool, string)

// the cardTest when everything is cool
func everythingIsAwesome(Pos, card.Card, Player) (bool, string) {
	return true, ""
}


// This is for building
func (player Player) PlayerChooses(allowedFrom map[int] bool, phase int) (pos Pos, cost int, upgrade bool) {
	cost = 0
	upgrade = false
	if player.Human {
		choices := player.humanChooses("build", allowedFrom, nil, nil, cardIsBuildable, true, 1) // stock, discardPile, checkBuildable, passAllowed, selectCount
		pos = choices[0] // you can only choose 1 card to build at a time
		if pos.From == NoCard {
			return
		}
		thiscard := player.CardByPos(pos)
		if player.TopCard(thiscard.Kind) != nil && player.TopCard(thiscard.Kind).Cost == thiscard.Cost - 1 {
			upgrade = true
		} else {
			cost = thiscard.Cost + player.Tableau.Discounts[thiscard.Material]
		}
		return
	} 
	pos.From = NoCard // this means there's no legal build
	pos.Index = -1
	cost = -1
	upgrade = false

	for space := 1; space <= 4; space++ {
		if !allowedFrom[space] {
			continue
		}
		var cardrange []*card.Card
		if space == FromHand {
			cardrange = player.Hand.Cards
		} else if space == FromStorage {
			cardrange = player.Tableau.Storage
		}
		value := 0 // 0 to 63
		for id, thiscard := range cardrange {
			if thiscard == nil {
				continue;
			}
			isBuildable, _ := cardIsBuildable(Pos{space, id}, *thiscard, player)
			if (!isBuildable) {
				continue;
			}

			if (pos.From == NoCard) || (player.CardValue(thiscard, phase) > value) {
			// compare the value of this card to the current high

				// this is now our new high
				pos.From = space
				pos.Index = id
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
				// TODO: modify the value of the card based on the cost
			}
		}
	}
	return
}


func cardIsBuildable(pos Pos, thiscard card.Card, player Player) (buildable bool, reason string) {
	buildable = false // our assumption
	// You can't build a soldier card higher than your military card
	if thiscard.Kind == card.Soldiers {
		if (player.Tableau.Stack[card.Military] == nil) {
			reason = "No military power"
			return
		}
		if player.TopCard(card.Military).Cost < thiscard.Cost {
			reason = "Not enough military power"
			return
		}
	}

	// make sure the card isn't of a kind already built
	// or that if it is, it's only one higher than the current card
	if player.Tableau.Stack[thiscard.Kind] != nil {
		if player.TopCard(thiscard.Kind).Cost > thiscard.Cost - 1 {
			reason = "You already played a more powerful card of this kind"
			return
		}
		if player.TopCard(thiscard.Kind).Cost < thiscard.Cost - 1 {
			reason = "The card of this kind you already played is must be one less to upgrade"
			return
		}
	} else {
		// and that you can afford the card
		discountedCost := thiscard.Cost + player.Tableau.Discounts[thiscard.Material]
		// -1 to count because you must account for the card itself
		availableCards := player.Hand.Count - 1
		// Add in the cards in storage
		for _, thiscard := range player.Tableau.Storage {
			if thiscard != nil {
				availableCards++
			}
		}
		if discountedCost > availableCards {
			reason = "You can't afford it"
			return
		}
	}
	buildable = true
	reason = ""
	return
}


func (currentPlayer Player) clearScreen(state string) {
	cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    cmd.Run()
    fmt.Printf(state)
}


// Depending on the situation, the player may choose from his hand, cards in storage, the discard and the stock, or no card at all
// So the selection is represented as where the card is coming from, and the position.
// if you have to choose multiple cards, it will prevent you from choosing the same card twice
func (player Player) humanChooses(
	verb string,
	allowedFrom map[int] bool, 
	stock *card.Hand, 
	discardPile *card.Hand, 
	cardIsValid cardTest, 
	passAllowed bool,
	selectCount int) (positions []Pos) {

	choiceId := 0 // this is the number the human will key in to make their choice
	choice := make(map[int]Pos) // keep track of what each choice points to

	player.clearScreen(player.State)
	fmt.Println("-=* ", strings.ToUpper(verb), " *=-")
	if passAllowed {
		fmt.Println("0 . no", verb)
		choice[choiceId] = Pos{NoCard, 0}
	}

	choiceId++ // only pass is ever 0, so if no pass, we still move up by one

	location := "" // for telling people where they're playing from
	var cardrange []*card.Card
	for space := 1; space <= 4; space++ {
		if !allowedFrom[space] {
			continue
		}

		if space == FromDiscard {
			// only if there's a card available on the discard
			if discardPile.PullPos == 0 { // TODO: test this case
				continue
			}
			thiscard := discardPile.Cards[discardPile.PullPos]
			fmt.Printf("%d. DISCARD %s: %s\n", choiceId, thiscard, thiscard.Rule)
			choice[choiceId] = Pos{space, 0}
			continue

		} else if space == FromStock {
			fmt.Printf("%d. STOCK\n", choiceId)
			choice[choiceId] = Pos{space, 0}
			choiceId++
			continue // to to the next space

		} else if space == FromHand {
			cardrange = player.Hand.Cards
			location = ""
		} else if space == FromStorage {
			cardrange = player.Tableau.Storage
			location = "STORAGE "
		}

		for id, thiscard := range cardrange {
			if thiscard == nil {
				continue
			}
			isValid, reason := cardIsValid(Pos{space, id}, *thiscard, player)
			if (!isValid) {
				fmt.Printf("   %s%s (%s)\n", location, thiscard, reason)
				continue
			}
			fmt.Printf("%d. %s%s: %s\n", choiceId, location, thiscard, thiscard.Rule)
			choice[choiceId] = Pos{space, id}
			choiceId++
		}
	}

	// if they have more than one choice, offer them a redo option
	if selectCount > 1 {
		fmt.Printf("9. I messed up\nChoose %d\n", selectCount)
	}

	positions = make([]Pos, selectCount)

	if choiceId == 1 {
		// they don't really have a choice, just select 0: no card for them
		return
	}
	if choiceId == 2 && selectCount == 1 && !passAllowed {
		// there's only one choice, so just make it for them
		positions[0] = choice[1]
		return
	}
	// this outer "for" is to allow the user to restart their choice
	for ;; {
		i := 0
		tempChoice := make(map[int]Pos)
		// make a copy of the map, so we can remove elements as we go
		for k, v := range choice {
			tempChoice[k] = v
		}
		for ; i < selectCount; i++ {
			pos, input := queryPos(verb, tempChoice)
			if input == 9 {
				fmt.Printf("Start over selecting your cards\n")
				break; // if they get here, start over
			}
			positions[i] = pos
			if pos.From == NoCard {
				return
			}
			// remove that choice from the list so they can't select it again
			delete(tempChoice, input)
		}

		if i == selectCount {
			break; // if we got here they finished their selection
		}
	}
	return
}


func queryPos(verb string, choice map[int]Pos) (Pos, int) {
	// loop until they select a valid response
	for ;; {
		fmt.Printf("Choose a card to %s:\n", verb)
		var input int
		fmt.Scan(&input)

		if input == 9 {
			return Pos{}, 9
		}

		_, ok := choice[input] // check if the value given is in the choices
		if !ok {
			continue
		}
		return choice[input], input
	}
}


func (player Player) ChooseDiscards(protected Pos, cost int, phase int) (discards []Pos) {
	if player.Human {
		return player.HumanChooseDiscards(protected, cost)
	}
	// Consider if you'd rather use your stored cards.  You may value them differently, especially if you have
	// an upgrade for your storage that may refill the spot.  To not get too complicated, let's just compare
	// on the basis of the raw value, and if any of the stored cards are less than the card in the hand,
	// we'll remove those instead.

	discards = make([]Pos, cost) 
	// use cost to track your index position in the discards
	excludeList := make([][]bool, 3) // there are 3 spaces where this is valid: nothing, hand, and storage
	excludeList[FromHand] = make([]bool, player.Hand.Max)
	excludeList[FromStorage] = make([]bool, 2)
	excludeList[protected.From][protected.Index] = true
	for cost > 0 {
		pos, _ := player.LowestValueCard(phase, excludeList)
		excludeList[pos.From][pos.Index] = true
		cost--
		discards[cost] = pos
	}
	return
}


func (player Player) HumanChooseDiscards(protected Pos, cost int) (discards []Pos) {
	discards = make([]Pos, cost) 
	excludeProtected := func(pos Pos, thiscard card.Card, player Player) (bool, string) {
		if pos == protected {
			return false, "You can't discard this card"
		} else {
			return true, ""
		}
	}
	return player.humanChooses(
		"discard",
		legalDiscardFrom, 
		nil, // stock
		nil, // discardPile
		excludeProtected,
		false, // pass allowed
		cost, // selectCount
	)
}


func (currentPlayer Player) ChooseTrash(phase int) (trashPos []Pos) {
	if currentPlayer.Human {
		return currentPlayer.humanChooses(
			"trash",
			legalTrashFrom,
			nil, // stock
			nil, // discardPile
			everythingIsAwesome,
			true, // pass allowed
			1, // selectCount
		)

	}

	cardsTrashed := 0
	trashPos = make([]Pos, currentPlayer.Tableau.TrashBonus)
	for currentPlayer.Tableau.TrashBonus > cardsTrashed {
		// Strategy would be you won't discard a card over a certain value
		// and don't discard unless you get a draw, or you have too many cards
		oneTrash, value := currentPlayer.LowestValueCard(phase, nil)

		// if you're in the hand limit, don't trash if you have no draw bonus, 
		// or if your lowest value card is still valuable
		// outside of the hand limit, go ahead and trash
		if currentPlayer.Hand.Count <= currentPlayer.Hand.Limit { 
			if (currentPlayer.Tableau.DrawBonus == 0) || (value > 31) { // values are from 0-63
				oneTrash.From = NoCard
				trashPos[cardsTrashed] = oneTrash
				break
			}
		}
		cardsTrashed++
	}
	return
}


//TODO: attach player names to player object
func (currentPlayer Player) TrashCards(Poses []Pos, trash *card.Hand) (count int) {
	// if no cards to trash, or no cards of low enough value, get out
	for _, pos := range Poses {
		// if they chose none, just bail
		if pos.From == NoCard {
			break;
	    }
		fmt.Printf("Current player trashes %s\n", currentPlayer.CardByPos(pos))
		currentPlayer.Spend(pos, trash)
		count++
	}
	return
}



// TODO: I should be able to combine this routine with computerChooses, by passing in a "Playable function" 
// and a "compare function"
func (player Player) LowestValueCard(phase int, excludeList [][]bool) (pos Pos, value int) {
	pos.From = NoCard // this means there's no card available
	pos.Index = -1
	if excludeList == nil {
		excludeList = make([][]bool, 3) // there are 3 spaces where this is valid: nothing, hand, and storage
	}

	for space := 1; space <= 2; space++ {
		var cardrange []*card.Card
		if space == FromHand {
			cardrange = player.Hand.Cards
			if excludeList[space] == nil {
				excludeList[space] = make([]bool, player.Hand.Max)
			}
		} else if space == FromStorage {
			cardrange = player.Tableau.Storage
			if excludeList[space] == nil {
				excludeList[space] = make([]bool, 2) // there are a max of 2 storage in the current rules
			}
		}
		value := 64 // 0 to 63
		for id, thiscard := range cardrange {
			if (thiscard == nil) || excludeList[space][id] {
				continue
			}

			if (pos.From == NoCard) || (player.CardValue(thiscard, phase) < value) {
			// compare the value of this card to the current low

				// this is now our new low
				pos.From = space
				pos.Index = id
				value = player.CardValue(thiscard, phase)
				// TODO: modify the value of the card based on the cost
			}
		}
	}
	return
}


func (player Player) HighestValueCard(phase int, excludeList []bool) (highPos int, value int) {
	highPos = -1
	value = 0 // 0 to 63
	if excludeList == nil {
		excludeList = make([]bool, player.Hand.Max + 2) // the +2 are for "stored cards"
	}
	for id, thiscard := range player.Hand.Cards {
		if (thiscard == nil) || (excludeList[id]) {
			continue
		}
	// compare the value of this card to the current low
		if (highPos == -1) || (player.CardValue(thiscard, phase) > value) {
			// this is now our new high
			highPos = id
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


func (player *Player) Build(buildPos Pos, discards []Pos, discardPile *card.Hand, upgrade bool) {
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
	(*player).Spend(buildPos, nil)

	for _, discardPos := range discards {
		(*player).Spend(discardPos, discardPile)
	}
}


// TODO: this could be done better
func (player *Player) ChooseStore(stock *card.Hand, discardPile *card.Hand, phase int) (chosen *card.Card) {
	var pos Pos
	if (*player).Human {
		pos = (*player).humanChooseStore(stock, discardPile)
	} else {
		// if the best card in the discard or hand is less than 31, just draw from the stock
		discardValue := player.CardValue(discardPile.Cards[discardPile.PullPos], phase) 
		handPos, handValue := player.HighestValueCard(phase, nil)
		if (discardValue < 32) && (handValue < 32) {
			// draw from the stock
			pos = Pos{FromStock, 0}
			fmt.Sprintf("Player fills storage from Stock: %s", (*stock).Cards[(*stock).PullPos])
		} else if (discardValue < handValue) {
			// draw from the hand
			pos = Pos{FromHand, handPos}
			fmt.Sprintf("Player fills storage from Hand: %s", (*player).Hand.Cards[handPos])
		} else {
			pos = Pos{FromDiscard, 0}
			fmt.Sprintf("Player fills storage from Discard: %s", discardPile.Cards[discardPile.PullPos])
		}
	}

	if pos.From == FromStock {
		chosen = (*stock).Cards[(*stock).PullPos] // pull from the current pull position in the stock
		(*stock).PullPos--
//		return player.CardByPos(pos)
	} else if pos.From == FromDiscard {
		chosen = (*discardPile).Cards[(*discardPile).PullPos]
		(*discardPile).Cards[(*discardPile).PullPos] = nil
		(*discardPile).PullPos-- // a new card is the top of the discard
	} else if pos.From == FromHand {
		chosen = (*player).Hand.Cards[pos.Index]
		(*player).Hand.RemoveCard(pos.Index, nil)
	}
	return
}


// TODO: pick 2 if that's the option
// Choose from the hand, stock and discard pile, remove the card from the source, and pass it back
func (player *Player) humanChooseStore(stock *card.Hand, discardPile *card.Hand) (pos Pos) {
	fmt.Println("You may store a card.  Please choose:")
	choices := (*player).humanChooses("store", legalStoreFrom, stock, discardPile, everythingIsAwesome, false, 1)
	pos = choices[0] // you can only choose 1
	return
}


func (player *Player) Spend(pos Pos, discardPile *card.Hand) {
	if pos.From == FromStorage {
		(*player).Tableau.RemoveFromStorage(pos.Index, discardPile)
	} else if pos.From == FromHand {
		(*player).Hand.RemoveCard(pos.Index, discardPile)
	} else {
		panic("can only player.Spend from Storage or Hand")
	}
}


func (player Player) TopCard(kind int) (top *card.Card) {
	if player.Tableau.Stack[kind] == nil {
		return nil
	}
	return player.Tableau.Stack[kind].Cards[player.Tableau.Stack[kind].PullPos]
}


func (player Player) CardByPos(pos Pos) (returnCard *card.Card) {
	if pos.From == FromStorage {
   		returnCard = player.Tableau.Storage[pos.Index]
   	} else if pos.From == FromHand {
		returnCard = player.Hand.Cards[pos.Index]
   	} else {
   		panic("can only player.CardByPos from Hand or Storage")
   	}
   	return
}


func (currentPlayer Player) humanChooseAttack(opponent Player) (steal int) {
	steal = -1
	// if the opponent has a defensive building, you have to do that
	attackPower := currentPlayer.TopCard(card.Soldiers).Cost + currentPlayer.Tableau.AttackBonus
	if opponent.Tableau.Stack[card.Defensive] != nil {
		// make sure they can handle the defensive building
		if attackPower >= opponent.TopCard(card.Defensive).Cost {
			// you can take their defensive card
			for ;; { // loop until you get a valid response
				fmt.Printf("Would you like to use your soldier to take your opponent's %s (y/n)?\n", opponent.TopCard(card.Defensive).Name)
				var input string
				fmt.Scan(&input)
				if input == "y" {
					steal = card.Defensive
					return
				} else if input == "n" {
					return
				}
			}
		}
		return
	}
	found := false
	choice := make(map[int] int) // keep track of what each choice points to
	options := "--ATTACK--\n0. No attack\n"
	choice[0] = -1
	choiceId := 1 // this is the number the human will key in to make their choice
	for kind := 0; kind <= 9; kind++ {
		if opponent.Tableau.Stack[kind] != nil && attackPower >= opponent.TopCard(kind).Cost {
			found = true
			options += fmt.Sprintf("%d. %s\n", choiceId, opponent.TopCard(kind))
			choice[choiceId] = kind
			choiceId++
		}
	}
	
	if found {
		fmt.Printf(options)
		for ;; { // loop until you get a valid response
			fmt.Printf("Choose a card to take from your opponent:\n")
			var input int
			fmt.Scan(&input)
			_, ok := choice[input] // check if the value given is in the choices
			if ok {
				return choice[input]
			}
		}
	}
	return
}


func (currentPlayer Player) HumanWantsRedraw() (bool) {
	for ;; { // loop until you get a valid response
		fmt.Printf("Would you like to trash your hand and redraw %d cards (y/n)?:\n", currentPlayer.Hand.Count)
		var input string
		fmt.Scan(&input)
		if input == "y" {
			return true
		} else if input == "n" {
			return false
		}
	}
}


func (currentPlayer Player) ChooseAttack(opponent Player, phase int) (steal int) {
	steal = -1
	// for now, I'll just attack as soon as I can, but I will try to take the best card
	if currentPlayer.Tableau.Stack[card.Soldiers] == nil {
		return
	}

	if currentPlayer.Human {
		return currentPlayer.humanChooseAttack(opponent)
	}

	// if the opponent has a defensive building, you have to do that
	if opponent.Tableau.Stack[card.Defensive] != nil {
		// make sure they can handle the defensive building
		if (currentPlayer.TopCard(card.Soldiers).Cost + currentPlayer.Tableau.AttackBonus) >= opponent.TopCard(card.Defensive).Cost {
			// you can take their defensive card
			steal = card.Defensive
		}
		return
	}

	// Loop through the Tableau cards and find the best card to take (if you can take one)
	value := -1
	bestKind := -1
	for kind := 0; kind <= 9; kind++ {
		if opponent.Tableau.Stack[kind] != nil {
			if (currentPlayer.TopCard(card.Soldiers).Cost + currentPlayer.Tableau.AttackBonus) >= opponent.TopCard(kind).Cost {
				// note, it's how this player values the card, not the opponent
				if (value == -1) || (currentPlayer.CardValue(opponent.TopCard(kind), phase) > value) {
					value = currentPlayer.CardValue(opponent.TopCard(kind), phase)
					bestKind = kind
				}
			}
		}
	}
	if bestKind != -1 {
		steal = bestKind
	}
	return
}


func (currentPlayer *Player) Draw(discardPile *card.Hand, stock *card.Hand, phase int) {
    if (*currentPlayer).Tableau.DrawFromDiscardPower < 1 {
	    stock.RandomPull(2, (*currentPlayer).Hand) // this will only pull up to the hand limit
	    return
	}
	// here we should use "drawFromDiscardPower" when it's greater than one
	drawCount := (*currentPlayer).Hand.Max - (*currentPlayer).Hand.Count
	if drawCount == 0 {
		return
	}
	if drawCount > 2 { // you can't draw more than 2
		drawCount = 2
	}
	// loop through the draws you have
	for ; drawCount > 0; drawCount-- {
		if discardPile.PullPos == -1 { // the discard pile is empty, must pull from stock
        	stock.RandomPull(1, (*currentPlayer).Hand)
		} else if (*currentPlayer).Human {
			for ;; { // loop until you get a valid response
				fmt.Printf("Would you like to draw from the discard '%s' (y/n)?\n", discardPile.Cards[discardPile.PullPos])
				var input string
				fmt.Scan(&input)
				if input == "y" {
		   			discardPile.TopPull(1, (*currentPlayer).Hand)
		   			break;
				} else if input == "n" {
					// pull the remaining cards from the stock
		        	stock.RandomPull(drawCount, (*currentPlayer).Hand)
		        	return;
				}
			}

       	} else if (*currentPlayer).CardValue(discardPile.Cards[discardPile.PullPos], phase) > 31 {
   			fmt.Printf("Player draws %s from the discard\n", discardPile.Cards[discardPile.PullPos])
   			discardPile.TopPull(1, (*currentPlayer).Hand)
        } else {
        	stock.RandomPull(1, (*currentPlayer).Hand)
        }
    }

}


func (currentPlayer Player) VictoryPoints() (vp int) {
	vp = 0
	for kind := 0; kind <= 9; kind++ {
		if currentPlayer.Tableau.Stack[kind] != nil {
			vp += currentPlayer.TopCard(kind).VictoryPoints
		}
	}
	return
}

