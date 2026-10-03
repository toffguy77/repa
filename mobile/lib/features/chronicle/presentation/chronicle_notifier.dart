import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/providers/api_provider.dart';
import '../data/chronicle_repository.dart';
import '../domain/chronicle.dart';

final chronicleRepositoryProvider = Provider<ChronicleRepository>((ref) {
  return ChronicleRepository(ref.watch(apiServiceProvider));
});

class ChronicleState {
  final bool loading;
  final String? error;
  final Chronicle? chronicle;

  const ChronicleState({this.loading = false, this.error, this.chronicle});
}

class ChronicleNotifier extends StateNotifier<ChronicleState> {
  final ChronicleRepository _repo;
  final String groupId;

  ChronicleNotifier(this._repo, this.groupId) : super(const ChronicleState());

  Future<void> load() async {
    state = ChronicleState(loading: true, chronicle: state.chronicle);
    try {
      state = ChronicleState(chronicle: await _repo.getChronicle(groupId));
    } on AppException catch (e) {
      state = ChronicleState(error: e.message, chronicle: state.chronicle);
    }
  }
}

final chronicleProvider = StateNotifierProvider.autoDispose
    .family<ChronicleNotifier, ChronicleState, String>((ref, groupId) {
  return ChronicleNotifier(ref.watch(chronicleRepositoryProvider), groupId);
});
