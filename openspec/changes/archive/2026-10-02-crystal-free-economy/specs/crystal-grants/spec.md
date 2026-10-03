# Spec Delta

## Purpose

How a user obtains crystals without paying — the welcome grant, referral rewards, and
achievement payouts — and the guarantee that no grant is ever paid twice.

## ADDED Requirements

### Requirement: Every new user can try the core hook for free

A newly registered user SHALL receive a one-time welcome grant sufficient for exactly one
detector, so the product's central hook is experienced before it is sold.

#### Scenario: A new user starts with a usable balance

- **WHEN** a user completes registration
- **THEN** their crystal balance is enough to buy one detector
- **AND** the grant appears in their crystal history with a readable reason

#### Scenario: The welcome grant is paid once

- **GIVEN** a user who has already received the welcome grant
- **WHEN** the grant is attempted again
- **THEN** their balance does not change

#### Scenario: Spending the welcome grant does not re-trigger it

- **GIVEN** a user who spent their welcome grant on a detector
- **WHEN** they open the app again
- **THEN** no further welcome grant is made

### Requirement: Inviting someone pays when they actually play

When a member brought in by another user completes their first voting session, the inviter
SHALL be granted crystals. The grant SHALL NOT be triggered by the join alone.

#### Scenario: The inviter is paid after the invitee votes

- **GIVEN** a user who joined through another member's shared card
- **WHEN** they complete their first voting session
- **THEN** the member who invited them receives a crystal grant
- **AND** the grant names the referral as its reason

#### Scenario: Joining alone pays nothing

- **GIVEN** a user who joined through another member's shared card
- **WHEN** they have not completed a voting session
- **THEN** the inviter has received no referral grant

#### Scenario: A referral pays once per invited member

- **GIVEN** an invitee whose first completed session already paid their inviter
- **WHEN** that invitee completes further sessions
- **THEN** no further grant is made for that invitee

#### Scenario: An unattributed join pays nobody

- **GIVEN** a user who joined without a referrer
- **WHEN** they complete their first voting session
- **THEN** no referral grant is made

#### Scenario: Self-referral pays nothing

- **GIVEN** a join whose referrer is the joining user
- **WHEN** they complete their first voting session
- **THEN** no grant is made

### Requirement: A shared card identifies who shared it

A card's invite link SHALL identify the member who shared it, so a referral reward reaches the
person who actually invited rather than the group's admin.

#### Scenario: The card's link carries its owner

- **WHEN** a card is generated for a member
- **THEN** the invite link encoded on it identifies that member as the referrer

#### Scenario: Joining through a card records the referrer

- **GIVEN** a card shared by a member
- **WHEN** someone joins through its link
- **THEN** their membership records that member as the referrer

#### Scenario: A link without a referrer still joins

- **WHEN** someone joins through a link carrying no referrer
- **THEN** the join succeeds and no referrer is recorded

#### Scenario: An invalid referrer does not fail the join

- **WHEN** someone joins through a link whose referrer is not a real member of that group
- **THEN** the join succeeds and no referrer is recorded

### Requirement: Milestone achievements pay out

Achievements that mark sustained play or successful recruiting SHALL carry crystal grants, so
the retention and acquisition loops pay in the same currency the product sells.

#### Scenario: A voting-streak milestone grants crystals

- **WHEN** a user unlocks a voting-streak milestone achievement
- **THEN** they receive a crystal grant naming that achievement

#### Scenario: The recruiter achievement grants crystals

- **WHEN** a user unlocks the recruiter achievement
- **THEN** they receive a crystal grant

#### Scenario: A non-granting achievement pays nothing

- **WHEN** a user unlocks an achievement that is not a milestone payout
- **THEN** their balance does not change

### Requirement: Grants are idempotent

Every grant SHALL carry a deterministic identity, and attempting the same grant more than once
SHALL pay exactly once. A retry, a replayed job, or a concurrent duplicate SHALL NOT double-pay.

#### Scenario: A repeated grant attempt pays once

- **WHEN** the same grant is attempted twice
- **THEN** the balance reflects a single payment

#### Scenario: A duplicate attempt is not an error for the caller

- **WHEN** a grant that was already paid is attempted again
- **THEN** the caller is not told the operation failed

#### Scenario: Different grants to the same user both pay

- **GIVEN** a user who received the welcome grant
- **WHEN** they also earn a referral grant
- **THEN** both appear in their history and both count toward the balance

### Requirement: The user can see where free crystals came from

Crystal history SHALL show each grant with a reason a user can understand, so a balance that
grew without a purchase is not mysterious.

#### Scenario: History distinguishes grants from purchases

- **WHEN** a user reads their crystal history after a welcome grant and a purchase
- **THEN** the two entries are distinguishable and each carries its own reason
