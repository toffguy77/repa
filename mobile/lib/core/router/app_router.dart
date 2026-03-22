import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:go_router/go_router.dart';
import '../providers/auth_provider.dart';
import '../providers/api_provider.dart';
import '../../features/auth/presentation/phone_screen.dart';
import '../../features/auth/presentation/otp_screen.dart';
import '../../features/auth/presentation/profile_setup_screen.dart';
import '../../features/home/home_screen.dart';
import '../../features/groups/presentation/create_group_screen.dart';
import '../../features/groups/presentation/join_group_screen.dart';
import '../../features/groups/presentation/group_screen.dart';

const _pendingInviteCodeKey = 'pending_invite_code';

class _RouterNotifier extends ChangeNotifier {
  final FlutterSecureStorage _storage;
  AuthState _authState = const AuthState();

  _RouterNotifier(this._storage);

  void update(AuthState authState) {
    if (_authState.status != authState.status ||
        _authState.needsProfileSetup != authState.needsProfileSetup) {
      _authState = authState;
      notifyListeners();
    }
    _authState = authState;
  }

  String? redirect(GoRouterState state) {
    final isAuth = _authState.status == AuthStatus.authenticated;
    final isAuthRoute = state.matchedLocation.startsWith('/auth');
    final isDeeplink = state.matchedLocation.startsWith('/join/');

    if (_authState.status == AuthStatus.unknown) return null;

    if (!isAuth) {
      if (isDeeplink) {
        final code = state.pathParameters['code'] ?? '';
        if (code.isNotEmpty) {
          _storage.write(key: _pendingInviteCodeKey, value: code);
        }
      }
      if (!isAuthRoute) return '/auth/phone';
      return null;
    }

    if (isAuth && _authState.needsProfileSetup) {
      if (state.matchedLocation != '/auth/setup') return '/auth/setup';
      return null;
    }

    if (isAuth && isAuthRoute) return '/home';

    return null;
  }
}

final _routerNotifierProvider =
    ChangeNotifierProvider<_RouterNotifier>((ref) {
  final storage = ref.watch(secureStorageProvider);
  final notifier = _RouterNotifier(storage);
  ref.listen(authProvider, (_, next) {
    notifier.update(next);
  });
  return notifier;
});

final routerProvider = Provider<GoRouter>((ref) {
  final notifier = ref.watch(_routerNotifierProvider);
  final storage = ref.watch(secureStorageProvider);

  return GoRouter(
    initialLocation: '/auth/phone',
    refreshListenable: notifier,
    redirect: (context, state) => notifier.redirect(state),
    routes: [
      GoRoute(
        path: '/auth/phone',
        builder: (context, state) => const PhoneScreen(),
      ),
      GoRoute(
        path: '/auth/otp',
        builder: (context, state) {
          final phone = state.extra as String? ?? '';
          return OtpScreen(phone: phone);
        },
      ),
      GoRoute(
        path: '/auth/setup',
        builder: (context, state) => const ProfileSetupScreen(),
      ),
      GoRoute(
        path: '/home',
        builder: (context, state) => const HomeScreen(),
        redirect: (context, state) async {
          final code = await storage.read(key: _pendingInviteCodeKey);
          if (code != null && code.isNotEmpty) {
            await storage.delete(key: _pendingInviteCodeKey);
            return '/groups/join?code=$code';
          }
          return null;
        },
      ),
      GoRoute(
        path: '/groups/create',
        builder: (context, state) => const CreateGroupScreen(),
      ),
      GoRoute(
        path: '/groups/join',
        builder: (context, state) {
          final code = state.uri.queryParameters['code'] ??
              state.extra as String?;
          return JoinGroupScreen(initialCode: code);
        },
      ),
      GoRoute(
        path: '/groups/:id',
        builder: (context, state) {
          final id = state.pathParameters['id']!;
          return GroupScreen(groupId: id);
        },
      ),
      // Deeplink: /join/:code — redirect handles auth-gating via _RouterNotifier
      GoRoute(
        path: '/join/:code',
        redirect: (context, state) {
          final code = state.pathParameters['code'] ?? '';
          return '/groups/join?code=$code';
        },
      ),
    ],
  );
});
