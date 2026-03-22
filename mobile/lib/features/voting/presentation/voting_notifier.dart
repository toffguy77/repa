import 'dart:math';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/providers/api_provider.dart';
import '../data/voting_repository.dart';
import '../domain/voting.dart';

final votingRepositoryProvider = Provider<VotingRepository>((ref) {
  final api = ref.watch(apiServiceProvider);
  return VotingRepository(api);
});

class VotingState {
  final bool loading;
  final String? error;
  final VotingSession? session;
  final int currentIndex;
  final String? selectedTargetId;
  final bool submitting;
  final bool completed;
  final GroupVotingProgress? groupProgress;

  const VotingState({
    this.loading = false,
    this.error,
    this.session,
    this.currentIndex = 0,
    this.selectedTargetId,
    this.submitting = false,
    this.completed = false,
    this.groupProgress,
  });

  List<VotingQuestion> get unansweredQuestions =>
      session?.questions.where((q) => !q.answered).toList() ?? [];

  VotingQuestion? get currentQuestion {
    final questions = unansweredQuestions;
    if (currentIndex >= questions.length) return null;
    return questions[currentIndex];
  }

  int get totalQuestions => session?.questions.length ?? 0;
  int get answeredQuestions =>
      (session?.progress.answered ?? 0) + currentIndex;

  VotingState copyWith({
    bool? loading,
    String? error,
    VotingSession? session,
    int? currentIndex,
    String? selectedTargetId,
    bool? submitting,
    bool? completed,
    GroupVotingProgress? groupProgress,
    bool clearError = false,
    bool clearSelection = false,
  }) {
    return VotingState(
      loading: loading ?? this.loading,
      error: clearError ? null : (error ?? this.error),
      session: session ?? this.session,
      currentIndex: currentIndex ?? this.currentIndex,
      selectedTargetId:
          clearSelection ? null : (selectedTargetId ?? this.selectedTargetId),
      submitting: submitting ?? this.submitting,
      completed: completed ?? this.completed,
      groupProgress: groupProgress ?? this.groupProgress,
    );
  }
}

class VotingNotifier extends StateNotifier<VotingState> {
  final VotingRepository _repo;
  final String seasonId;
  final Random _random = Random();

  List<VotingTarget> _shuffledTargets = [];

  VotingNotifier(this._repo, this.seasonId) : super(const VotingState());

  List<VotingTarget> get shuffledTargets => _shuffledTargets;

  Future<void> loadSession() async {
    state = state.copyWith(loading: true, clearError: true);
    try {
      final session = await _repo.getVotingSession(seasonId);
      final unanswered = session.questions.where((q) => !q.answered).toList();

      if (unanswered.isEmpty) {
        final progress = await _repo.getVotingProgress(seasonId);
        state = state.copyWith(
          loading: false,
          session: session,
          completed: true,
          groupProgress: progress,
        );
        return;
      }

      _shuffleTargets(session.targets);
      state = state.copyWith(
        loading: false,
        session: session,
        currentIndex: 0,
      );
    } on AppException catch (e) {
      state = state.copyWith(loading: false, error: e.message);
    }
  }

  void _shuffleTargets(List<VotingTarget> targets) {
    _shuffledTargets = List.of(targets)..shuffle(_random);
  }

  Future<void> selectTarget(String targetId) async {
    if (state.submitting || state.selectedTargetId != null) return;

    final question = state.currentQuestion;
    if (question == null) return;

    state = state.copyWith(selectedTargetId: targetId, submitting: true);

    try {
      // Fire API call and 400ms delay in parallel so the user sees the
      // selection animation for at least 400ms (spec requirement).
      final results = await Future.wait([
        _repo.castVote(
          seasonId: seasonId,
          questionId: question.questionId,
          targetId: targetId,
        ),
        Future.delayed(const Duration(milliseconds: 400)),
      ]);
      final _ = results[0] as VoteResultData;

      final unanswered = state.unansweredQuestions;
      final nextIndex = state.currentIndex + 1;

      if (nextIndex >= unanswered.length) {
        final progress = await _repo.getVotingProgress(seasonId);
        state = state.copyWith(
          submitting: false,
          completed: true,
          groupProgress: progress,
        );
      } else {
        _shuffleTargets(state.session!.targets);
        state = state.copyWith(
          submitting: false,
          currentIndex: nextIndex,
          clearSelection: true,
        );
      }
    } on AppException catch (e) {
      state = state.copyWith(
        submitting: false,
        error: e.message,
        clearSelection: true,
      );
    }
  }

  void clearError() {
    state = state.copyWith(clearError: true);
  }
}

final votingProvider = StateNotifierProvider.autoDispose
    .family<VotingNotifier, VotingState, String>((ref, seasonId) {
  return VotingNotifier(ref.watch(votingRepositoryProvider), seasonId);
});

/// Standalone provider for the complete screen — fetches progress
/// independently so it doesn't depend on the autoDispose votingProvider.
final groupVotingProgressProvider = FutureProvider.autoDispose
    .family<GroupVotingProgress, String>((ref, seasonId) {
  final repo = ref.watch(votingRepositoryProvider);
  return repo.getVotingProgress(seasonId);
});
