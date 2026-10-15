# Spec Delta

## ADDED Requirements

### Requirement: A Telegram-sourced join is attributed to Telegram

A join that originates from an invite the bot handed out in a chat SHALL be recorded with
Telegram as the channel, so chat-driven acquisition is measurable separately from card-driven
acquisition.

#### Scenario: Joining from the bot's link is attributed

- **GIVEN** an invite link the bot sent in a connected chat
- **WHEN** someone opens it and joins
- **THEN** their membership records Telegram as the source

#### Scenario: Telegram joins appear in the funnel

- **WHEN** an administrator reads statistics after a Telegram-sourced join
- **THEN** the Telegram source is counted separately from the card source
