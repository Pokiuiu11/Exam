package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type Card struct {
	Suit  string // ดอก: Spades, Hearts, Diamonds, Clubs
	Value string // เลขไพ่: 2-10, J, Q, K, A
}

func (c Card) String() string {
	return fmt.Sprintf("[%s %s]", c.Value, c.Suit)
}

var valueRank = map[string]int{
	"2": 2, "3": 3, "4": 4, "5": 5, "6": 6,
	"7": 7, "8": 8, "9": 9, "10": 10,
	"J": 11, "Q": 12, "K": 13, "A": 14,
}

func newDeck() []Card {
	suits := []string{"Spades", "Hearts", "Diamonds", "Clubs"}
	values := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

	var deck []Card
	for _, suit := range suits {
		for _, value := range values {
			deck = append(deck, Card{Suit: suit, Value: value})
		}
	}
	return deck
}

func shuffle(deck []Card) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

func deal(deck []Card) ([]Card, []Card) {
	hand := deck[:5]
	remaining := deck[5:]
	return hand, remaining
}

type HandRank int

const (
	HighCard      HandRank = 1
	OnePair       HandRank = 2
	TwoPair       HandRank = 3
	ThreeOfAKind  HandRank = 4
	Straight      HandRank = 5
	Flush         HandRank = 6
	FullHouse     HandRank = 7
	FourOfAKind   HandRank = 8
	StraightFlush HandRank = 9
	RoyalFlush    HandRank = 10
)

func rankName(r HandRank) string {
	names := map[HandRank]string{
		HighCard:      "High Card",
		OnePair:       "One Pair",
		TwoPair:       "Two Pair",
		ThreeOfAKind:  "Three of a Kind",
		Straight:      "Straight",
		Flush:         "Flush",
		FullHouse:     "Full House",
		FourOfAKind:   "Four of a Kind",
		StraightFlush: "Straight Flush",
		RoyalFlush:    "Royal Flush",
	}
	return names[r]
}

type HandResult struct {
	Rank    HandRank // ประเภทไพ่
	Kickers []int    // ค่าไพ่เรียงจากมากไปน้อย
}

func evaluateHand(hand []Card) HandResult {
	// เรียงลำดับไพ่ในมือจากแต้มมากไปน้อย
	sorted := make([]Card, len(hand))
	copy(sorted, hand)
	sort.Slice(sorted, func(i, j int) bool {
		return valueRank[sorted[i].Value] > valueRank[sorted[j].Value]
	})

	// นับจำนวนไพ่ที่แต้มซ้ำกัน
	counts := map[string]int{}
	for _, c := range sorted {
		counts[c.Value]++
	}

	// จัดชุดแต้มสำหรับใช้เปรียบเทียบกรณีเสมอ (Kickers)
	kickers := buildKickers(sorted, counts)

	// ตรวจดอกเหมือน (Flush) และ แต้มเรียงกัน (Straight)
	isFlush := checkFlush(sorted)
	isStraight, straightHigh := checkStraight(sorted)

	switch {
	case isFlush && isStraight && straightHigh == 14:
		return HandResult{RoyalFlush, kickers}
	case isFlush && isStraight:
		return HandResult{StraightFlush, kickers}
	case hasN(counts, 4):
		return HandResult{FourOfAKind, kickers}
	case hasN(counts, 3) && hasN(counts, 2):
		return HandResult{FullHouse, kickers}
	case isFlush:
		return HandResult{Flush, kickers}
	case isStraight:
		return HandResult{Straight, kickers}
	case hasN(counts, 3):
		return HandResult{ThreeOfAKind, kickers}
	case countPairs(counts) == 2:
		return HandResult{TwoPair, kickers}
	case countPairs(counts) == 1:
		return HandResult{OnePair, kickers}
	default:
		return HandResult{HighCard, kickers}
	}
}

// buildKickers เรียงค่าไพ่โดยให้ค่าที่มีจำนวนมากกว่าอยู่หน้า
// เช่น Full House KKK AA -> [13,13,13,14,14]
func buildKickers(sorted []Card, counts map[string]int) []int {
	type valCount struct {
		val   int
		count int
	}
	seen := map[int]bool{}
	var groups []valCount
	for _, c := range sorted {
		v := valueRank[c.Value]
		if !seen[v] {
			seen[v] = true
			groups = append(groups, valCount{v, counts[c.Value]})
		}
	}

	sort.Slice(groups, func(i, j int) bool {
		if groups[i].count != groups[j].count {
			return groups[i].count > groups[j].count
		}
		return groups[i].val > groups[j].val
	})
	var kickers []int
	for _, g := range groups {
		for k := 0; k < g.count; k++ {
			kickers = append(kickers, g.val)
		}
	}
	return kickers
}

// checkFlush ตรวจว่าไพ่ทั้ง 5 ใบเป็นดอกเดียวกันไหม
func checkFlush(sorted []Card) bool {
	suit := sorted[0].Suit
	for _, c := range sorted {
		if c.Suit != suit {
			return false
		}
	}
	return true
}

// checkStraight ตรวจว่าไพ่เรียงลำดับติดกันไหม
// คืนค่า: (เป็น straight?, ค่าสูงสุด)
func checkStraight(sorted []Card) (bool, int) {
	vals := []int{}
	for _, c := range sorted {
		vals = append(vals, valueRank[c.Value])
	}
	sort.Sort(sort.Reverse(sort.IntSlice(vals)))

	normal := true
	for i := 1; i < len(vals); i++ {
		if vals[i-1]-vals[i] != 1 {
			normal = false
			break
		}
	}
	if normal {
		return true, vals[0]
	}

	if vals[0] == 14 {
		wheel := []int{14, 5, 4, 3, 2}
		isWheel := true
		for i := range vals {
			if vals[i] != wheel[i] {
				isWheel = false
				break
			}
		}
		if isWheel {
			return true, 5
		}
	}
	return false, 0
}

// hasN ตรวจว่ามีไพ่ที่ซ้ำกัน n ใบไหม
func hasN(counts map[string]int, n int) bool {
	for _, c := range counts {
		if c == n {
			return true
		}
	}
	return false
}

// countPairs นับจำนวน pair ในมือ
func countPairs(counts map[string]int) int {
	pairs := 0
	for _, c := range counts {
		if c == 2 {
			pairs++
		}
	}
	return pairs
}

// ==================== เปรียบเทียบผู้ชนะ ====================

// compareHands เปรียบเทียบสองมือ
// คืนค่า: 1 = มือแรกชนะ, -1 = มือสองชนะ, 0 = เสมอ
func compareHands(a, b HandResult) int {
	if a.Rank != b.Rank {
		if a.Rank > b.Rank {
			return 1
		}
		return -1
	}

	for i := 0; i < len(a.Kickers) && i < len(b.Kickers); i++ {
		if a.Kickers[i] > b.Kickers[i] {
			return 1
		} else if a.Kickers[i] < b.Kickers[i] {
			return -1
		}
	}
	return 0
}

// ==================== Main ====================

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// 1. สร้างและสับไพ่
	deck := newDeck()
	shuffle(deck)

	// 2. แจกไพ่ให้ผู้เล่น 4 คน คนละ 5 ใบ
	numPlayers := 4
	hands := make([][]Card, numPlayers)
	for i := 0; i < numPlayers; i++ {
		hands[i], deck = deal(deck)
	}

	// 3. ประเมินไพ่แต่ละมือ
	results := make([]HandResult, numPlayers)
	for i, hand := range hands {
		results[i] = evaluateHand(hand)
	}

	// 4. แสดงผลไพ่แต่ละผู้เล่น
	fmt.Println()
	for i, hand := range hands {
		cards := []string{}
		for _, c := range hand {
			cards = append(cards, c.String())
		}
		fmt.Printf("Player %d: %s -> %s\n",
			i+1,
			strings.Join(cards, " "),
			rankName(results[i].Rank),
		)
	}
	fmt.Printf("\nCards left in deck: %d\n", len(deck))

	// 5. หาผู้ชนะ
	winnerIdx := []int{0}
	for i := 1; i < numPlayers; i++ {
		cmp := compareHands(results[i], results[winnerIdx[0]])
		if cmp > 0 {
			winnerIdx = []int{i}
		} else if cmp == 0 {
			winnerIdx = append(winnerIdx, i)
		}
	}

	if len(winnerIdx) == 1 {
		w := winnerIdx[0]
		fmt.Printf("*** Winner is Player %d with %s! ***\n", w+1, rankName(results[w].Rank))
	} else {
		names := []string{}
		for _, w := range winnerIdx {
			names = append(names, fmt.Sprintf("Player %d", w+1))
		}
		fmt.Printf("*** It's a tie between %s with %s! ***\n",
			strings.Join(names, " and "),
			rankName(results[winnerIdx[0]].Rank),
		)
	}
}
