package card

import (
	"fmt"
)

const Farm = 0
const Market = 1
const Storage = 2
const Supply = 3
const Manufacturing = 4
const Military = 5
const Defensive = 6
const Civic = 7
const School = 8
const Soldiers = 9
const Wood = 0
const Metal = 1
const Stone = 2
const Soldier = 3

var cardType = map[int]string{
	Farm: "Farm",
	Market: "Market",
	Storage: "Storage",
	Supply: "Supply",
	Manufacturing: "Manufacturing",
	Military: "Military",
	Defensive: "Defensive",
	Civic: "Civic",
	School: "School",
	Soldiers: "Soldiers",
}

var materials = map[int]string{
	Wood: "wood",
	Metal: "metal",
	Stone: "stone",
	Soldier: "soldier",
}

type Card struct {
	Name string
	Cost int
	Kind int
	Material int
	VictoryPoints int
	CostModifier []int
	BuildBonus int
	DrawFromDiscardPower int // we use an int instead of a bool so we can count multiple instances of the power
	TrashBonus int
	DrawBonus int
	AttackBonus int
	Rule string
}

func (c Card)String() string{
    return fmt.Sprintf("%s(%s %d : %s)", c.Name, cardType[c.Kind], c.Cost, materials[c.Material])
}


func (hand *Hand) RemoveCard(pos int, pile *Hand) {
	// move the card somewhere else if given
	if pile != nil {
		(*pile).Cards[(*pile).PullPos + 1] = (*hand).Cards[pos]
		(*pile).PullPos += 1
	}
	// remove the card from the hand
	(*hand).Cards[pos] = nil
	(*hand).Count--
}


func (tableau *Tableau) RemoveFromStorage(pos int, pile *Hand) {
	// move the card somewhere else if given
	if pile != nil {
		(*pile).Cards[(*pile).PullPos + 1] = (*tableau).Storage[pos]
		(*pile).PullPos += 1
	}
	// remove the card from the hand
	(*tableau).Storage[pos] = nil
}

/* 
 I'm storing a "hand", which is a collection of cards, by pointing to the canonical card
 in the deck.  This is an attempt to minimize the memory use which using a linked list would
 take.  If a card is removed from a hand, the array position is niled out, so looping over hands,
 you need to skip past nils.
*/
type Hand struct{
	Cards []*Card
	Count int // how many positions are currently filled
	Limit int // the size that above which you must discard
	Max int // the largest a hand could get in the middle of a hand
	PullPos int// for Hands that are a stock, this is the top of the pile
}


func (h Hand)String() string{
	var output string
	for _, card := range h.Cards {
		if card == nil {
			continue
		}
		output += fmt.Sprintf("%s ", *card)
	}
	return output
}


// dump the hand
func (from *Hand) Reset() {
	for	pos := 0; pos < (*from).Max; pos++ { 
		(*from).Cards[pos] = nil
		(*from).Count--
	}
}


//TODO: prevent more cards from being pulled than exist in the from hand, and signal the from hand is empty
// For a given hand "from", pull "pullCount" number of cards randomly, and add them to 
// the specified hand "receiving", but only in open positions.  If more card were called for, it skips them
func (from *Hand) RandomPull(pullCount int, receiving *Hand) {
	for	placePos := 0; placePos < (*receiving).Max; placePos++ { // the first spot in the receiving hand
		if (*receiving).Count >= (*receiving).Limit {
			break;
		}
		if ((*receiving).Cards[placePos] == nil) {
//fmt.Println("Stock pull", (*from).PullPos)
			if (*from).PullPos >= len((*from).Cards) {
				return
			}
			(*receiving).Cards[placePos] = (*from).Cards[(*from).PullPos] // pull from the current pull position in the stock
			(*from).PullPos++ // a new card is the top of the stock
			(*receiving).Count++
			pullCount--
			if (pullCount == 0) {
				return
			}
		}
	}
}


type Tableau struct {
	Stack map[int]*Hand
	Storage map[int] *Card
	Discounts []int
	Fill int // keep track of how filled the tableau is
	VictoryPoints int
	BuildBonus int
	DrawFromDiscardPower int
	TrashBonus int
	DrawBonus int
	AttackBonus int
}

func (t Tableau)String() string{
	var output string
	for i := 0; i < 10; i++ {
		if t.Stack[i] != nil {
			output += fmt.Sprintf("\n  %s:\t%s", cardType[i], t.Stack[i].Cards[t.Stack[i].PullPos])
		}
	}
	return output
}


// remove the top card from that tableau stack, adding to the given hand
func (tableau *Tableau) RemoveTop(kind int, hand *Hand) {
	top := (*tableau).Stack[kind].PullPos
	if hand != nil {
		for	placePos := 0; placePos < (*hand).Max; placePos++ { // the first spot in the receiving hand
			if ((*hand).Cards[placePos] == nil) {
				// this may push them over the hand limit
				// we'll deal with that later when we discard
				(*hand).Cards[placePos] = (*tableau).Stack[kind].Cards[top]
				(*hand).Count++
				break
			}
		}
	}
	// remove the card from the tableau
	(*tableau).BuildBonus -= (*tableau).Stack[kind].Cards[top].BuildBonus
	(*tableau).DrawBonus -= (*tableau).Stack[kind].Cards[top].DrawBonus
	(*tableau).TrashBonus -= (*tableau).Stack[kind].Cards[top].TrashBonus
	(*tableau).DrawFromDiscardPower -= (*tableau).Stack[kind].Cards[top].DrawFromDiscardPower
	(*tableau).AttackBonus -= (*tableau).Stack[kind].Cards[top].AttackBonus
	(*tableau).Stack[kind].Cards[top] = nil
	top--

	// see if there's a card underneath
	if (*tableau).Stack[kind].Cards[top] == nil {
		// no card underneath, we're losing a stack
		(*tableau).Stack[kind] = nil
		if kind != Soldiers {
			(*tableau).Fill-- 
		}
	} else {
		// there is a card underneath
		(*tableau).Stack[kind].PullPos = top
		(*tableau).BuildBonus += (*tableau).Stack[kind].Cards[top].BuildBonus
		(*tableau).DrawBonus += (*tableau).Stack[kind].Cards[top].DrawBonus
		(*tableau).TrashBonus += (*tableau).Stack[kind].Cards[top].TrashBonus
		(*tableau).DrawFromDiscardPower += (*tableau).Stack[kind].Cards[top].DrawFromDiscardPower
		(*tableau).AttackBonus += (*tableau).Stack[kind].Cards[top].AttackBonus
	}
}


var Deck = []Card{
	{"Fowlery", 1, Farm, Wood, 0, []int{0,0,0,-1}, 0, 0, 0, 0, 0, "-1 to recruit soldier"},
	{"Pig Farm", 2, Farm, Wood, 0, []int{0,0,0,-2}, 0, 0, 0, 0, 0, "-2 to recruit soldier"},
	{"Cow fields", 3, Farm, Wood, 0, []int{0,0,0,-3}, 0, 0, 0, 0, 0, "-3 to recruit soldier"},
	{"Manor", 4, Farm, Wood, 1, []int{0,0,0,-4}, 0, 0, 0, 0, 0, "-4 to recruit soldier; +1 VP"},

	{"Trading Post", 1, Market, Wood, 0, []int{0,0,0,0}, 0, 1, 0, 0, 0, "may draw from discard pile"},
	{"Bazaar", 2, Market, Wood, 0, []int{0,0,0,0}, 0, 1, 1, 0, 0, "may draw from discard pile; may trash 1"},
	{"Exchange", 3, Market, Wood, 0, []int{0,0,0,0}, 0, 1, 1, 1, 0, "may draw from discard pile; trash 1 to draw 1"},
	{"Faire", 4, Market, Wood, 1, []int{0,0,0,0}, 0, 1, 1, 2, 0, "may draw from discard pile; trash 1 to draw 2; +1 VP"},

	// Storage is 4 spaces.  Cards in storage may only be built, not discarded or trashed.  If storage card is raided, all storage goes with it
	{"Shed", 1, Storage, Wood, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "fill in storage space 1 from hand, draw or discard; may play that card but don't refill"},
	{"Warehouse", 2, Storage, Wood, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "refill storage spot 1 only if it's open"}, 
	{"Storehouse", 3, Storage, Wood, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "fill in storage space 2 from hand, draw or discard; may play that card but don't refill"},
	{"Vaults", 4, Storage, Wood, 1, []int{0,0,0,0}, 0, 0, 0, 0, 0, "refill any open storage; +1 VP"},

	{"Sawmill", 1, Supply, Metal, 0, []int{-1,0,0,0}, 0, 0, 0, 0, 0, "-1 to build Wood card"},
	{"Mine", 2, Supply, Metal, 0, []int{0,-1,0,0}, 0, 0, 0, 0, 0, "-1 to build metal card"},
	{"Quarry", 3, Supply, Metal, 0, []int{0,0,-1,0}, 0, 0, 0, 0, 0, "-1 to build stone card"},
	{"Gold stream", 4, Supply, Metal, 1, []int{-1,-1,-1,0}, 0, 0, 0, 0, 0, "-1 to build any card with a resource type, +1 VP"},

	{"Carpentery", 1, Manufacturing, Metal, 0, []int{-1,0,0,0}, 0, 0, 0, 0, 0, "-1 cost to build Wood card"},
	{"Blacksmith", 2, Manufacturing, Metal, 0, []int{0,-1,0,0}, 0, 0, 0, 0, 0, "-1 cost to build metal card"},
	{"Mason", 3, Manufacturing, Metal, 0, []int{0,0,-1,0}, 0, 0, 0, 0, 0, "-1 cost to build stone card"},
	{"Bank", 4, Manufacturing, Metal, 1, []int{-1,-1,-1,0}, 0, 0, 0, 0, 0, "-1 cost to build any card with a resource type; + 1 VP"},

	{"Armory", 1, Military, Metal, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Allows recruiting soldier up to level 1"},
	{"Garrison", 2, Military, Metal, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Allows recruiting soldier up to level 2"},
	{"Barrack", 3, Military, Metal, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Allows recruiting soldier up to level 3"},
	{"Fort", 4, Military, Metal, 1, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Allows recruiting soldier up to level 4; +1 VP"},

	{"Walls", 1, Defensive, Stone, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Protects all other buildings, may be taken by level 1 soldier"},
	{"Tower", 2, Defensive, Stone, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Protects all other buildings, may be taken by level 2 soldier"},
	{"Keep", 3, Defensive, Stone, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Protects all other buildings, may be taken by level 3 soldier"},
	{"Castle", 4, Defensive, Stone, 1, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Protects all other buildings, may be taken by level 4 soldier; +1 VP"},

	{"Chapel", 1, Civic, Stone, 1, []int{0,0,0,0}, 0, 0, 0, 0, 0, "+1 VP"},
	{"Church", 2, Civic, Stone, 2, []int{0,0,0,0}, 0, 0, 0, 0, 0, "+2 VP"},
	{"Town Hall", 3, Civic, Stone, 3, []int{0,0,0,0}, 0, 0, 0, 0, 0, "+3 VP"},
	{"Cathedral", 4, Civic, Stone, 4, []int{0,0,0,0}, 0, 0, 0, 0, 0, "+4 VP"},

	{"Novice", 1, School, Stone, 0, []int{0,0,0,0}, 1, 0, 0, 0, 0, "+1 build"},
	{"Adept", 2, School, Stone, 0, []int{0,0,0,0}, 1, 0, 0, 0, 1, "+1 build; +1 to Attack"},
	{"Mage", 3, School, Stone, 0, []int{0,0,0,0}, 2, 0, 0, 0, 1, "+2 builds; +1 to Attack"},
	{"Wizard", 4, School, Stone, 1, []int{0,0,0,0}, 2, 0, 0, 0, 2, "+2 builds; +2 to Attack; +1 VP"},
	
	{"Town Watch", 1, Soldiers, Soldier, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Only build if right military building built.  Optional: may take opponent card up to 1, may -1 opponent attack; trash after use"},
	{"Archers", 2, Soldiers, Soldier, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Only build if right military building built.  Optional: may take opponent card up to 2, may -2 opponent attack; trash after use"},
	{"Militia", 3, Soldiers, Soldier, 0, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Only build if right military building built.  Optional: may take opponent card up to 3, may -3 opponent attack; trash after use"},
	{"Knights", 4, Soldiers, Soldier, 1, []int{0,0,0,0}, 0, 0, 0, 0, 0, "Only build if right military building built.  Optional: may take opponent card up to 4, may -4 opponent attack; trash after use; +1 VP"},
}
