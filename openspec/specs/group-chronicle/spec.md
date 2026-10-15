# group-chronicle Specification

## Purpose
The shared record a group accumulates by playing: which results stood out in each past season, and who
in the group predicts it best.

## Requirements

### Requirement: A group keeps a readable record of its past seasons

A group SHALL have a chronicle: for every season that legitimately revealed, the standout result for
each of that season's questions — the question, the member who led it, and that member's share of the
vote. Any member of the group SHALL be able to read it.

The chronicle is a record, not a feed: it SHALL contain only what already happened and SHALL NOT change
when the current season's votes change.

#### Scenario: A member reads the group's history

- **GIVEN** a group with three revealed seasons
- **WHEN** a member opens the chronicle
- **THEN** all three seasons appear, newest first, each with the standout result per question

#### Scenario: The open season is not in the chronicle

- **GIVEN** a group with an open season that members are still voting in
- **WHEN** a member opens the chronicle
- **THEN** the open season does not appear, because its results do not exist yet

#### Scenario: A postponed or abandoned season leaves no entry

- **GIVEN** a season that reached its reveal time without enough voters and was postponed
- **WHEN** a member opens the chronicle
- **THEN** that season does not appear

#### Scenario: A new group's chronicle is empty, not broken

- **GIVEN** a group that has never revealed a season
- **WHEN** a member opens the chronicle
- **THEN** it says the group has no history yet and what will put something in it

#### Scenario: Non-members cannot read it

- **WHEN** someone who is not a member of the group requests the chronicle
- **THEN** the request is refused

### Requirement: The chronicle survives members leaving

A chronicle entry SHALL remain readable after the member it names leaves the group, because it records
something that happened. It SHALL NOT, however, name someone the reader has blocked.

#### Scenario: A departed member's result stays in the record

- **GIVEN** a chronicle entry naming a member who has since left
- **WHEN** a member reads the chronicle
- **THEN** the entry is still there

#### Scenario: A blocked member is not named to the blocker

- **GIVEN** a reader who has blocked another member
- **WHEN** they read the chronicle
- **THEN** entries whose standout result is that member are withheld from them

### Requirement: The group can see who predicts it best

The group SHALL have a standing that orders its members by how often their votes matched the result the
group arrived at. The standing SHALL state how many revealed seasons each member's place is based on,
because a place earned over one season and one earned over ten are not comparable.

#### Scenario: Members are ordered by how well they predict

- **GIVEN** a group whose members predict it with differing success
- **WHEN** a member opens the chronicle
- **THEN** the standing lists members from best to worst predictor, each with the number of seasons
  their place is based on

#### Scenario: A member who has never voted in a revealed season is not ranked

- **GIVEN** a member who joined after the last reveal
- **WHEN** the standing is shown
- **THEN** they are listed as not yet ranked rather than ranked at zero

#### Scenario: The standing is withheld until it means something

- **GIVEN** a group with fewer than two revealed seasons
- **WHEN** a member opens the chronicle
- **THEN** the standing is not shown, and the app says what it is waiting for

### Requirement: The standing exposes an order, not a measurement

The standing SHALL expose only each member's position and the number of seasons it is based on. It
SHALL NOT expose the underlying accuracy figure.

A published accuracy figure is an inference channel rather than merely a score, and the chronicle is
what makes it one: because the chronicle publishes which member won each question, a member whose
accuracy is at either extreme has their individual votes on those questions recovered exactly. The
figure is also a rolling average whose weight is the number of seasons played, so comparing one week's
figure with the next solves for that week's match count — and a standing that hands out every member's
figure at once makes that cheap to do for the whole group. A position cannot be arithmetically reduced
to a vote, and a position is all the standing needs to convey.

#### Scenario: The standing carries no accuracy figure

- **WHEN** the standing is returned
- **THEN** it contains each member's position and seasons played, and no accuracy figure, question,
  target or vote

#### Scenario: A member who matched every result is still only first

- **GIVEN** a small group in which one member matched the result on every question
- **WHEN** the standing is shown
- **THEN** it places them first without stating how often they matched
