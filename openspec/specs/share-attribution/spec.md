# share-attribution Specification

## Purpose
What a shared Repa artifact carries so a viewer can act on it, and what the system records
about how each member arrived — so the path from a shared card to a joined member is measured
rather than assumed.

## Requirements

### Requirement: A shared card carries a way back into the product

The generated card image SHALL include the group's invite code, a scannable code encoding the
invite link, and a short instruction. A viewer SHALL be able to act on the card without
needing any other message.

#### Scenario: The card shows the invite code

- **WHEN** a card is generated for a member of a group
- **THEN** the image contains that group's invite code

#### Scenario: The card shows a scannable code

- **WHEN** a card is generated
- **THEN** the image contains a QR code that encodes the group's invite link

#### Scenario: The card tells the viewer what to do

- **WHEN** a card is generated
- **THEN** the image contains a short instruction naming the app and the action

#### Scenario: Card generation still succeeds if the code cannot be rendered

- **GIVEN** scannable-code rendering fails
- **WHEN** a card is generated
- **THEN** the card is still produced with the invite code and instruction
- **AND** the failure is logged rather than losing the member's card

### Requirement: Sharing carries a personal invite link

Sharing a card SHALL attach a link containing the group's invite code and a marker for the
channel it was shared through. The share SHALL NOT be a generic product URL.

#### Scenario: The share text contains the group's code

- **WHEN** a member shares their card
- **THEN** the shared text contains a link including that group's invite code

#### Scenario: The channel is distinguishable

- **WHEN** the same card is shared to the system share sheet and to a Telegram chat
- **THEN** the two shares carry different channel markers

#### Scenario: Opening a shared link lands on the group

- **GIVEN** a shared link for a group
- **WHEN** a recipient opens it
- **THEN** they are taken to that group's join flow

### Requirement: The system records how a member arrived

When a user joins a group, the system SHALL record the route they came through: an opened
link, a typed code, a scanned or shared card, Telegram, or unknown when no source is given.

#### Scenario: Joining through a shared card is attributed

- **WHEN** a user joins after opening a link shared from a card
- **THEN** their membership records the card as the source

#### Scenario: Joining by typing a code is attributed

- **WHEN** a user joins by typing the code into the join field
- **THEN** their membership records the typed code as the source

#### Scenario: A join with no source is still allowed

- **WHEN** a user joins without the client supplying a source
- **THEN** the join succeeds and the source is recorded as unknown

#### Scenario: An unrecognised source does not fail the join

- **WHEN** a client supplies a source value the server does not recognise
- **THEN** the join succeeds and the source is recorded as unknown

### Requirement: Share actions are counted

The system SHALL record each share action with its season and channel, so shares and the joins
attributed to them can be compared.

#### Scenario: A share is recorded once per action

- **WHEN** a member shares their card twice
- **THEN** two share events are recorded for that member and season

#### Scenario: Share events do not identify who voted

- **WHEN** share events are read
- **THEN** they contain no vote, answer, or voter information

#### Scenario: A failed share recording does not break sharing

- **GIVEN** recording a share event fails
- **WHEN** a member shares their card
- **THEN** the share still happens from the member's point of view

### Requirement: Administrators can see the funnel

Admin statistics SHALL expose the number of shares and the number of joins attributed to each
source, so the acquisition loop can be judged rather than guessed.

#### Scenario: Shares and attributed joins are both visible

- **WHEN** an administrator reads statistics
- **THEN** the response includes share counts by channel and join counts by source

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
