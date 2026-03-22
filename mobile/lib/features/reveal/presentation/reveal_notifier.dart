import 'package:flutter_riverpod/flutter_riverpod.dart';
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

  const RevealState({
    this.phase = RevealPhase.loading,
    this.data,
    this.membersCards,
    this.detector,
    this.cardImageUrl,
    this.error,
    this.unlockingHidden = false,
    this.buyingDetector = false,
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
    } catch (e) {
      state = state.copyWith(
        phase: RevealPhase.waiting,
        error: e.toString(),
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
    } catch (e) {
      state = state.copyWith(
        unlockingHidden: false,
        error: e.toString(),
      );
    }
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
    } catch (e) {
      state = state.copyWith(
        buyingDetector: false,
        error: e.toString(),
      );
    }
  }

  Future<void> loadMembersCards() async {
    try {
      final cards = await _repo.getMembersCards(seasonId);
      state = state.copyWith(membersCards: cards);
    } catch (_) {}
  }

  void clearError() {
    state = state.copyWith(error: null);
  }
}

final revealProvider = StateNotifierProvider.autoDispose
    .family<RevealNotifier, RevealState, ({String seasonId, String status})>(
  (ref, args) {
    final api = ref.watch(apiServiceProvider);
    final repo = RevealRepository(api);
    return RevealNotifier(repo, args.seasonId, args.status);
  },
);
