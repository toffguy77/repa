# Spec Delta

## ADDED Requirements

### Requirement: A membership records who brought the member in

A membership SHALL record the member who referred it, when one is identified, separately from
the channel the member arrived through.

#### Scenario: Referrer and channel are recorded independently

- **GIVEN** a user joining through a card shared by a specific member
- **WHEN** the join completes
- **THEN** the membership records both the card as the channel and that member as the referrer

#### Scenario: A channel without a referrer is still recorded

- **GIVEN** a user joining through a plain link with no referrer
- **WHEN** the join completes
- **THEN** the channel is recorded and the referrer is empty
