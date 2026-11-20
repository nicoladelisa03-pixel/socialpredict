package markets

import (
	"context"
	"strings"
)

// PublicTradingActivity is the public market activity total. It has counts and
// YES/NO volume only. It does not name a trader or carry one person's shares,
// spend, value, or profit.
type PublicTradingActivity struct {
	BetCount         int   `json:"betCount"`
	YesVolume        int64 `json:"yesVolume"`
	NoVolume         int64 `json:"noVolume"`
	ParticipantCount int   `json:"participantCount"`
}

// GetPublicTradingActivity returns public totals for one visible market.
// A proposed, rejected, cancelled, or otherwise hidden market is not found.
func (s *Service) GetPublicTradingActivity(ctx context.Context, marketID int64) (PublicTradingActivity, error) {
	bets, err := s.GetMarketBets(ctx, marketID)
	if err != nil {
		return PublicTradingActivity{}, err
	}
	positions, err := s.GetMarketPositions(ctx, marketID)
	if err != nil {
		return PublicTradingActivity{}, err
	}
	return PublicTradingActivityFrom(bets, positions), nil
}

// GetPublicMarketGroupActivity returns the same public totals across a visible
// group's child markets. A hidden group is not found. One person who holds
// shares in more than one answer is counted once.
func (s *Service) GetPublicMarketGroupActivity(ctx context.Context, groupID int64) (PublicTradingActivity, error) {
	_, members, err := s.marketGroupActivityMembers(ctx, groupID)
	if err != nil {
		return PublicTradingActivity{}, err
	}

	var bets []*BetDisplayInfo
	participants := make(map[string]struct{})
	for _, member := range members {
		memberBets, err := s.GetMarketBets(ctx, member.MarketID)
		if err != nil {
			return PublicTradingActivity{}, err
		}
		bets = append(bets, memberBets...)

		memberPositions, err := s.GetMarketPositions(ctx, member.MarketID)
		if err != nil {
			return PublicTradingActivity{}, err
		}
		for _, position := range activeMarketPositions(memberPositions) {
			if position == nil || position.Username == "" {
				continue
			}
			participants[position.Username] = struct{}{}
		}
	}

	activity := PublicTradingActivityFrom(bets, nil)
	activity.ParticipantCount = len(participants)
	return activity, nil
}

// PublicTradingActivityFrom sums recorded bet amounts by outcome and counts
// holders who still have shares. The inputs may name traders; the result does not.
func PublicTradingActivityFrom(bets []*BetDisplayInfo, positions MarketPositions) PublicTradingActivity {
	var activity PublicTradingActivity
	for _, bet := range bets {
		if bet == nil {
			continue
		}
		activity.BetCount++
		switch strings.ToUpper(strings.TrimSpace(bet.Outcome)) {
		case "YES":
			activity.YesVolume += bet.Amount
		case "NO":
			activity.NoVolume += bet.Amount
		}
	}
	activity.ParticipantCount = len(activeMarketPositions(positions))
	return activity
}
