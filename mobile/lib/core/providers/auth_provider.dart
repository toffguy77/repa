import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../features/auth/domain/user.dart';
import '../api/api_client.dart';
import 'api_provider.dart';

enum AuthStatus { unknown, authenticated, unauthenticated }

class AuthState {
  final AuthStatus status;
  final User? user;
  final bool needsProfileSetup;
  final bool isNewUser;

  const AuthState({
    this.status = AuthStatus.unknown,
    this.user,
    this.needsProfileSetup = false,
    this.isNewUser = false,
  });

  AuthState copyWith({
    AuthStatus? status,
    User? user,
    bool? needsProfileSetup,
    bool? isNewUser,
  }) {
    return AuthState(
      status: status ?? this.status,
      user: user ?? this.user,
      needsProfileSetup: needsProfileSetup ?? this.needsProfileSetup,
      isNewUser: isNewUser ?? this.isNewUser,
    );
  }
}

bool _needsSetup(User user) {
  return user.avatarEmoji == null || user.birthYear == null;
}

class AuthNotifier extends StateNotifier<AuthState> {
  final Ref _ref;

  AuthNotifier(this._ref) : super(const AuthState());

  Future<void> checkAuth() async {
    final storage = _ref.read(secureStorageProvider);
    final token = await storage.read(key: tokenKey);
    if (token == null) {
      state = const AuthState(status: AuthStatus.unauthenticated);
      return;
    }
    try {
      final dio = _ref.read(dioProvider);
      final response = await dio.get('/auth/me');
      final data = response.data['data'] as Map<String, dynamic>;
      final user = User.fromJson(data);
      state = AuthState(
        status: AuthStatus.authenticated,
        user: user,
        needsProfileSetup: _needsSetup(user),
      );
    } catch (_) {
      await storage.delete(key: tokenKey);
      state = const AuthState(status: AuthStatus.unauthenticated);
    }
  }

  Future<void> login(String token, User user) async {
    final storage = _ref.read(secureStorageProvider);
    await storage.write(key: tokenKey, value: token);
    state = AuthState(
      status: AuthStatus.authenticated,
      user: user,
      needsProfileSetup: _needsSetup(user),
    );
  }

  void profileCompleted(User user) {
    state = AuthState(
      status: AuthStatus.authenticated,
      user: user,
      needsProfileSetup: false,
      isNewUser: true,
    );
  }

  void onboardingCompleted() {
    state = state.copyWith(isNewUser: false);
  }

  Future<void> logout() async {
    final storage = _ref.read(secureStorageProvider);
    await storage.delete(key: tokenKey);
    state = const AuthState(status: AuthStatus.unauthenticated);
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref);
});
