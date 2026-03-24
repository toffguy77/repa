import 'package:flutter/material.dart';
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
import '../../features/voting/presentation/voting_screen.dart';
import '../../features/voting/presentation/voting_complete_screen.dart';
import '../../features/reveal/presentation/reveal_screen.dart';
import '../../features/reveal/presentation/members_reveal_screen.dart';
import '../../features/profile/presentation/member_profile_screen.dart';
import '../../features/crystals/presentation/crystals_shop_screen.dart';
import '../../features/telegram/presentation/telegram_setup_screen.dart';
import '../../features/onboarding/presentation/onboarding_screen.dart';
import '../../features/question_vote/presentation/question_vote_screen.dart';

const _pendingInviteCodeKey = 'pending_invite_code';

// --- Page transitions ---

CustomTransitionPage<void> _slideFromRight(GoRouterState state, Widget child) {
  return CustomTransitionPage(
    key: state.pageKey,
    child: child,
    transitionsBuilder: (context, animation, secondaryAnimation, child) {
      final tween = Tween(begin: const Offset(1, 0), end: Offset.zero)
          .chain(CurveTween(curve: Curves.easeOutCubic));
      return SlideTransition(position: animation.drive(tween), child: child);
    },
  );
}

CustomTransitionPage<void> _slideFromBottom(GoRouterState state, Widget child) {
  return CustomTransitionPage(
    key: state.pageKey,
    child: child,
    transitionsBuilder: (context, animation, secondaryAnimation, child) {
      final tween = Tween(begin: const Offset(0, 1), end: Offset.zero)
          .chain(CurveTween(curve: Curves.easeOutCubic));
      return SlideTransition(position: animation.drive(tween), child: child);
    },
  );
}

CustomTransitionPage<void> _fadeTransition(GoRouterState state, Widget child) {
  return CustomTransitionPage(
    key: state.pageKey,
    child: child,
    transitionsBuilder: (context, animation, secondaryAnimation, child) {
      return FadeTransition(opacity: animation, child: child);
    },
  );
}

// --- Router notifier ---

class _RouterNotifier extends ChangeNotifier {
  final FlutterSecureStorage _storage;
  AuthState _authState = const AuthState();

  _RouterNotifier(this._storage);

  void update(AuthState authState) {
    final shouldNotify = _authState.status != authState.status ||
        _authState.needsProfileSetup != authState.needsProfileSetup ||
        _authState.isNewUser != authState.isNewUser;
    _authState = authState;
    if (shouldNotify) {
      notifyListeners();
    }
  }

  String? redirect(GoRouterState state) {
    final isAuth = _authState.status == AuthStatus.authenticated;
    final isAuthRoute = state.matchedLocation.startsWith('/auth');
    final isDeeplink = state.matchedLocation.startsWith('/join/');
    final isOnboarding = state.matchedLocation == '/onboarding';

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

    if (isAuth && _authState.isNewUser && !isOnboarding) {
      return '/onboarding';
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
      // Auth — fade transitions
      GoRoute(
        path: '/auth/phone',
        pageBuilder: (context, state) =>
            _fadeTransition(state, const PhoneScreen()),
      ),
      GoRoute(
        path: '/auth/otp',
        pageBuilder: (context, state) {
          final phone = state.extra as String? ?? '';
          return _fadeTransition(state, OtpScreen(phone: phone));
        },
      ),
      GoRoute(
        path: '/auth/setup',
        pageBuilder: (context, state) =>
            _fadeTransition(state, const ProfileSetupScreen()),
      ),

      // Onboarding — fade
      GoRoute(
        path: '/onboarding',
        pageBuilder: (context, state) =>
            _fadeTransition(state, const OnboardingScreen()),
      ),

      // Home — fade in
      GoRoute(
        path: '/home',
        pageBuilder: (context, state) =>
            _fadeTransition(state, const HomeScreen()),
        redirect: (context, state) async {
          final code = await storage.read(key: _pendingInviteCodeKey);
          if (code != null && code.isNotEmpty) {
            await storage.delete(key: _pendingInviteCodeKey);
            return '/groups/join?code=$code';
          }
          return null;
        },
      ),

      // Groups — slide from bottom (modal-style)
      GoRoute(
        path: '/groups/create',
        pageBuilder: (context, state) =>
            _slideFromBottom(state, const CreateGroupScreen()),
      ),
      GoRoute(
        path: '/groups/join',
        pageBuilder: (context, state) {
          final code = state.uri.queryParameters['code'] ??
              state.extra as String?;
          return _slideFromBottom(state, JoinGroupScreen(initialCode: code));
        },
      ),

      // Group detail — slide from right (drill-down)
      GoRoute(
        path: '/groups/:id',
        pageBuilder: (context, state) {
          final id = state.pathParameters['id']!;
          return _slideFromRight(state, GroupScreen(groupId: id));
        },
        routes: [
          GoRoute(
            path: 'members/:userId',
            pageBuilder: (context, state) {
              final groupId = state.pathParameters['id']!;
              final userId = state.pathParameters['userId']!;
              return _slideFromRight(
                state,
                MemberProfileScreen(groupId: groupId, userId: userId),
              );
            },
          ),
          GoRoute(
            path: 'question-vote',
            pageBuilder: (context, state) {
              final groupId = state.pathParameters['id']!;
              return _slideFromBottom(
                state,
                QuestionVoteScreen(groupId: groupId),
              );
            },
          ),
          GoRoute(
            path: 'telegram',
            pageBuilder: (context, state) {
              final groupId = state.pathParameters['id']!;
              return _slideFromRight(
                  state, TelegramSetupScreen(groupId: groupId));
            },
          ),
          GoRoute(
            path: 'reveal/:seasonId',
            pageBuilder: (context, state) {
              final groupId = state.pathParameters['id']!;
              final seasonId = state.pathParameters['seasonId']!;
              final status =
                  state.uri.queryParameters['status'] ?? 'REVEALED';
              return _slideFromBottom(
                state,
                RevealScreen(
                  groupId: groupId,
                  seasonId: seasonId,
                  seasonStatus: status,
                ),
              );
            },
            routes: [
              GoRoute(
                path: 'members',
                pageBuilder: (context, state) {
                  final groupId = state.pathParameters['id']!;
                  final seasonId = state.pathParameters['seasonId']!;
                  final status =
                      state.uri.queryParameters['status'] ?? 'REVEALED';
                  return _slideFromRight(
                    state,
                    MembersRevealScreen(
                      groupId: groupId,
                      seasonId: seasonId,
                      seasonStatus: status,
                    ),
                  );
                },
              ),
            ],
          ),
          GoRoute(
            path: 'vote/:seasonId',
            pageBuilder: (context, state) {
              final groupId = state.pathParameters['id']!;
              final seasonId = state.pathParameters['seasonId']!;
              return _slideFromBottom(
                state,
                VotingScreen(groupId: groupId, seasonId: seasonId),
              );
            },
            routes: [
              GoRoute(
                path: 'complete',
                pageBuilder: (context, state) {
                  final groupId = state.pathParameters['id']!;
                  final seasonId = state.pathParameters['seasonId']!;
                  return _fadeTransition(
                    state,
                    VotingCompleteScreen(
                        groupId: groupId, seasonId: seasonId),
                  );
                },
              ),
            ],
          ),
        ],
      ),

      // Shop — slide from bottom (modal)
      GoRoute(
        path: '/shop',
        pageBuilder: (context, state) =>
            _slideFromBottom(state, const CrystalsShopScreen()),
      ),

      // Deeplink redirect
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
