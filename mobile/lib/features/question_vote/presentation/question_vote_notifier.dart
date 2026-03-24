import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/providers/api_provider.dart';
import '../data/question_vote_repository.dart';
import '../domain/question_candidate.dart';

enum QuestionVoteStatus { loading, voting, voted, unavailable, error }

class QuestionVoteState {
  final QuestionVoteStatus status;
  final List<QuestionCandidate> candidates;
  final String? selectedId;
  final String? error;

  const QuestionVoteState({
    this.status = QuestionVoteStatus.loading,
    this.candidates = const [],
    this.selectedId,
    this.error,
  });
}

class QuestionVoteNotifier extends StateNotifier<QuestionVoteState> {
  final QuestionVoteRepository _repo;
  final String groupId;

  QuestionVoteNotifier(this._repo, this.groupId)
      : super(const QuestionVoteState());

  static bool isVotingWindowOpen() {
    final now = DateTime.now().toUtc();
    final weekday = now.weekday; // 1=Mon, 7=Sun
    // Sunday 09:00 UTC (12:00 MSK) to Monday 14:00 UTC (17:00 MSK)
    if (weekday == 7 && now.hour >= 9) return true;
    if (weekday == 1 && now.hour < 14) return true;
    return false;
  }

  Future<void> load() async {
    if (!isVotingWindowOpen()) {
      state = const QuestionVoteState(status: QuestionVoteStatus.unavailable);
      return;
    }

    state = const QuestionVoteState(status: QuestionVoteStatus.loading);
    try {
      final candidates = await _repo.getCandidates(groupId);
      state = QuestionVoteState(
        status: QuestionVoteStatus.voting,
        candidates: candidates,
      );
    } on AppException catch (e) {
      if (e.code == 'ALREADY_VOTED') {
        state = const QuestionVoteState(status: QuestionVoteStatus.voted);
      } else {
        state = QuestionVoteState(
          status: QuestionVoteStatus.error,
          error: e.message,
        );
      }
    }
  }

  Future<void> vote(String questionId) async {
    try {
      await _repo.vote(groupId, questionId);
      state = QuestionVoteState(
        status: QuestionVoteStatus.voted,
        candidates: state.candidates,
        selectedId: questionId,
      );
    } on AppException catch (e) {
      state = QuestionVoteState(
        status: state.status,
        candidates: state.candidates,
        error: e.message,
      );
    }
  }
}

final questionVoteRepositoryProvider = Provider<QuestionVoteRepository>((ref) {
  final api = ref.watch(apiServiceProvider);
  return QuestionVoteRepository(api);
});

final questionVoteProvider = StateNotifierProvider.family<
    QuestionVoteNotifier, QuestionVoteState, String>((ref, groupId) {
  final repo = ref.watch(questionVoteRepositoryProvider);
  return QuestionVoteNotifier(repo, groupId);
});
