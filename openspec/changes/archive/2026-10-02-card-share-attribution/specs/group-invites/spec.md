# Spec Delta

## ADDED Requirements

### Requirement: An invite can be delivered by scanning

An invite SHALL be deliverable as a scannable code in addition to a link and a typed code, so
a person looking at someone's screen or a printed card can join without typing anything.

#### Scenario: Scanning a card's code opens the join flow

- **GIVEN** a card showing a group's scannable code
- **WHEN** a person scans it with a phone camera
- **THEN** the invite link opens and offers to join that group

#### Scenario: The scannable code encodes the same invite as the text code

- **WHEN** a card is generated for a group
- **THEN** the scannable code resolves to the same group as the printed code next to it
