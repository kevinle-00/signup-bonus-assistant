package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevinle-00/signup-bonus-assistant/backend/internal/recommendations"
)

type CardOfferRepository interface {
	ListActiveCardOffers(ctx context.Context) ([]recommendations.CardOffer, error)
}

type PostgresCardOfferRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresCardOfferRepository(pool *pgxpool.Pool) *PostgresCardOfferRepository {
	return &PostgresCardOfferRepository{pool: pool}
}

func (r *PostgresCardOfferRepository) ListActiveCardOffers(ctx context.Context) ([]recommendations.CardOffer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id::text,
			issuer,
			card_name,
			reward_program,
			reward_type,
			network,
			signup_bonus_points,
			estimated_bonus_value_cents,
			cashback_cents,
			minimum_spend_cents,
			spend_window_days,
			annual_fee_cents,
			travel_credit_cents,
			later_bonus_points,
			later_bonus_condition,
			later_bonus_included_in_mvp_value,
			offer_expires_at,
			eligibility_rules,
			terms_summary
		FROM card_offers
		WHERE is_active = TRUE
		ORDER BY issuer, card_name
	`)
	if err != nil {
		return nil, fmt.Errorf("query active card offers: %w", err)
	}
	defer rows.Close()

	offers := []recommendations.CardOffer{}
	for rows.Next() {
		offer, err := scanCardOffer(rows)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active card offers: %w", err)
	}

	return offers, nil
}

type cardOfferScanner interface {
	Scan(dest ...any) error
}

func scanCardOffer(scanner cardOfferScanner) (recommendations.CardOffer, error) {
	var offer recommendations.CardOffer
	var rewardType string
	var network string
	var offerExpiresAt pgtype.Date
	var eligibilityRulesJSON []byte
	var termsSummaryJSON []byte

	if err := scanner.Scan(
		&offer.ID,
		&offer.Issuer,
		&offer.CardName,
		&offer.RewardProgram,
		&rewardType,
		&network,
		&offer.SignupBonusPoints,
		&offer.EstimatedBonusValueCents,
		&offer.CashbackCents,
		&offer.MinimumSpendCents,
		&offer.SpendWindowDays,
		&offer.AnnualFeeCents,
		&offer.TravelCreditCents,
		&offer.LaterBonusPoints,
		&offer.LaterBonusCondition,
		&offer.LaterBonusIncludedInMVPValue,
		&offerExpiresAt,
		&eligibilityRulesJSON,
		&termsSummaryJSON,
	); err != nil {
		return recommendations.CardOffer{}, fmt.Errorf("scan card offer: %w", err)
	}

	offer.RewardType = recommendations.RewardType(rewardType)
	offer.Network = recommendations.CardNetwork(network)
	if offerExpiresAt.Valid {
		expiresAt := time.Date(offerExpiresAt.Time.Year(), offerExpiresAt.Time.Month(), offerExpiresAt.Time.Day(), 0, 0, 0, 0, time.UTC)
		offer.OfferExpiresAt = &expiresAt
	}
	if err := json.Unmarshal(eligibilityRulesJSON, &offer.EligibilityRules); err != nil {
		return recommendations.CardOffer{}, fmt.Errorf("decode eligibility rules for %s: %w", offer.CardName, err)
	}
	if err := json.Unmarshal(termsSummaryJSON, &offer.TermsSummary); err != nil {
		return recommendations.CardOffer{}, fmt.Errorf("decode terms summary for %s: %w", offer.CardName, err)
	}

	return offer, nil
}
