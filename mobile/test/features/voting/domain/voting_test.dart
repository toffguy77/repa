import 'package:flutter_test/flutter_test.dart';
import 'package:repa/features/voting/domain/voting.dart';

void main() {
  group('VotingQuestion', () {
    test('fromJson creates correct instance', () {
      final json = {
        'question_id': 'q1',
        'text': 'Кто первым убежит при пожаре?',
        'category': 'FUNNY',
        'answered': false,
      };

      final question = VotingQuestion.fromJson(json);

      expect(question.questionId, 'q1');
      expect(question.text, 'Кто первым убежит при пожаре?');
      expect(question.category, 'FUNNY');
      expect(question.answered, false);
    });

    test('toJson produces correct map', () {
      const question = VotingQuestion(
        questionId: 'q1',
        text: 'Test?',
        category: 'HOT',
        answered: true,
      );

      final json = question.toJson();

      expect(json['question_id'], 'q1');
      expect(json['text'], 'Test?');
      expect(json['category'], 'HOT');
      expect(json['answered'], true);
    });
  });

  group('VotingTarget', () {
    test('fromJson with all fields', () {
      final json = {
        'user_id': 'u1',
        'username': 'alice',
        'avatar_emoji': '😎',
        'avatar_url': 'https://example.com/avatar.jpg',
      };

      final target = VotingTarget.fromJson(json);

      expect(target.userId, 'u1');
      expect(target.username, 'alice');
      expect(target.avatarEmoji, '😎');
      expect(target.avatarUrl, 'https://example.com/avatar.jpg');
    });

    test('fromJson with nullable fields null', () {
      final json = {
        'user_id': 'u2',
        'username': 'bob',
        'avatar_emoji': null,
        'avatar_url': null,
      };

      final target = VotingTarget.fromJson(json);

      expect(target.avatarEmoji, isNull);
      expect(target.avatarUrl, isNull);
    });
  });

  group('VotingSession', () {
    test('fromJson creates full session', () {
      final json = {
        'season_id': 's1',
        'questions': [
          {
            'question_id': 'q1',
            'text': 'Q1?',
            'category': 'FUNNY',
            'answered': false,
          },
          {
            'question_id': 'q2',
            'text': 'Q2?',
            'category': 'HOT',
            'answered': true,
          },
        ],
        'targets': [
          {
            'user_id': 'u1',
            'username': 'alice',
            'avatar_emoji': '🐱',
            'avatar_url': null,
          },
        ],
        'progress': {'answered': 1, 'total': 2},
      };

      final session = VotingSession.fromJson(json);

      expect(session.seasonId, 's1');
      expect(session.questions.length, 2);
      expect(session.questions[0].answered, false);
      expect(session.questions[1].answered, true);
      expect(session.targets.length, 1);
      expect(session.targets[0].username, 'alice');
      expect(session.progress.answered, 1);
      expect(session.progress.total, 2);
    });
  });

  group('VoteResultData', () {
    test('fromJson creates correct result', () {
      final json = {
        'vote': {
          'question_id': 'q1',
          'target_id': 'u1',
        },
        'progress': {'answered': 3, 'total': 5},
      };

      final result = VoteResultData.fromJson(json);

      expect(result.vote.questionId, 'q1');
      expect(result.vote.targetId, 'u1');
      expect(result.progress.answered, 3);
      expect(result.progress.total, 5);
    });
  });

  group('GroupVotingProgress', () {
    test('fromJson creates progress', () {
      final json = {
        'voted_count': 3,
        'total_count': 5,
        'quorum_reached': true,
        'quorum_threshold': 0.5,
        'user_voted': true,
      };

      final progress = GroupVotingProgress.fromJson(json);

      expect(progress.votedCount, 3);
      expect(progress.totalCount, 5);
      expect(progress.quorumReached, true);
      expect(progress.quorumThreshold, 0.5);
      expect(progress.userVoted, true);
    });

    test('fromJson with quorum not reached', () {
      final json = {
        'voted_count': 1,
        'total_count': 8,
        'quorum_reached': false,
        'quorum_threshold': 0.4,
        'user_voted': false,
      };

      final progress = GroupVotingProgress.fromJson(json);

      expect(progress.quorumReached, false);
      expect(progress.quorumThreshold, 0.4);
      expect(progress.userVoted, false);
    });
  });
}
