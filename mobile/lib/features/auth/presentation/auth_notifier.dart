import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/providers/api_provider.dart';
import '../../../core/providers/auth_provider.dart';
import '../data/auth_repository.dart';

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  final api = ref.watch(apiServiceProvider);
  return AuthRepository(api);
});

// --- OTP Send ---

class OtpSendState {
  final bool loading;
  final String? error;

  const OtpSendState({this.loading = false, this.error});
}

class OtpSendNotifier extends StateNotifier<OtpSendState> {
  final AuthRepository _repo;

  OtpSendNotifier(this._repo) : super(const OtpSendState());

  Future<bool> send(String phone) async {
    state = const OtpSendState(loading: true);
    try {
      await _repo.sendOtp(phone);
      state = const OtpSendState();
      return true;
    } on AppException catch (e) {
      state = OtpSendState(error: e.message);
      return false;
    }
  }
}

final otpSendProvider =
    StateNotifierProvider<OtpSendNotifier, OtpSendState>((ref) {
  return OtpSendNotifier(ref.watch(authRepositoryProvider));
});

// --- OTP Verify ---

class OtpVerifyState {
  final bool loading;
  final String? error;

  const OtpVerifyState({this.loading = false, this.error});
}

class OtpVerifyNotifier extends StateNotifier<OtpVerifyState> {
  final AuthRepository _repo;
  final AuthNotifier _authNotifier;

  OtpVerifyNotifier(this._repo, this._authNotifier)
      : super(const OtpVerifyState());

  void reset() {
    state = const OtpVerifyState();
  }

  Future<bool> verify(String phone, String code) async {
    state = const OtpVerifyState(loading: true);
    try {
      final result = await _repo.verifyOtp(phone, code);
      await _authNotifier.login(result.token, result.user);
      state = const OtpVerifyState();
      return true;
    } on AppException catch (e) {
      state = OtpVerifyState(error: e.message);
      return false;
    }
  }
}

final otpVerifyProvider =
    StateNotifierProvider<OtpVerifyNotifier, OtpVerifyState>((ref) {
  return OtpVerifyNotifier(
    ref.watch(authRepositoryProvider),
    ref.watch(authProvider.notifier),
  );
});

// --- Profile Setup ---

class ProfileSetupState {
  final bool loading;
  final String? error;
  final bool? usernameAvailable;
  final bool checkingUsername;

  const ProfileSetupState({
    this.loading = false,
    this.error,
    this.usernameAvailable,
    this.checkingUsername = false,
  });
}

class ProfileSetupNotifier extends StateNotifier<ProfileSetupState> {
  final AuthRepository _repo;
  final AuthNotifier _authNotifier;

  ProfileSetupNotifier(this._repo, this._authNotifier)
      : super(const ProfileSetupState());

  Future<void> checkUsername(String username) async {
    if (username.length < 3) {
      state = const ProfileSetupState(usernameAvailable: null);
      return;
    }
    state = const ProfileSetupState(checkingUsername: true);
    try {
      final available = await _repo.checkUsername(username);
      state = ProfileSetupState(usernameAvailable: available);
    } on AppException {
      state = const ProfileSetupState(usernameAvailable: null);
    }
  }

  Future<bool> submit({
    required String username,
    required String avatarEmoji,
    required int birthYear,
  }) async {
    state = const ProfileSetupState(loading: true);
    try {
      final user = await _repo.updateProfile(
        username: username,
        avatarEmoji: avatarEmoji,
        birthYear: birthYear,
      );
      _authNotifier.profileCompleted(user);
      state = const ProfileSetupState();
      return true;
    } on AppException catch (e) {
      state = ProfileSetupState(error: e.message);
      return false;
    }
  }
}

final profileSetupProvider =
    StateNotifierProvider<ProfileSetupNotifier, ProfileSetupState>((ref) {
  return ProfileSetupNotifier(
    ref.watch(authRepositoryProvider),
    ref.watch(authProvider.notifier),
  );
});
