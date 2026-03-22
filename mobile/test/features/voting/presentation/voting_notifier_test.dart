import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/features/voting/data/voting_repository.dart';
import 'package:repa/features/voting/domain/voting.dart';
import 'package:repa/features/voting/presentation/voting_notifier.dart';
import 'package:repa/core/api/api_client.dart';

class MockVotingRepository extends Mock implements VotingRepository {}

VotingSession _makeSession({
  int answeredCount = 0,
  int totalQuestions = 3,
}) {
  final questions = List.generate(
    totalQuestions,
    (i) => VotingQuestion(
      questionId: 'q$i',
      text: 'Question $i?',
      category: 'FUNNY',
      answered: i < answeredCount,
    ),
  );

  return VotingSession(
    seasonId: 's1',
    questions: questions,
    targets: const [
      VotingTarget(userId: 'u1', username: 'alice'),
      VotingTarget(userId: 'u2', username: 'bob'),
      VotingTarget(userId: 'u3', username: 'charlie'),
    ],
    progress: VotingProgress(answered: answeredCount, total: totalQuestions),
  );
}

const _groupProgress = GroupVotingProgress(
  votedCount: 3,
  totalCount: 5,
  quorumReached: true,
  quorumThreshold: 0.5,
  userVoted: true,
);

void main() {
  late MockVotingRepository mockRepo;

  setUp(() {
    mockRepo = MockVotingRepository();
  });

  group('VotingNotifier.loadSession', () {
    test('loads session and sets state', () async {
      final session = _makeSession();
      when(() => mockRepo.getVotingSession('s1'))
          .thenAnswer((_) async => session);

      final notifier = VotingNotifier(mockRepo, 's1');

      await notifier.loadSession();

      expect(notifier.state.loading, false);
      expect(notifier.state.session, session);
      expect(notifier.state.currentIndex, 0);
      expect(notifier.state.completed, false);
      expect(notifier.shuffledTargets.length, 3);
    });

    test('marks completed if all questions answered', () async {
      final session = _makeSession(answeredCount: 3, totalQuestions: 3);
      when(() => mockRepo.getVotingSession('s1'))
          .thenAnswer((_) async => session);
      when(() => mockRepo.getVotingProgress('s1'))
          .thenAnswer((_) async => _groupProgress);

      final notifier = VotingNotifier(mockRepo, 's1');

      await notifier.loadSession();

      expect(notifier.state.completed, true);
      expect(notifier.state.groupProgress, _groupProgress);
    });

    test('sets error on failure', () async {
      when(() => mockRepo.getVotingSession('s1'))
          .thenThrow(AppException('Network error'));

      final notifier = VotingNotifier(mockRepo, 's1');

      await notifier.loadSession();

      expect(notifier.state.loading, false);
      expect(notifier.state.error, 'Network error');
    });

    test('resumes from partially answered session', () async {
      final session = _makeSession(answeredCount: 1, totalQuestions: 3);
      when(() => mockRepo.getVotingSession('s1'))
          .thenAnswer((_) async => session);

      final notifier = VotingNotifier(mockRepo, 's1');

      await notifier.loadSession();

      expect(notifier.state.unansweredQuestions.length, 2);
      expect(notifier.state.currentQuestion?.questionId, 'q1');
    });
  });

  group('VotingNotifier.selectTarget', () {
    test('casts vote and advances to next question', () async {
      final session = _makeSession();
      when(() => mockRepo.getVotingSession('s1'))
          .thenAnswer((_) async => session);
      when(() => mockRepo.castVote(
            seasonId: 's1',
            questionId: 'q0',
            targetId: 'u1',
          )).thenAnswer((_) async => const VoteResultData(
            vote: VoteInfo(questionId: 'q0', targetId: 'u1'),
            progress: VotingProgress(answered: 1, total: 3),
          ));

      final notifier = VotingNotifier(mockRepo, 's1');
      await notifier.loadSession();
      await notifier.selectTarget('u1');

      expect(notifier.state.currentIndex, 1);
      expect(notifier.state.submitting, false);
      expect(notifier.state.selectedTargetId, isNull);
    });

    test('completes voting on last question', () async {
      final session = _makeSession(answeredCount: 2, totalQuestions: 3);
      when(() => mockRepo.getVotingSession('s1'))
          .thenAnswer((_) async => session);
      when(() => mockRepo.castVote(
            seasonId: 's1',
            questionId: 'q2',
            targetId: 'u1',
          )).thenAnswer((_) async => const VoteResultData(
            vote: VoteInfo(questionId: 'q2', targetId: 'u1'),
            progress: VotingProgress(answered: 3, total: 3),
          ));
      when(() => mockRepo.getVotingProgress('s1'))
          .thenAnswer((_) async => _groupProgress);

      final notifier = VotingNotifier(mockRepo, 's1');
      await notifier.loadSession();

      expect(notifier.state.unansweredQuestions.length, 1);

      await notifier.selectTarget('u1');

      expect(notifier.state.completed, true);
      expect(notifier.state.groupProgress, _groupProgress);
    });

    test('handles vote error', () async {
      final session = _makeSession();
      when(() => mockRepo.getVotingSession('s1'))
          .thenAnswer((_) async => session);
      when(() => mockRepo.castVote(
            seasonId: 's1',
            questionId: 'q0',
            targetId: 'u1',
          )).thenThrow(AppException('Already voted'));

      final notifier = VotingNotifier(mockRepo, 's1');
      await notifier.loadSession();
      await notifier.selectTarget('u1');

      expect(notifier.state.error, 'Already voted');
      expect(notifier.state.submitting, false);
      expect(notifier.state.selectedTargetId, isNull);
    });

    test('ignores tap while submitting', () async {
      final session = _makeSession();
      when(() => mockRepo.getVotingSession('s1'))
          .thenAnswer((_) async => session);

      final notifier = VotingNotifier(mockRepo, 's1');
      await notifier.loadSession();

      // Simulate submitting state
      when(() => mockRepo.castVote(
            seasonId: 's1',
            questionId: 'q0',
            targetId: 'u1',
          )).thenAnswer((_) async {
        // During this, try selecting another target
        await notifier.selectTarget('u2');
        return const VoteResultData(
          vote: VoteInfo(questionId: 'q0', targetId: 'u1'),
          progress: VotingProgress(answered: 1, total: 3),
        );
      });

      await notifier.selectTarget('u1');

      // u2 should not have been called
      verifyNever(() => mockRepo.castVote(
            seasonId: 's1',
            questionId: 'q0',
            targetId: 'u2',
          ));
    });
  });

  group('VotingState', () {
    test('currentQuestion returns null when no session', () {
      const state = VotingState();
      expect(state.currentQuestion, isNull);
    });

    test('totalQuestions returns 0 when no session', () {
      const state = VotingState();
      expect(state.totalQuestions, 0);
    });

    test('copyWith preserves existing values', () {
      final session = _makeSession();
      final state = VotingState(session: session, currentIndex: 2);
      final newState = state.copyWith(loading: true);

      expect(newState.loading, true);
      expect(newState.session, session);
      expect(newState.currentIndex, 2);
    });

    test('clearError removes error', () {
      const state = VotingState(error: 'some error');
      final newState = state.copyWith(clearError: true);
      expect(newState.error, isNull);
    });

    test('clearSelection removes selectedTargetId', () {
      const state = VotingState(selectedTargetId: 'u1');
      final newState = state.copyWith(clearSelection: true);
      expect(newState.selectedTargetId, isNull);
    });
  });
}
