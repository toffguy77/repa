import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import '../../../core/api/api_client.dart';
import '../../../core/providers/api_provider.dart';
import '../../../core/providers/auth_provider.dart';
import '../../auth/domain/user.dart';

class PushPref {
  final String category;
  final String label;
  final bool enabled;

  const PushPref({required this.category, required this.label, this.enabled = true});

  PushPref copyWith({bool? enabled}) =>
      PushPref(category: category, label: label, enabled: enabled ?? this.enabled);
}

class SettingsState {
  final User? user;
  final List<PushPref> pushPrefs;
  final bool savingAvatar;
  final String? avatarError;
  final bool deletingAccount;

  const SettingsState({
    this.user,
    this.pushPrefs = const [],
    this.savingAvatar = false,
    this.avatarError,
    this.deletingAccount = false,
  });

  SettingsState copyWith({
    User? user,
    List<PushPref>? pushPrefs,
    bool? savingAvatar,
    String? avatarError,
    bool? deletingAccount,
  }) {
    return SettingsState(
      user: user ?? this.user,
      pushPrefs: pushPrefs ?? this.pushPrefs,
      savingAvatar: savingAvatar ?? this.savingAvatar,
      avatarError: avatarError,
      deletingAccount: deletingAccount ?? this.deletingAccount,
    );
  }
}

class SettingsNotifier extends StateNotifier<SettingsState> {
  final Ref _ref;

  SettingsNotifier(this._ref) : super(const SettingsState()) {
    _init();
  }

  void _init() {
    final authState = _ref.read(authProvider);
    state = state.copyWith(
      user: authState.user,
      pushPrefs: [
        PushPref(category: 'SEASON_START', label: 'Старт голосования'),
        PushPref(category: 'REMINDER', label: 'Напоминание проголосовать'),
        PushPref(category: 'REVEAL', label: 'Reveal готов'),
        PushPref(category: 'REACTION', label: 'Реакции на карточку'),
        PushPref(category: 'NEXT_SEASON', label: 'Следующий сезон'),
      ],
    );
  }

  Future<void> pickAndUploadAvatar(ImageSource source) async {
    final picker = ImagePicker();
    final image = await picker.pickImage(
      source: source,
      maxWidth: 512,
      maxHeight: 512,
      imageQuality: 85,
    );
    if (image == null) return;

    state = state.copyWith(savingAvatar: true, avatarError: null);
    try {
      final api = _ref.read(apiServiceProvider);
      final file = await MultipartFile.fromFile(
        image.path,
        filename: 'avatar.jpg',
      );
      final response = await api.uploadAvatar(file);
      final data = response['data'] as Map<String, dynamic>;
      final avatarUrl = data['avatar_url'] as String?;
      if (avatarUrl != null && state.user != null) {
        final updated = state.user!.copyWith(avatarUrl: avatarUrl);
        state = state.copyWith(user: updated, savingAvatar: false);
      } else {
        state = state.copyWith(savingAvatar: false);
      }
    } on DioException catch (e) {
      state = state.copyWith(
        savingAvatar: false,
        avatarError: parseError(e).message,
      );
    }
  }

  List<PushPref> _updatePrefAt(int index, bool enabled) {
    return [
      for (int i = 0; i < state.pushPrefs.length; i++)
        if (i == index) state.pushPrefs[i].copyWith(enabled: enabled)
        else state.pushPrefs[i],
    ];
  }

  Future<void> togglePushPref(int index, bool enabled) async {
    state = state.copyWith(pushPrefs: _updatePrefAt(index, enabled));

    try {
      final api = _ref.read(apiServiceProvider);
      await api.updatePushPreferences(state.pushPrefs[index].category, enabled);
    } on DioException {
      // Revert on failure
      state = state.copyWith(pushPrefs: _updatePrefAt(index, !enabled));
    }
  }

  Future<void> logout() async {
    await _ref.read(authProvider.notifier).logout();
  }

  Future<bool> deleteAccount() async {
    state = state.copyWith(deletingAccount: true);
    try {
      final api = _ref.read(apiServiceProvider);
      await api.deleteAccount();
      await _ref.read(authProvider.notifier).logout();
      return true;
    } on DioException {
      state = state.copyWith(deletingAccount: false);
      return false;
    }
  }
}

final settingsProvider =
    StateNotifierProvider.autoDispose<SettingsNotifier, SettingsState>((ref) {
  return SettingsNotifier(ref);
});
