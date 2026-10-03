# anticipation Specification

## Purpose
What a member may learn about votes concerning them before the Reveal — enough to feel watched,
never enough to know anything — so the week between votes and results has something in it.

## Requirements

### Requirement: A member can see how many people answered about them

During an open season, a member SHALL be able to see the number of other members who have answered
at least one question about them.

#### Scenario: The count reflects voters, not votes

- **GIVEN** a season where one member answered three questions about me
- **WHEN** I read my anticipation state
- **THEN** the count is 1

#### Scenario: The count starts at zero

- **GIVEN** a season in which nobody has voted about me
- **WHEN** I read my anticipation state
- **THEN** the count is 0

#### Scenario: My own votes do not count toward it

- **GIVEN** a season in which I voted about other people but nobody voted about me
- **WHEN** I read my anticipation state
- **THEN** the count is 0

#### Scenario: The count is available before I have voted myself

- **GIVEN** a member who has not voted
- **WHEN** they read their anticipation state
- **THEN** the count is returned rather than refused

### Requirement: The count never identifies anyone

The anticipation state SHALL contain no name, user id, avatar, question, attribute, or answer. A
member SHALL learn only a number and, from Thursday, one category emoji.

#### Scenario: No identities in the response

- **WHEN** a member reads their anticipation state
- **THEN** the response contains no member name, user id, or avatar

#### Scenario: No attributes in the response

- **WHEN** a member reads their anticipation state
- **THEN** the response contains no question text, attribute, or percentage

#### Scenario: The count cannot be used to deduce an individual

- **GIVEN** a group of three members where exactly one has voted about me
- **WHEN** I read my anticipation state
- **THEN** I am told one person has answered, and nothing that distinguishes which

### Requirement: The leading category is teased as an emoji, late in the week

From Thursday onward, if a member has received votes, the anticipation state SHALL include the emoji
of the category they have received the most votes in — the emoji alone, never the category name or
the question.

#### Scenario: The teaser appears on Thursday

- **GIVEN** it is Thursday and I have received votes
- **WHEN** I read my anticipation state
- **THEN** the leading category's emoji is included

#### Scenario: No teaser earlier in the week

- **GIVEN** it is Monday and I have received votes
- **WHEN** I read my anticipation state
- **THEN** no teaser is included

#### Scenario: No teaser without votes

- **GIVEN** it is Thursday and nobody has voted about me
- **WHEN** I read my anticipation state
- **THEN** no teaser is included

#### Scenario: The teaser is an emoji and nothing else

- **WHEN** a teaser is included
- **THEN** it is a single category emoji, with no category name, question text, or count attached

### Requirement: Anticipation is not available once results exist

The anticipation state SHALL only be served while a season is open. After the Reveal the member has
their card, and a partial signal alongside a full result is noise.

#### Scenario: A revealed season has no anticipation state

- **GIVEN** a season that has revealed
- **WHEN** a member requests their anticipation state
- **THEN** the request is refused and they are directed to their results

#### Scenario: Only members can read it

- **WHEN** someone who is not a member of the group requests the anticipation state
- **THEN** the request is refused

### Requirement: A vote about a member reaches them promptly

When someone answers a question about a member, that member SHALL receive a signal that the number
went up, without being told who voted or what they chose.

#### Scenario: The first vote about me notifies me

- **GIVEN** nobody has voted about me yet
- **WHEN** another member answers a question about me
- **THEN** I receive a notification that someone answered

#### Scenario: The notification names nobody and nothing

- **WHEN** I receive that notification
- **THEN** it contains no voter name and no attribute

#### Scenario: Repeated votes do not become a stream of notifications

- **GIVEN** I have already been notified today that someone answered about me
- **WHEN** further members answer questions about me on the same day
- **THEN** I am not notified again that day

#### Scenario: My own voting never notifies me

- **WHEN** I answer questions about other people
- **THEN** I receive no notification about myself
