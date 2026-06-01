-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE card_offers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  issuer TEXT NOT NULL,
  card_name TEXT NOT NULL,
  reward_program TEXT NOT NULL,
  reward_type TEXT NOT NULL CHECK (reward_type IN (
    'qantas_points',
    'velocity_points',
    'cashback',
    'flexible_points',
    'bank_points',
    'travel_perks'
  )),
  network TEXT NOT NULL CHECK (network IN ('visa', 'mastercard', 'amex')),

  signup_bonus_points INTEGER NOT NULL DEFAULT 0,
  estimated_bonus_value_cents INTEGER NOT NULL DEFAULT 0,
  cashback_cents INTEGER NOT NULL DEFAULT 0,
  minimum_spend_cents INTEGER NOT NULL,
  spend_window_days INTEGER NOT NULL,

  annual_fee_cents INTEGER NOT NULL DEFAULT 0,
  travel_credit_cents INTEGER NOT NULL DEFAULT 0,

  later_bonus_points INTEGER NOT NULL DEFAULT 0,
  later_bonus_condition TEXT,
  later_bonus_included_in_mvp_value BOOLEAN NOT NULL DEFAULT FALSE,

  offer_expires_at DATE,
  source_url TEXT,
  source_checked_at DATE,
  data_quality TEXT NOT NULL DEFAULT 'verified' CHECK (data_quality IN ('verified', 'generated_sample')),

  eligibility_rules JSONB NOT NULL DEFAULT '[]'::jsonb,
  terms_summary JSONB NOT NULL DEFAULT '[]'::jsonb,

  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT card_offers_issuer_card_name_unique UNIQUE (issuer, card_name)
);

CREATE INDEX card_offers_active_idx ON card_offers (is_active);
CREATE INDEX card_offers_reward_type_idx ON card_offers (reward_type);
CREATE INDEX card_offers_network_idx ON card_offers (network);
CREATE INDEX card_offers_estimated_bonus_value_idx ON card_offers (estimated_bonus_value_cents DESC);

-- +goose Down
DROP TABLE IF EXISTS card_offers;
