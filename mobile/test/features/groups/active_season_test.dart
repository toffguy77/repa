import 'package:flutter_test/flutter_test.dart';
import 'package:repa/features/groups/domain/group.dart';

ActiveSeason season(Map<String, dynamic> overrides) => ActiveSeason.fromJson({
      'id': 's1',
      'status': 'VOTING',
      'reveal_at': '2026-10-09T17:00:00Z',
      'voted_count': 2,
      'total_count': 4,
      'user_voted': true,
      ...overrides,
    });

void main() {
  group('ActiveSeason reveal state parsing', () {
    test('parses every backend state', () {
      expect(season({'reveal_state': 'SCHEDULED'}).revealState,
          SeasonRevealState.scheduled);
      expect(season({'reveal_state': 'WAITING_FOR_MEMBERS'}).revealState,
          SeasonRevealState.waitingForMembers);
      expect(season({'reveal_state': 'WAITING_FOR_VOTERS'}).revealState,
          SeasonRevealState.waitingForVoters);
      expect(season({'reveal_state': 'POSTPONED'}).revealState,
          SeasonRevealState.postponed);
      expect(season({'reveal_state': 'REVEALED'}).revealState,
          SeasonRevealState.revealed);
    });

    test('falls back to scheduled for a missing or unknown state', () {
      expect(season({}).revealState, SeasonRevealState.scheduled);
      expect(season({'reveal_state': 'SOMETHING_NEW'}).revealState,
          SeasonRevealState.scheduled);
    });

    test('round-trips the needed counts', () {
      final s = season({
        'reveal_state': 'WAITING_FOR_MEMBERS',
        'members_needed': 1,
        'voters_needed': 2,
      });
      expect(s.membersNeeded, 1);
      expect(s.votersNeeded, 2);
      expect(s.toJson()['reveal_state'], 'WAITING_FOR_MEMBERS');
      expect(s.toJson()['members_needed'], 1);
    });

    test('defaults the needed counts to zero', () {
      expect(season({}).membersNeeded, 0);
      expect(season({}).votersNeeded, 0);
    });
  });

  group('ActiveSeason reveal copy', () {
    test('waiting for members explains how many to invite', () {
      final s =
          season({'reveal_state': 'WAITING_FOR_MEMBERS', 'members_needed': 1});
      expect(s.revealHeadline, 'Нужно больше людей');
      expect(s.revealExplanation, contains('1 человека'));
      expect(s.isWaitingForPeople, isTrue);
    });

    test('waiting for voters explains how many must vote', () {
      final s =
          season({'reveal_state': 'WAITING_FOR_VOTERS', 'voters_needed': 2});
      expect(s.revealHeadline, 'Нужно больше голосов');
      expect(s.revealExplanation, contains('2 человека'));
      expect(s.isWaitingForPeople, isTrue);
    });

    test('postponed reassures that votes are kept', () {
      final s = season({'reveal_state': 'POSTPONED'});
      expect(s.revealHeadline, 'Репа перенесена');
      expect(s.revealExplanation, contains('сохранены'));
      expect(s.isWaitingForPeople, isFalse);
    });

    test('scheduled keeps the Friday promise', () {
      final s = season({'reveal_state': 'SCHEDULED'});
      expect(s.revealExplanation, contains('пятницу'));
      expect(s.isWaitingForPeople, isFalse);
    });

    test('plural forms follow Russian rules', () {
      expect(
          season({'reveal_state': 'WAITING_FOR_VOTERS', 'voters_needed': 1})
              .revealExplanation,
          contains('1 человека'));
      expect(
          season({'reveal_state': 'WAITING_FOR_VOTERS', 'voters_needed': 5})
              .revealExplanation,
          contains('5 человек'));
      expect(
          season({'reveal_state': 'WAITING_FOR_VOTERS', 'voters_needed': 11})
              .revealExplanation,
          contains('11 человек'));
      expect(
          season({'reveal_state': 'WAITING_FOR_VOTERS', 'voters_needed': 22})
              .revealExplanation,
          contains('22 человека'));
    });
  });
}
