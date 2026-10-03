# Spec Delta

## Purpose

Defines when a group's voting seasons open and close, how a newly created group reaches
its first Reveal without waiting for a calendar week, and what the system does when a
Reveal cannot legitimately happen.

## ADDED Requirements

### Requirement: Kickoff season opens immediately

When a group is created, the system SHALL create its first season as a kickoff season
that is open for voting immediately, so the founder and every member who joins can vote
without waiting for a calendar date.

#### Scenario: Founder can vote right after creating a group

- **WHEN** a user creates a group
- **THEN** the group has a season with status `VOTING` whose voting window has already started
- **AND** the founder can open the voting session for that season without an error

#### Scenario: Joining member can vote immediately

- **WHEN** a user joins a group whose kickoff season is still in `VOTING`
- **THEN** that user can open the voting session and cast votes for the remaining questions

### Requirement: Kickoff season reveals shortly after it first becomes eligible

A kickoff season SHALL be scheduled to reveal within one hour of the moment the group
first satisfies every Reveal eligibility condition, rather than on the next calendar
Friday.

#### Scenario: Reveal is scheduled when the third voter finishes

- **GIVEN** a group with at least 3 members whose kickoff season has 2 members who have completed voting
- **WHEN** a third member completes voting
- **THEN** the kickoff season's reveal time is set to one hour from that moment
- **AND** all members are notified that the Reveal is coming within the hour

#### Scenario: Reveal time is not pushed back by later voters

- **GIVEN** a kickoff season whose reveal time has already been scheduled
- **WHEN** a further member completes voting before the reveal time
- **THEN** the reveal time is unchanged

#### Scenario: Kickoff season reveals at its scheduled time

- **GIVEN** a kickoff season whose scheduled reveal time has passed and which is still eligible
- **WHEN** the system processes pending reveals
- **THEN** the season status becomes `REVEALED` and every member has a card with their results

### Requirement: Kickoff season falls back to the weekly schedule

A kickoff season that never becomes eligible SHALL NOT reveal. The system SHALL keep it
open and SHALL fall back to the regular weekly Friday Reveal once the group becomes
eligible, so a slow-growing group still joins the normal rhythm.

#### Scenario: Group that never reaches the floor keeps waiting

- **GIVEN** a group of 2 members whose kickoff season has been open for a week
- **WHEN** the system processes pending reveals
- **THEN** the season remains in `VOTING` and no card is produced
- **AND** members see that the Reveal is waiting for more people

### Requirement: Weekly seasons target the upcoming Friday with a guaranteed voting window

A weekly season SHALL open at creation time and SHALL reveal at the next Friday 20:00 MSK
that is at least 48 hours away, so voting is never shorter than two days and never longer
than one week plus two days.

#### Scenario: Season created early in the week reveals the same week

- **WHEN** a weekly season is created on a Monday
- **THEN** its reveal time is that same week's Friday at 20:00 MSK

#### Scenario: Season created too close to Friday reveals the following week

- **WHEN** a weekly season is created on a Thursday
- **THEN** its reveal time is the following week's Friday at 20:00 MSK, because the same week's Friday is less than 48 hours away

#### Scenario: Season voting window is open from creation

- **WHEN** a weekly season is created
- **THEN** its voting window starts at creation time, not on a later Monday

### Requirement: An eligible group always has an open season

The system SHALL ensure that any group eligible to reveal has a season open for voting,
creating one promptly after the previous season finishes rather than only at a fixed
weekly moment.

#### Scenario: Weekly season follows a mid-week kickoff reveal

- **GIVEN** a group whose kickoff season reveals on a Wednesday
- **WHEN** the reveal completes
- **THEN** a weekly season is open for voting for that group within the hour

#### Scenario: Friday-to-Sunday discussion window is preserved

- **GIVEN** a weekly season that reveals on Friday at 20:00 MSK
- **WHEN** the reveal completes
- **THEN** no new season opens before that season's end time on Sunday
- **AND** a new season is open for voting by Monday

#### Scenario: Group with no season is repaired automatically

- **GIVEN** an eligible group with no season in `VOTING` and whose most recent season ended more than one hour ago
- **WHEN** the system runs its season maintenance
- **THEN** a weekly season is created for that group

### Requirement: Reveal requires a minimum level of participation

A season SHALL NOT reveal unless the group has at least 3 members and at least 3 members
have completed voting. These floors apply in addition to the existing quorum percentage
and SHALL NOT be bypassed by the forced-reveal retry path.

#### Scenario: Solo group never produces an empty card

- **GIVEN** a group with one member and a season whose reveal time has passed
- **WHEN** the system processes the reveal, including all retry attempts
- **THEN** the season is never set to `REVEALED`
- **AND** no card is generated for that member

#### Scenario: Percentage quorum alone is not sufficient

- **GIVEN** a group of 4 members where 2 members have completed voting, satisfying the 40% quorum
- **WHEN** the system processes the reveal after exhausting its retry attempts
- **THEN** the season is not revealed, because fewer than 3 members have voted

#### Scenario: Floor met and quorum met reveals normally

- **GIVEN** a group of 6 members where 4 members have completed voting
- **WHEN** the reveal time passes
- **THEN** the season is revealed and results are aggregated

### Requirement: An ineligible season is postponed, not abandoned

When a season reaches its reveal time but cannot legitimately reveal, the system SHALL
postpone it to the next weekly Reveal slot and SHALL keep voting open, so accumulated
votes are preserved rather than discarded.

#### Scenario: Season is postponed by one week

- **GIVEN** a season that is ineligible to reveal after exhausting its retry attempts
- **WHEN** the system finishes processing the reveal
- **THEN** the season's reveal time moves to the next Friday 20:00 MSK and its status stays `VOTING`
- **AND** votes already cast remain valid

#### Scenario: Members are told the Reveal was postponed

- **WHEN** a season is postponed because the participation floor was not met
- **THEN** group members are notified that the Reveal is waiting and how many more people need to vote

#### Scenario: Previously postponed season reveals once eligible

- **GIVEN** a season that was postponed last week and has since reached the participation floor
- **WHEN** its new reveal time passes
- **THEN** the season reveals normally

### Requirement: Clients can explain the state of a pending Reveal

Group and season responses SHALL expose enough information for the app to tell a member
whether the Reveal is scheduled, waiting for more people, or postponed, including how
many additional voters are required.

#### Scenario: App shows how many more voters are needed

- **GIVEN** a group of 2 members with an open kickoff season
- **WHEN** a member loads the group
- **THEN** the response states that the Reveal is waiting and that 1 more member needs to join and vote

#### Scenario: App shows the scheduled kickoff reveal time

- **GIVEN** a kickoff season whose reveal has been scheduled for one hour from now
- **WHEN** a member loads the group
- **THEN** the response includes that reveal time so the app can count down to it
