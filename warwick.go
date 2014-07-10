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
)

var logLevel = 1

func log(level int, message string) {
	if logLevel >= level { 
		fmt.Println(message);
	}
}

type Card struct {
	name string
	cost int
	kind int
	material int
	victoryPoints int
	costModifier []int
	buildBonus int
	drawFromDiscardPower int // we use an int instead of a bool so we can count multiple instances of the power
	trashBonus int
	drawBonus int
	rule string
}

func (c Card)String() string{
    return fmt.Sprintf("%s(%s %d : %s)", c.name, cardType[c.kind], c.cost, materials[c.material])
}

/* 
 I'm storing a "hand", which is a collection of cards, by pointing to the canonical card
 in the deck.  This is an attempt to minimize the memory use which using a linked list would
 take.  If a card is removed from a hand, the array position is niled out, so looping over hands,
 you need to skip past nils.
*/
type Hand struct{
	cards []*Card
	count int // how many positions are currently filled
	limit int // the size that above which you must discard
	max int // the largest a hand could get in the middle of a hand
	pullPos int// for Hands that are a stock, this is the top of the pile
}

func (h Hand)String() string{
	var output string
	for _, card := range h.cards {
		if card == nil {
			continue
		}
		if logLevel > 1 { output += fmt.Sprintf("%s ", *card) }
	}
	return output
}

type Tableau struct {
	stack map[int]*Hand
	discounts []int
	fill int // keep track of how filled the tableau is
	victoryPoints int
	buildBonus int
	drawFromDiscardPower int
	trashBonus int
	drawBonus int
}

func (t Tableau)String() string{
	var output string
	for i := 0; i < 10; i++ {
		if t.stack[i] != nil {
			output += fmt.Sprintf("\n  %s:\t%s", cardType[i], t.stack[i].cards[t.stack[i].pullPos])
		}
	}
	return output
}

type player struct {
	hand *Hand
	tableau *Tableau
	strategy [][][]int // the inputs are the turn, the card kind, and the card cost
}


//TODO: prevent more cards from being pulled than exist in the from hand, and signal the from hand is empty
// For a given hand "from", pull "pullCount" number of cards randomly, and add them to 
// the specified hand "receiving", but only in open positions.  If more card were called for, it skips them
func (from *Hand) randomPull(pullCount int, receiving *Hand) {
	for	placePos := 0; placePos < (*receiving).max; placePos++ { // the first spot in the receiving hand
		if (*receiving).count >= (*receiving).limit {
			break;
		}
		if ((*receiving).cards[placePos] == nil) {
//fmt.Println("Stock pull", (*from).pullPos)
			if (*from).pullPos >= len((*from).cards) {
				return
			}
			(*receiving).cards[placePos] = (*from).cards[(*from).pullPos] // pull from the current pull position in the stock
			(*from).pullPos++ // a new card is the top of the stock
			(*receiving).count++
			pullCount--
			if (pullCount == 0) {
				return
			}
		}
	}
}

// a return of (-1, -1) means no legal build
func (player player) chooseBuild(phase int) (buildPos int, cost int, upgrade bool) {
	buildPos = -1
	cost = -1
	upgrade = false

	// loop through the cards finding buildable cards
	buildable := make([]*Card, player.hand.max)
	for i, card := range player.hand.cards {
		if card == nil {
			continue
		}
		// You can't build a soldier card higher than your military card
		if card.kind == soldiers {
			if (player.tableau.stack[military] == nil) {
				continue
			}
			militaryPos := player.tableau.stack[military].pullPos
			if player.tableau.stack[military].cards[militaryPos].cost < card.cost {
				continue
			}
		}

		// make sure the card isn't of a kind already built
		// or that if it is, it's only one higher than the current card
		if player.tableau.stack[card.kind] != nil {
			pullPos := player.tableau.stack[card.kind].pullPos
			if player.tableau.stack[card.kind].cards[pullPos].cost != card.cost - 1 {
				continue
			}
		} else {
			// and that you can afford the card
			discountedCost := card.cost + player.tableau.discounts[card.material]
			// -1 to count because you must account for the card itself
			if discountedCost > (player.hand.count - 1) {
				continue
			}
		}

		buildable[i] = card
	}

	value := 0 // 0 to 63
	for id, card := range buildable {
		if card != nil {
			// compare the value of this card to the current high
			if (buildPos == -1) || (card.value(player, phase) > value) {
				// this is now our new high
				buildPos = id
				cost = card.cost + player.tableau.discounts[card.material]
				upgrade = false
				if player.tableau.stack[card.kind] != nil {
					pullPos := player.tableau.stack[card.kind].pullPos
					if player.tableau.stack[card.kind].cards[pullPos].cost == card.cost - 1 {
						upgrade = true
						cost = 0
					}
				}
				value = card.value(player, phase)
				// modify the value of the card based on the cost
			}
		}
	}

	return
}


func (player player) chooseDiscards(protected int, cost int, phase int) (discards []int) {
	discards = make([]int, cost) 
	// use cost to track your index position in the discards
	excludeList := make([]bool, player.hand.max)
	excludeList[protected] = true
	for cost > 0 {
		id, _ := player.lowestValueCard(phase, excludeList)
		excludeList[id] = true
		cost--
		discards[cost] = id
	}
	return
}


func (player player) lowestValueCard(phase int, excludeList []bool) (lowPos int, value int) {
	lowPos = -1
	value = 0 // 0 to 63
	if excludeList == nil {
		excludeList = make([]bool, player.hand.max)
	}
	for id, card := range player.hand.cards {
		if (card == nil) || (excludeList[id]) {
			continue
		}
		// compare the value of this card to the current low
		if (lowPos == -1) || (card.value(player, phase) < value) {
			// this is now our new low
			lowPos = id
			value = card.value(player, phase)
		}
	}
	return
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


func (card Card) value(player player, phase int) (value int) {
	// is the card already in the tableau, or less than a value in the tableau?
	// the hard part would be figuring the odds that a card may be taken by an 
	// opposing soldier
	modifier := 0
	if player.tableau.stack[card.kind] != nil {
		pullPos := player.tableau.stack[card.kind].pullPos
		posCost := player.tableau.stack[card.kind].cards[pullPos].cost 
		if posCost == card.cost {
			value = 1 // we'll make it one instead of zero, in case a soldier takes it
			return
		} else if posCost > card.cost {
			value = 0 // technically it may not be zero if a soldier takes it
			return
		} else if posCost < (card.cost - 1) {
			// in this case, our card isn't playable yet, but may be in the future
			modifier = -10
		}
	} 
	// if the card is still playable get the base value of the card, which depends on the player's strategy
	value = player.strategy[phase][card.kind][card.cost]
	value += modifier
	return
}


func (hand *Hand) removeCard(pos int, pile *Hand) {
	// move the card somewhere else if given
	if pile != nil {
		(*pile).cards[(*pile).pullPos + 1] = (*hand).cards[pos]
		(*pile).pullPos += 1
	}
	// remove the card from the hand
	(*hand).cards[pos] = nil
	(*hand).count--
}


// remove the top card from that tableau stack, adding to the given hand
func (tableau *Tableau) removeTop(kind int, hand *Hand) {
	top := (*tableau).stack[kind].pullPos
	if hand != nil {
		for	placePos := 0; placePos < (*hand).max; placePos++ { // the first spot in the receiving hand
			if ((*hand).cards[placePos] == nil) {
				// this may push them over the hand limit
				// we'll deal with that later when we discard
				(*hand).cards[placePos] = (*tableau).stack[kind].cards[top]
				(*hand).count++
				break
			}
		}
	}
	// remove the card from the tableau
	(*tableau).buildBonus -= (*tableau).stack[kind].cards[top].buildBonus
	(*tableau).drawBonus -= (*tableau).stack[kind].cards[top].drawBonus
	(*tableau).trashBonus -= (*tableau).stack[kind].cards[top].trashBonus
	(*tableau).drawFromDiscardPower -= (*tableau).stack[kind].cards[top].drawFromDiscardPower
	(*tableau).stack[kind].cards[top] = nil
	top--

	// see if there's a card underneath
	if (*tableau).stack[kind].cards[top] == nil {
		// no card underneath, we're losing a stack
		(*tableau).stack[kind] = nil
		if kind != soldiers {
			(*tableau).fill-- 
		}
	} else {
		// there is a card underneath
		(*tableau).stack[kind].pullPos = top
		(*tableau).buildBonus += (*tableau).stack[kind].cards[top].buildBonus
		(*tableau).drawBonus += (*tableau).stack[kind].cards[top].drawBonus
		(*tableau).trashBonus += (*tableau).stack[kind].cards[top].trashBonus
		(*tableau).drawFromDiscardPower += (*tableau).stack[kind].cards[top].drawFromDiscardPower
	}
}


// dump the hand
func (from *Hand) reset() {
	for	pos := 0; pos < (*from).max; pos++ { 
		(*from).cards[pos] = nil
		(*from).count--
	}
}


func (player *player) build(buildPos int, discards []int, discardPile *Hand, upgrade bool) {
	buildCard := (*player).hand.cards[buildPos]
	kind := buildCard.kind
	if (*player).tableau.stack[kind] == nil { // initialize
		(*player).tableau.stack[kind] = new(Hand)
		(*player).tableau.stack[kind].cards = make([]*Card, 5)
		(*player).tableau.stack[kind].pullPos = 0
	}
	// if we're upgrading an existing card, keep track of where it was
	var pullPos int
	if upgrade {
		pullPos = (*player).tableau.stack[kind].pullPos // this is the current card in power, if it's an upgrade
	}

	// the position in the tableau could have a card of different values, so put it in the spot for that value
	tableauPos := buildCard.cost
	(*player).tableau.stack[kind].cards[tableauPos] = buildCard
	(*player).tableau.stack[kind].pullPos = tableauPos

	// update the discount power the player has, and the victory points
	for i := 0; i < 4; i++ {
		if upgrade {
			// if they already have a power, subtract the old power before you add the new one
			(*player).tableau.discounts[i] -= (*player).tableau.stack[kind].cards[pullPos].costModifier[i]
		}
		(*player).tableau.discounts[i] += buildCard.costModifier[i]
	}

	// add victory points
	if upgrade {
		(*player).tableau.victoryPoints -= (*player).tableau.stack[kind].cards[pullPos].victoryPoints
		if (logLevel > 1) && ((*player).tableau.stack[kind].cards[pullPos].victoryPoints > 0) { 
			fmt.Println("Remove", (*player).tableau.stack[kind].cards[pullPos].victoryPoints, "victoryPoints")
		}
	}

	(*player).tableau.victoryPoints += buildCard.victoryPoints
	if (logLevel > 1) && (buildCard.victoryPoints > 0) { 
		fmt.Println("Add", buildCard.victoryPoints, "victoryPoints")
	}

	// if appropriate, cache the bonus info at the tableau level
	// NB the value is added or subtracted, because multiple cards could be in effect
	if upgrade {
		(*player).tableau.buildBonus -= (*player).tableau.stack[kind].cards[pullPos].buildBonus
		(*player).tableau.drawBonus -= (*player).tableau.stack[kind].cards[pullPos].drawBonus
		(*player).tableau.trashBonus -= (*player).tableau.stack[kind].cards[pullPos].trashBonus
		(*player).tableau.drawFromDiscardPower -= (*player).tableau.stack[kind].cards[pullPos].drawFromDiscardPower
	}
	(*player).tableau.buildBonus += buildCard.buildBonus
	(*player).tableau.drawBonus += buildCard.drawBonus
	(*player).tableau.trashBonus += buildCard.trashBonus
	(*player).tableau.drawFromDiscardPower += buildCard.drawFromDiscardPower

	// if it's not a soldier, and it's not an upgrade, add to the tableau fill count
	if (buildCard.kind != soldiers) && !upgrade {
		(*player).tableau.fill++
	}

	// remove the card from the hand
	(*player).hand.removeCard(buildPos, nil)

	for _, discardPos := range discards {
		(*player).hand.removeCard(discardPos, discardPile)
	}
}


const farm = 0
const market = 1
const storage = 2
const supply = 3
const manufacturing = 4
const military = 5
const defensive = 6
const civic = 7
const school = 8
const soldiers = 9
const wood = 0
const metal = 1
const stone = 2
const soldier = 3

var cardType = map[int]string{
	farm: "Farm",
	market: "Market",
	storage: "Storage",
	supply: "Supply",
	manufacturing: "Manufacturing",
	military: "Military",
	defensive: "Defensive",
	civic: "Civic",
	school: "School",
	soldiers: "Soldiers",
}

var materials = map[int]string{
	wood: "wood",
	metal: "metal",
	stone: "stone",
	soldier: "soldier",
}

var Deck = []Card{
	{"Fowlery", 1, farm, wood, 0, []int{0,0,0,-1}, 0, 0, 0, 0,"-1 to recruit soldier"},
	{"Pig farm", 2, farm, wood, 0, []int{0,0,0,-2}, 0, 0, 0, 0,"-2 to recruit soldier"},
	{"Cow fields", 3, farm, wood, 0, []int{0,0,0,-3}, 0, 0, 0, 0,"-3 to recruit soldier"},
	{"Manor", 4, farm, wood, 1, []int{0,0,0,-4}, 0, 0, 0, 0,"-4 to recruit soldier; +1 VP"},

	{"Trading Post", 1, market, wood, 0, []int{0,0,0,0}, 0, 1, 0, 0, "may draw from discard pile"},
	{"Bazaar", 2, market, wood, 0, []int{0,0,0,0}, 0, 1, 1, 0, "may draw from discard pile; may trash 1"},
	{"Exchange", 3, market, wood, 0, []int{0,0,0,0}, 0, 1, 1, 1, "may draw from discard pile; trash 1 to draw 1"},
	{"Faire", 4, market, wood, 1, []int{0,0,0,0}, 0, 1, 1, 2, "may draw from discard pile; trash 1 to draw 2; +1 VP"},

	// Storage is 4 spaces.  Cards in storage may only be built, not discarded or trashed.  If storage card is raided, all storage goes with it
	{"Shed", 1, storage, wood, 0, []int{0,0,0,0}, 0, 0, 0, 0, "fill in storage space 1; may put card in open storage spaces"},
	{"Warehouse", 2, storage, wood, 0, []int{0,0,0,0}, 0, 0, 0, 0, "fill in open storage spaces up to 2; may put card in open storage spaces"}, 
	{"Storehouse", 3, storage, wood, 0, []int{0,0,0,0}, 0, 0, 0, 0, "fill in open storage spaces up to 3; may put card in open storage spaces"},
	{"Vaults", 4, storage, wood, 1, []int{0,0,0,0}, 0, 0, 0, 0, "fill in open storage spaces up to 4; may put card in open storage spaces; +1 VP"},

	{"Sawmill", 1, supply, metal, 0, []int{-1,0,0,0}, 0, 0, 0, 0, "-1 to build wood card"},
	{"Mine", 2, supply, metal, 0, []int{0,-1,0,0}, 0, 0, 0, 0, "-1 to build metal card"},
	{"Quarry", 3, supply, metal, 0, []int{0,0,-1,0}, 0, 0, 0, 0, "-1 to build stone card"},
	{"Gold stream", 4, supply, metal, 1, []int{-1,-1,-1,0}, 0, 0, 0, 0, "-1 to build any card with a resource type, +1 VP"},

	{"Carpentery", 1, manufacturing, metal, 0, []int{-1,0,0,0}, 0, 0, 0, 0, "-1 cost to build wood card"},
	{"Blacksmith", 2, manufacturing, metal, 0, []int{0,-1,0,0}, 0, 0, 0, 0, "-1 cost to build metal card"},
	{"Mason", 3, manufacturing, metal, 0, []int{0,0,-1,0}, 0, 0, 0, 0, "-1 cost to build stone card"},
	{"Bank", 4, manufacturing, metal, 1, []int{-1,-1,-1,0}, 0, 0, 0, 0, "-1 cost to build any card with a resource type; + 1 VP"},

	{"Armory", 1, military, metal, 0, []int{0,0,0,0}, 0, 0, 0, 0, "Allows recruiting soldier up to level 1"},
	{"Garrison", 2, military, metal, 0, []int{0,0,0,0}, 0, 0, 0, 0, "Allows recruiting soldier up to level 2"},
	{"Barrack", 3, military, metal, 0, []int{0,0,0,0}, 0, 0, 0, 0, "Allows recruiting soldier up to level 3"},
	{"Fort", 4, military, metal, 1, []int{0,0,0,0}, 0, 0, 0, 0, "Allows recruiting soldier up to level 4; +1 VP"},

	{"Walls", 1, defensive, stone, 0, []int{0,0,0,0}, 0, 0, 0, 0, "Protects all other buildings, may be taken by level 1 soldier"},
	{"Tower", 2, defensive, stone, 0, []int{0,0,0,0}, 0, 0, 0, 0, "Protects all other buildings, may be taken by level 2 soldier"},
	{"Keep", 3, defensive, stone, 0, []int{0,0,0,0}, 0, 0, 0, 0, "Protects all other buildings, may be taken by level 3 soldier"},
	{"Castle", 4, defensive, stone, 1, []int{0,0,0,0}, 0, 0, 0, 0, "Protects all other buildings, may be taken by level 4 soldier; +1 VP"},

	{"Chapel", 1, civic, stone, 1, []int{0,0,0,0}, 0, 0, 0, 0, "+1 VP"},
	{"Church", 2, civic, stone, 2, []int{0,0,0,0}, 0, 0, 0, 0, "+2 VP"},
	{"Town Hall", 3, civic, stone, 3, []int{0,0,0,0}, 0, 0, 0, 0, "+3 VP"},
	{"Cathedral", 4, civic, stone, 4, []int{0,0,0,0}, 0, 0, 0, 0, "+4 VP"},

	{"Novice", 1, school, stone, 0, []int{0,0,0,0}, 1, 0, 0, 0, "+1 build"},
	{"Adept", 2, school, stone, 0, []int{0,0,0,0}, 1, 0, 0, 0, "+1 build; +1 to Attack"},
	{"Mage", 3, school, stone, 0, []int{0,0,0,0}, 2, 0, 0, 0, "+2 builds; +1 to Attack"},
	{"Wizard", 4, school, stone, 1, []int{0,0,0,0}, 2, 0, 0, 0, "+2 builds; +2 to Attack; +1 VP"},
	
	{"Town Watch", 1, soldiers, soldier, 0, []int{0,0,0,0}, 0, 0, 0, 0, "Only build if right military building built.  Optional: may take opponent card up to 1, may -1 opponent attack; trash after use"},
	{"Archers", 2, soldiers, soldier, 0, []int{0,0,0,0}, 0, 0, 0, 0, "Only build if right military building built.  Optional: may take opponent card up to 2, may -2 opponent attack; trash after use"},
	{"Militia", 3, soldiers, soldier, 0, []int{0,0,0,0}, 0, 0, 0, 0, "Only build if right military building built.  Optional: may take opponent card up to 3, may -3 opponent attack; trash after use"},
	{"Knights", 4, soldiers, soldier, 1, []int{0,0,0,0}, 0, 0, 0, 0, "Only build if right military building built.  Optional: may take opponent card up to 4, may -4 opponent attack; trash after use; +1 VP"},


}



func main() {
	rand.Seed( time.Now().UTC().UnixNano() )

	// double the deck.  This is the canonical reference of all cards.
	var	allCards = append(Deck[:], Deck[:]...)
	// the stock, which can shrink, is a reference to all cards
	var stock Hand
	stock.cards = make([]*Card, len(allCards))
	stock.pullPos = 0 // the position representing the current position to draw from
	/* There are two ways we could randomize, one would be randomize the stock and keep a pointer of where we currently are,
		which has an up-front randomization cost, but all subsequent pulls are cheap.  
	*/
	permutation := rand.Perm(len(allCards))
	for i, v := range permutation {
		stock.cards[v] = &allCards[i]
	}

	var discardPile Hand
	discardPile.cards = make([]*Card, len(allCards))
	discardPile.pullPos = -1

	var trash Hand
	trash.cards = make([]*Card, len(allCards))
	// trash is never pulled from, so no pull position

	players := make([]player, 2);
	// initialize the players
	for id := range players {
		players[id].hand = &Hand{}
		players[id].hand.limit = 5
		players[id].hand.max = 7
		// create the hand with an extra 2 slots beyond the limit, which could happen
		// if you use a soldier and then do an exchange
		players[id].hand.cards = make([]*Card, players[id].hand.max)
		// do the initial draw of 5 cards
		stock.randomPull(5, players[id].hand)
		// initize the tableaus.  The tableau is a map indexed by a card type constant
		// the map points to a small hand which is the potential stack of cards as someone upgrades
		// there are 10 types of cards, so each slot must be initialized
		players[id].tableau = &Tableau{}
		players[id].tableau.stack = make(map[int]*Hand)
		players[id].tableau.discounts = make([]int, 4)
		players[id].tableau.victoryPoints = 0
		players[id].tableau.buildBonus = 0

		// the player strategy should be loaded from somewhere.  For now, set it all to 32
		// instead of 1 value per turn, do 3 columns for beginning, middle and end.
		// Value can be set by cost to start with.  Value may be adjusted by changes in cost.
		// value could be affected at time of spend by what may be discarded as well.
		players[id].strategy = make([][][]int, 3)
		for phase := 0; phase <= 2; phase++ {
			players[id].strategy[phase] = make([][]int, 10)
			for kind := 0; kind <= 9; kind++ {
				players[id].strategy[phase][kind] = make([]int, 5)
				for cost := 1; cost <= 4; cost++ {
					players[id].strategy[phase][kind][cost] = cost * 16 - 1
				}
			}
		}
	}


	turnLimit := 0 // you can use this to cut a game short for dev purposes
	turnCount := 0
	gameOver := false
	// play until the deck runs out
	// or until the first player fills everything in their table (soldier doesn't matter)
	for (stock.pullPos < len(allCards)) && ((turnLimit == 0) || (turnCount < turnLimit)) && !gameOver {
		turnCount++
		phase := turnToPhase(turnCount)

		// for safety
		// if you can't build any of the cards in your hand (because those positions are filled), you can get stuck
		if turnCount > 29 {
			fmt.Println("The game went to 30 turns--ending as a safety")
			gameOver = true
		}
		// turns
		var opponent player
		for id, player := range players {
			if id == 0 {
				opponent = players[1]
			} else {
				opponent = players[0]
			}

			// if we're coming back to this player and they already have 9 cards, it's time to stop
			if player.tableau.fill == 9 {
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
			log(2, fmt.Sprintf("Player %d hand: %s", id, player.hand))
			log(2, fmt.Sprintf("Player %d tableau: %s", id, player.tableau))

			builds := 0

			// we check it each time, since if you build the card, you get to use it immediately
			for builds < (player.tableau.buildBonus + 1) {
			    buildPos, cost, upgrade := player.chooseBuild(phase)

			    var discards []int
			    if buildPos != -1 {
			    	log(1, fmt.Sprintf("Player %d builds %s for %d", id, player.hand.cards[buildPos], cost))
				    if cost > 0 {
			    		discards = player.chooseDiscards(buildPos, cost, phase)
			    		if logLevel > 1 {
							fmt.Println("Player", id, "discards:")
							for _, pos := range discards {
								fmt.Println(player.hand.cards[pos])
							}	
						}
					}
					player.build(buildPos, discards, &discardPile, upgrade)
					log(2, fmt.Sprintf("Player %d has %d cards left", id, player.hand.count))
					builds++

			    } else {
			    	break;
			    }
			}

			if (player.hand.count == player.hand.limit) && (builds == 0) {
			   	// if the player can't build, but they have a full hand, they will get stuck.  Invoke the hand reset rule
		    	player.hand.reset()
		    	stock.randomPull(5, players[id].hand)
			   	fmt.Println("Player", id, "dumps their hand and redraws") 
			   	// if you recycle your hand, you don't get to do any builds, attacks, exchanges
			   	continue;
			}

			// Attack
			// for now, I'll just attack as soon as I can, but I will try to take the best card
			if player.tableau.stack[soldiers] != nil {
				soldierPos := player.tableau.stack[soldiers].pullPos
				soldierCost := player.tableau.stack[soldiers].cards[soldierPos].cost
				// if the opponent has a defensive building, you have to do that
				if opponent.tableau.stack[defensive] != nil {
					defensivePos := opponent.tableau.stack[defensive].pullPos
					// make sure they can handle the defensive building
					if soldierCost >= opponent.tableau.stack[defensive].cards[defensivePos].cost {
						// you can take their defensive card
						log(1, fmt.Sprintf("Player %d uses %s and takes opponent's %s", id, player.tableau.stack[soldiers].cards[soldierPos],
							opponent.tableau.stack[defensive].cards[defensivePos]))
						opponent.tableau.removeTop(defensive, player.hand)
						// then loose your attack card
						player.tableau.removeTop(soldiers, nil) // TODO: remove to trash
					}
				} else {
					// Loop through the tableau cards and find the best card to take (if you can take one)
					value := -1
					bestKind := -1
					for kind := 0; kind <= 9; kind++ {
						if opponent.tableau.stack[kind] != nil {
							stack := opponent.tableau.stack[kind]
							kindPos := stack.pullPos
							if soldierCost >= stack.cards[kindPos].cost {
								// note, it's how this player values the card, not the opponent
								if (value == -1) || (stack.cards[kindPos].value(player, phase) > value) {
									value = stack.cards[kindPos].value(player, phase)
									bestKind = kind
								}
							}
						}
					}
					if bestKind != -1 {
						log(1, fmt.Sprintf("Player %d uses %s and takes opponent's %s", id, player.tableau.stack[soldiers].cards[soldierPos],
							opponent.tableau.stack[bestKind].cards[opponent.tableau.stack[bestKind].pullPos]))
						opponent.tableau.removeTop(bestKind, player.hand)
						// then loose your attack card
						player.tableau.removeTop(soldiers, nil) // TODO: remove to trash
					}
				}
			}

			// ------- STORE --------- //


			// ------- TRASH --------- //
			cardsTrashed := 0
			for (cardsTrashed < player.tableau.trashBonus) && (player.hand.count > 0) {
				// Strategy would be you won't discard a card over a certain value
				// and don't discard unless you get a draw, or you have too many cards
				trashPos, value := player.lowestValueCard(phase, nil)

				// if you're in the hand limit, don't trash if you have no draw bonus, 
				// or if your lowest value card is still valuable
				// outside of the hand limit, go ahead and trash
				if player.hand.count <= player.hand.limit { 
					if (player.tableau.drawBonus == 0) || (value > 31) { // values are from 0-63
						trashPos = -1
					}
				}
				// if no cards to trash, or no cards of low enough value, get out
				if trashPos == -1 {
					break;
				}
			    log(1, fmt.Sprintf("Player %d trashes %s", id, player.hand.cards[trashPos]))
				player.hand.removeCard(trashPos, &trash)
				cardsTrashed += 1
			}
			// you must trash card to get the draw bonus under the current rules
			if (player.tableau.drawBonus > 0) && (cardsTrashed > 0) {
				stock.randomPull(player.tableau.drawBonus, players[id].hand)
			    log(1, fmt.Sprintf("Player %d bonus draws %d", id, player.tableau.drawBonus))
			}

			// ------- DRAW --------- //
			// TODO: here we should use "drawFromDiscardPower" when it's greater than one
	        if (player.tableau.drawFromDiscardPower >= 1) && (discardPile.pullPos > -1) {
	        	if discardPile.cards[discardPile.pullPos].value(player, phase) > 31 {
	        		log(1, fmt.Sprintf("Player %d draws %s from the discard", id, discardPile.cards[discardPile.pullPos]))
	        	}
	        }
		    stock.randomPull(2, players[id].hand) // this will only pull up to the hand limit
		    log(2, fmt.Sprintf("Player %d draws up to %d", id, player.hand.count))

		    // ------- DISCARD --------- //
		    for player.hand.count > player.hand.limit {
		    	fmt.Println("=================== Player", id, "has", player.hand.count, "cards =================");
		    	trashPos, _ := player.lowestValueCard(phase, nil)
		    	player.hand.removeCard(trashPos, &trash)
		    }

		}
	    fmt.Println("----END OF TURN----")
	}

	// determine the winner
	if logLevel > 0 {
		for id, player := range players {
			 fmt.Println("Player", id, "tableau:  ", player.tableau) 
		}
		fmt.Println("Player 0", players[0].tableau.victoryPoints, "-", players[1].tableau.victoryPoints, "Player 1")
		if players[0].tableau.victoryPoints == players[1].tableau.victoryPoints {
			fmt.Println("Tie game")
		} else if players[0].tableau.victoryPoints < players[1].tableau.victoryPoints {
			fmt.Println("Player 1 wins!")
		} else {
			fmt.Println("Player 0 wins!")				
		}
	}

}