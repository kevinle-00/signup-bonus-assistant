#!/usr/bin/env python3
"""Generate SQL seed data from curated card offer YAML."""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

import yaml


ROOT = Path(__file__).resolve().parents[1]
INPUT_PATH = ROOT / "data" / "card_offers_curated.yaml"
OUTPUT_PATH = ROOT / "backend" / "seed" / "card_offers_seed.sql"

POINT_VALUE_CENTS = {
    "qantas_points": 1.0,
    "velocity_points": 1.0,
    "bank_points": 0.4,
    "flexible_points": 0.8,
    "cashback": 0.0,
    "travel_perks": 0.0,
}

COLUMNS = [
    "issuer",
    "card_name",
    "reward_program",
    "reward_type",
    "network",
    "signup_bonus_points",
    "estimated_bonus_value_cents",
    "cashback_cents",
    "minimum_spend_cents",
    "spend_window_days",
    "annual_fee_cents",
    "travel_credit_cents",
    "later_bonus_points",
    "later_bonus_condition",
    "later_bonus_included_in_mvp_value",
    "offer_expires_at",
    "source_url",
    "source_checked_at",
    "data_quality",
    "eligibility_rules",
    "terms_summary",
    "is_active",
]


def sql_string(value: Any) -> str:
    if value is None:
        return "NULL"
    value = str(value)
    return "'" + value.replace("'", "''") + "'"


def load_curated_yaml() -> dict[str, Any]:
    with INPUT_PATH.open("r", encoding="utf-8") as source:
        return yaml.safe_load(source)


def sql_date(value: str | None) -> str:
    return "NULL" if value in (None, "") else sql_string(value)


def sql_bool(value: bool) -> str:
    return "TRUE" if value else "FALSE"


def sql_json(value: Any) -> str:
    encoded = json.dumps(value, ensure_ascii=False, separators=(",", ":"))
    return sql_string(encoded) + "::jsonb"


def infer_rule_type(note: str) -> str:
    lowered = note.lower()
    if "american express card members" in lowered:
        return "new_amex_card_members_only"
    if "manual review" in lowered:
        return "manual_review"
    if "new card" in lowered or "new cardholders" in lowered:
        return "new_cardholders_only"
    if "currently hold" in lowered or "current cardholder" in lowered:
        return "not_current_cardholder"
    if "held" in lowered or "last 24 months" in lowered or "last 18 months" in lowered:
        return "not_held_recently"
    return "manual_review"


def note_to_string(note: Any) -> str:
    if isinstance(note, dict):
        return "; ".join(f"{key}: {value}" for key, value in note.items())
    return str(note)


def infer_window_days(note: str) -> int | None:
    lowered = note.lower()
    if "24 months" in lowered:
        return 730
    if "18 months" in lowered:
        return 540
    return None


def eligibility_rules(notes: list[str]) -> list[dict[str, Any]]:
    rules = []
    for note in notes:
        note = note_to_string(note)
        rule: dict[str, Any] = {
            "type": infer_rule_type(note),
            "description": note,
        }
        window_days = infer_window_days(note)
        if window_days is not None:
            rule["windowDays"] = window_days
        rules.append(rule)
    return rules


def estimated_bonus_value_cents(offer: dict[str, Any]) -> int:
    reward_type = offer["reward_type"]
    if reward_type not in POINT_VALUE_CENTS:
        raise ValueError(f"Unknown reward_type {reward_type!r} for {offer['card_name']}")
    point_value = POINT_VALUE_CENTS[reward_type]
    points = int(offer.get("signup_bonus_points") or 0)
    cashback = int(offer.get("cashback_cents") or 0)
    return round(points * point_value + cashback)


def sql_value(value: Any) -> str:
    if value is None:
        return "NULL"
    if isinstance(value, bool):
        return sql_bool(value)
    if isinstance(value, int):
        return str(value)
    if isinstance(value, str):
        return sql_string(value)
    raise TypeError(f"Unsupported SQL value: {value!r}")


def row_values(offer: dict[str, Any]) -> list[str]:
    data_quality = offer.get("data_quality") or "verified"
    rules = eligibility_rules(offer.get("eligibility_notes") or [])
    terms = [note_to_string(note) for note in offer.get("terms_notes") or []]

    values: dict[str, Any] = {
        "issuer": offer["issuer"],
        "card_name": offer["card_name"],
        "reward_program": offer["reward_program"],
        "reward_type": offer["reward_type"],
        "network": offer["network"],
        "signup_bonus_points": int(offer.get("signup_bonus_points") or 0),
        "estimated_bonus_value_cents": estimated_bonus_value_cents(offer),
        "cashback_cents": int(offer.get("cashback_cents") or 0),
        "minimum_spend_cents": int(offer["minimum_spend_cents"]),
        "spend_window_days": int(offer["spend_window_days"]),
        "annual_fee_cents": int(offer.get("annual_fee_cents") or 0),
        "travel_credit_cents": int(offer.get("travel_credit_cents") or 0),
        "later_bonus_points": int(offer.get("later_bonus_points") or 0),
        "later_bonus_condition": offer.get("later_bonus_condition"),
        "later_bonus_included_in_mvp_value": bool(offer.get("later_bonus_included_in_mvp_value") or False),
        "offer_expires_at": offer.get("offer_expires_at"),
        "source_url": offer.get("source_url"),
        "source_checked_at": offer.get("source_checked_at"),
        "data_quality": data_quality,
        "eligibility_rules": rules,
        "terms_summary": terms,
        "is_active": True,
    }

    output = []
    for column in COLUMNS:
        if column in {"eligibility_rules", "terms_summary"}:
            output.append(sql_json(values[column]))
        elif column in {"offer_expires_at", "source_checked_at"}:
            output.append(sql_date(values[column]))
        else:
            output.append(sql_value(values[column]))
    return output


def main() -> int:
    data = load_curated_yaml()

    offers = data.get("credit_card_signup_offers") or []
    if not offers:
        raise SystemExit(f"No offers found in {INPUT_PATH}")

    rows = [row_values(offer) for offer in offers]

    lines = [
        "-- Generated by scripts/generate_card_offer_seed.py.",
        "-- Do not edit this file directly; update data/card_offers_curated.yaml and regenerate.",
        "",
        "INSERT INTO card_offers (",
        "  " + ",\n  ".join(COLUMNS),
        ") VALUES",
    ]

    rendered_rows = []
    for row in rows:
        rendered_rows.append("  (" + ", ".join(row) + ")")
    lines.append(",\n".join(rendered_rows))

    update_columns = [column for column in COLUMNS if column not in {"issuer", "card_name"}]
    lines.extend(
        [
            "ON CONFLICT (issuer, card_name) DO UPDATE SET",
            "  " + ",\n  ".join(f"{column} = EXCLUDED.{column}" for column in update_columns) + ";",
            "",
        ]
    )

    OUTPUT_PATH.parent.mkdir(parents=True, exist_ok=True)
    OUTPUT_PATH.write_text("\n".join(lines), encoding="utf-8")
    print(f"Wrote {len(rows)} card offers to {OUTPUT_PATH.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
