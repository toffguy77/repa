import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/providers/api_provider.dart';
import '../data/reveal_repository.dart';
import '../domain/reveal.dart';

enum RevealPhase { loading, waiting, ready, opening, revealed }

class RevealState {
  final RevealPhase phase;
  final RevealData? data;
  final List<MemberCard>? membersCards;
  final DetectorResult? detector;
  final String? cardImageUrl;
  final String? error;
  final bool unlockingHidden;
  final bool buyingDetector;

  /// What is known about votes concerning this member before the Reveal. Null until loaded.
  final AnticipationState? anticipation;
  // Reactions keyed by target user ID
  final Map<String, ReactionCounts> reactions;

  const RevealState({
    this.phase = RevealPhase.loading,
    this.data,
    this.membersCards,
    this.detector,
    this.cardImageUrl,
    this.error,
    this.unlockingHidden = false,
    this.buyingDetector = false,
    this.anticipation,
    this.reactions = const {},
  });

  RevealState copyWith({
    RevealPhase? phase,
    RevealData? data,
    List<MemberCard>? membersCards,
    DetectorResult? detector,
    String? cardImageUrl,
    String? error,
    bool? unlockingHidden,
    bool? buyingDetector,
    AnticipationState? anticipation,
    Map<String, ReactionCounts>? reactions,
  }) {
    return RevealState(
      phase: phase ?? this.phase,
      data: data ?? this.data,
      membersCards: membersCards ?? this.membersCards,
      detector: detector ?? this.detector,
      cardImageUrl: cardImageUrl ?? this.cardImageUrl,
      error: error,
      unlockingHidden: unlockingHidden ?? this.unlockingHidden,
      buyingDetector: buyingDetector ?? this.buyingDetector,
      anticipation: anticipation ?? this.anticipation,
      reactions: reactions ?? this.reactions,
    );
  }
}

class RevealNotifier extends StateNotifier<RevealState> {
  final RevealRepository _repo;
  final String seasonId;
  final String seasonStatus;

  RevealNotifier(this._repo, this.seasonId, this.seasonStatus)
      : super(const RevealState());

  Future<void> load() async {
    state = state.copyWith(phase: RevealPhase.loading, error: null);

    if (seasonStatus != 'REVEALED') {
      state = state.copyWith(phase: RevealPhase.waiting);
      // The waiting screen is where the mid-week pushes land, so it needs something to say.
      await loadAnticipation();
      return;
    }

    try {
      final data = await _repo.getReveal(seasonId);
      state = state.copyWith(
        phase: RevealPhase.ready,
        data: data,
      );

      // Fetch card URL in background
      _loadCardUrl();
    } on AppException catch (e) {
      state = state.copyWith(
        phase: RevealPhase.waiting,
        error: e.message,
      );
    } catch (_) {
      state = state.copyWith(
        phase: RevealPhase.waiting,
        error: 'Не удалось загрузить данные',
      );
    }
  }

  void startOpening() {
    state = state.copyWith(phase: RevealPhase.opening);
  }

  void finishOpening() {
    state = state.copyWith(phase: RevealPhase.revealed);
  }

  Future<void> _loadCardUrl() async {
    try {
      final result = await _repo.getMyCardUrl(seasonId);
      if (result.imageUrl != null) {
        state = state.copyWith(cardImageUrl: result.imageUrl);
      }
    } catch (_) {}
  }

  Future<void> openHidden() async {
    if (state.unlockingHidden) return;
    state = state.copyWith(unlockingHidden: true, error: null);

    try {
      final data = await _repo.openHidden(seasonId);
      state = state.copyWith(data: data, unlockingHidden: false);
    } on AppException catch (e) {
      state = state.copyWith(
        unlockingHidden: false,
        error: e.message,
      );
    } catch (_) {
      state = state.copyWith(
        unlockingHidden: false,
        error: 'Не удалось открыть атрибуты',
      );
    }
  }

  /// Loads the mid-week signal. A failure leaves the waiting screen as it was: the anticipation
  /// state is an enrichment, not a prerequisite for the screen.
  Future<void> loadAnticipation() async {
    try {
      final state = await _repo.getAnticipation(seasonId);
      this.state = this.state.copyWith(anticipation: state);
    } catch (_) {}
  }

  Future<void> loadDetector() async {
    try {
      final result = await _repo.getDetector(seasonId);
      state = state.copyWith(detector: result);
    } catch (_) {}
  }

  Future<void> buyDetector() async {
    if (state.buyingDetector) return;
    state = state.copyWith(buyingDetector: true, error: null);

    try {
      final result = await _repo.buyDetector(seasonId);
      state = state.copyWith(detector: result, buyingDetector: false);
    } on AppException catch (e) {
      state = state.copyWith(
        buyingDetector: false,
        error: e.message,
      );
    } catch (_) {
      state = state.copyWith(
        buyingDetector: false,
        error: 'Не удалось купить детектор',
      );
    }
  }

  /// Buys the middle rung: one voter revealed partially.
  Future<void> buyDetectorHint() async {
    if (state.buyingDetector) return;
    state = state.copyWith(buyingDetector: true, error: null);

    try {
      final result = await _repo.buyDetectorHint(seasonId);
      state = state.copyWith(detector: result, buyingDetector: false);
    } on AppException catch (e) {
      state = state.copyWith(buyingDetector: false, error: e.message);
    } catch (_) {
      state = state.copyWith(
        buyingDetector: false,
        error: 'Не удалось получить подсказку',
      );
    }
  }

  Future<void> loadMembersCards() async {
    try {
      final cards = await _repo.getMembersCards(seasonId);
      state = state.copyWith(membersCards: cards);
    } catch (_) {}
  }

  Future<void> loadReactions(String targetId) async {
    try {
      final result = await _repo.getReactions(seasonId, targetId);
      final updated = Map<String, ReactionCounts>.from(state.reactions);
      updated[targetId] = result;
      state = state.copyWith(reactions: updated);
    } catch (_) {}
  }

  Future<void> sendReaction(String targetId, String emoji) async {
    // Optimistic update
    final current = state.reactions[targetId];
    final oldCounts = Map<String, int>.from(current?.counts ?? {});
    final oldEmoji = current?.myEmoji;

    final newCounts = Map<String, int>.from(oldCounts);
    if (oldEmoji != null) {
      newCounts[oldEmoji] = (newCounts[oldEmoji] ?? 1) - 1;
      if (newCounts[oldEmoji]! <= 0) newCounts.remove(oldEmoji);
    }
    newCounts[emoji] = (newCounts[emoji] ?? 0) + 1;

    final updated = Map<String, ReactionCounts>.from(state.reactions);
    updated[targetId] = ReactionCounts(counts: newCounts, myEmoji: emoji);
    state = state.copyWith(reactions: updated);

    try {
      final result = await _repo.createReaction(seasonId, targetId, emoji);
      final refreshed = Map<String, ReactionCounts>.from(state.reactions);
      refreshed[targetId] = result;
      state = state.copyWith(reactions: refreshed);
    } catch (_) {
      // Rollback
      final rollback = Map<String, ReactionCounts>.from(state.reactions);
      rollback[targetId] = ReactionCounts(counts: oldCounts, myEmoji: oldEmoji);
      state = state.copyWith(reactions: rollback);
    }
  }

  void clearError() {
    state = state.copyWith(error: null);
  }
}

/// The reveal repository, exposed so screens can report a share without going through the
/// notifier — a share is not part of the reveal's state.
final revealRepositoryProvider = Provider<RevealRepository>(
  (ref) => RevealRepository(ref.watch(apiServiceProvider)),
);

final revealProvider = StateNotifierProvider.autoDispose
    .family<RevealNotifier, RevealState, ({String seasonId, String status})>(
  (ref, args) {
    final api = ref.watch(apiServiceProvider);
    final repo = RevealRepository(api);
    return RevealNotifier(repo, args.seasonId, args.status);
  },
);
