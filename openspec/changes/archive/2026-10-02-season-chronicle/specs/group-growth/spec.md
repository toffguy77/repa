# Spec Delta

## Purpose

What the product tells members about the consequences of their group's size.

## ADDED Requirements

### Requirement: Members are told what the next group-size threshold changes

While a group is below a size threshold that changes what the group can do, the app SHALL state how
many more members are needed and what crossing it changes.

#### Scenario: A group that cannot reveal yet

- **GIVEN** a group with two members
- **WHEN** a member opens the group
- **THEN** the app says one more member is needed and that the Reveal cannot happen until then

#### Scenario: A group that can reveal but has no detector

- **GIVEN** a group with three members
- **WHEN** a member opens the group
- **THEN** the app says two more members are needed and that the detector becomes available then

#### Scenario: The threshold is counted from members the reader can be rated by

- **GIVEN** a group of five in which the reader has blocked two members
- **WHEN** the app states the threshold
- **THEN** it counts the members the reader can actually be rated by, consistent with the anonymity
  warning

### Requirement: The product SHALL NOT invent thresholds

Above the largest threshold that actually changes something, the app SHALL say that the group is big
enough rather than naming a further number. A size at which nothing changes SHALL NOT be presented as
an unlock.

#### Scenario: A group past every threshold

- **GIVEN** a group of twelve members
- **WHEN** a member opens the group
- **THEN** the app states the group is big enough and names no further target

#### Scenario: A harder rule is not sold as a reward

- **GIVEN** a group approaching a size at which the share of members required to vote increases
- **WHEN** the app describes group size
- **THEN** it does not present that size as something to unlock

### Requirement: The thresholds come from the rules themselves

The sizes and the descriptions the app shows SHALL be derived from the same definitions the reveal and
detector rules use, so a rule change cannot leave the explanation behind.

#### Scenario: Changing a rule changes what members are told

- **WHEN** the minimum group size for the detector changes
- **THEN** the number the app shows changes with it, without a second place to edit
