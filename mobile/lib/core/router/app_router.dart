import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../providers/auth_provider.dart';
import '../../features/auth/presentation/phone_screen.dart';
import '../../features/auth/presentation/otp_screen.dart';
import '../../features/auth/presentation/profile_setup_screen.dart';
import '../../features/home/home_screen.dart';

class _RouterNotifier extends ChangeNotifier {
  AuthState _authState = const AuthState();

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

    if (_authState.status == AuthStatus.unknown) return null;

    if (!isAuth && !isAuthRoute) return '/auth/phone';

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
  final notifier = _RouterNotifier();
  ref.listen(authProvider, (_, next) {
    notifier.update(next);
  });
  return notifier;
});

final routerProvider = Provider<GoRouter>((ref) {
  final notifier = ref.watch(_routerNotifierProvider);

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
      ),
    ],
  );
});
