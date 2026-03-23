import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../api/api_service.dart';
import '../providers/api_provider.dart';
import '../router/app_router.dart';

@pragma('vm:entry-point')
Future<void> _firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp();
}

final pushServiceProvider = Provider<PushService>((ref) {
  final api = ref.watch(apiServiceProvider);
  final router = ref.watch(routerProvider);
  return PushService(api, router);
});

class PushService {
  final ApiService _api;
  final GoRouter _router;

  PushService(this._api, this._router);

  Future<void> init() async {
    await Firebase.initializeApp();
    FirebaseMessaging.onBackgroundMessage(_firebaseMessagingBackgroundHandler);

    final settings = await FirebaseMessaging.instance.requestPermission(
      alert: true,
      badge: true,
      sound: true,
    );

    if (settings.authorizationStatus == AuthorizationStatus.authorized ||
        settings.authorizationStatus == AuthorizationStatus.provisional) {
      await _registerToken();
    }

    // Foreground messages — handled in-app (snackbar), not system notification
    FirebaseMessaging.onMessage.listen(_handleForeground);

    // Tap on notification when app is in background
    FirebaseMessaging.onMessageOpenedApp.listen(_handleTap);

    // Tap on notification when app was terminated
    final initialMessage =
        await FirebaseMessaging.instance.getInitialMessage();
    if (initialMessage != null) {
      _handleTap(initialMessage);
    }

    // Token refresh
    FirebaseMessaging.instance.onTokenRefresh.listen((token) {
      _registerTokenValue(token);
    });
  }

  Future<void> _registerToken() async {
    try {
      final token = await FirebaseMessaging.instance.getToken();
      if (token != null) {
        await _registerTokenValue(token);
      }
    } catch (e) {
      debugPrint('FCM token registration failed: $e');
    }
  }

  Future<void> _registerTokenValue(String token) async {
    try {
      final platform =
          defaultTargetPlatform == TargetPlatform.iOS ? 'ios' : 'android';
      await _api.registerFCMToken(token, platform);
    } catch (e) {
      debugPrint('FCM token API call failed: $e');
    }
  }

  void _handleForeground(RemoteMessage message) {
    // Foreground push — in-app snackbar is handled by the UI layer
    // listening to a stream. For MVP, we just log it.
    debugPrint('FCM foreground: ${message.data}');
  }

  void _handleTap(RemoteMessage message) {
    final data = message.data;
    final screen = data['screen'] as String?;
    final groupId = data['groupId'] as String?;
    final seasonId = data['seasonId'] as String?;

    if (screen == null) return;

    switch (screen) {
      case 'reveal':
        if (groupId != null && seasonId != null) {
          _router.go('/groups/$groupId/reveal/$seasonId');
        }
      case 'reveal-waiting':
        if (groupId != null) {
          _router.go('/groups/$groupId');
        }
      case 'vote':
        if (groupId != null && seasonId != null) {
          _router.go('/groups/$groupId/vote/$seasonId');
        }
      case 'question-vote':
        if (groupId != null) {
          _router.go('/groups/$groupId');
        }
      case 'shop':
        _router.go('/shop');
    }
  }
}
