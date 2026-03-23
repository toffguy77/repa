import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/providers/api_provider.dart';
import '../../groups/data/groups_repository.dart';
import '../../groups/presentation/groups_notifier.dart';
import '../data/telegram_repository.dart';
import '../domain/telegram_connect.dart';

class TelegramSetupState {
  final bool loading;
  final bool disconnecting;
  final bool sharing;
  final String? error;
  final TelegramConnectCode? connectCode;
  final bool connected;
  final String? chatUsername;

  const TelegramSetupState({
    this.loading = false,
    this.disconnecting = false,
    this.sharing = false,
    this.error,
    this.connectCode,
    this.connected = false,
    this.chatUsername,
  });

  TelegramSetupState copyWith({
    bool? loading,
    bool? disconnecting,
    bool? sharing,
    String? error,
    TelegramConnectCode? connectCode,
    bool? connected,
    String? chatUsername,
  }) {
    return TelegramSetupState(
      loading: loading ?? this.loading,
      disconnecting: disconnecting ?? this.disconnecting,
      sharing: sharing ?? this.sharing,
      error: error,
      connectCode: connectCode ?? this.connectCode,
      connected: connected ?? this.connected,
      chatUsername: chatUsername ?? this.chatUsername,
    );
  }
}

class TelegramSetupNotifier extends StateNotifier<TelegramSetupState> {
  final TelegramRepository _telegramRepo;
  final GroupsRepository _groupsRepo;
  final String groupId;

  TelegramSetupNotifier(this._telegramRepo, this._groupsRepo, this.groupId)
      : super(const TelegramSetupState());

  void init({String? telegramUsername}) {
    if (telegramUsername != null) {
      state = TelegramSetupState(
        connected: true,
        chatUsername: telegramUsername,
      );
    }
  }

  Future<void> generateCode() async {
    state = state.copyWith(loading: true, error: null);
    try {
      final code = await _telegramRepo.generateCode(groupId);
      state = state.copyWith(loading: false, connectCode: code);
    } on AppException catch (e) {
      state = state.copyWith(loading: false, error: e.message);
    }
  }

  Future<bool> verifyConnection() async {
    state = state.copyWith(loading: true, error: null);
    try {
      final detail = await _groupsRepo.getGroup(groupId);
      if (detail.group.telegramUsername != null) {
        state = TelegramSetupState(
          connected: true,
          chatUsername: detail.group.telegramUsername,
        );
        return true;
      } else {
        state = state.copyWith(
          loading: false,
          error: 'Бот ещё не подключён. Проверьте, что выполнили все шаги.',
        );
        return false;
      }
    } on AppException catch (e) {
      state = state.copyWith(loading: false, error: e.message);
      return false;
    }
  }

  Future<void> disconnect() async {
    state = state.copyWith(disconnecting: true, error: null);
    try {
      await _telegramRepo.disconnect(groupId);
      state = const TelegramSetupState();
    } on AppException catch (e) {
      state = state.copyWith(disconnecting: false, error: e.message);
    }
  }

  void clearError() {
    state = state.copyWith(error: null);
  }

}

final telegramSetupProvider = StateNotifierProvider.autoDispose
    .family<TelegramSetupNotifier, TelegramSetupState, String>(
  (ref, groupId) {
    final api = ref.watch(apiServiceProvider);
    final telegramRepo = TelegramRepository(api);
    final groupsRepo = ref.watch(groupsRepositoryProvider);
    return TelegramSetupNotifier(telegramRepo, groupsRepo, groupId);
  },
);

// Standalone provider for share-to-telegram action
final shareToTelegramProvider = Provider.autoDispose<TelegramRepository>((ref) {
  final api = ref.watch(apiServiceProvider);
  return TelegramRepository(api);
});
