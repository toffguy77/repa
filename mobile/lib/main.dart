import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/providers/auth_provider.dart';
import 'core/router/app_router.dart';
import 'core/services/push_service.dart';
import 'core/theme/app_theme.dart';
import 'core/widgets/offline_banner.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const ProviderScope(child: RepaApp()));
}

class RepaApp extends ConsumerStatefulWidget {
  const RepaApp({super.key});

  @override
  ConsumerState<RepaApp> createState() => _RepaAppState();
}

class _RepaAppState extends ConsumerState<RepaApp> {
  bool _pushInitialized = false;

  @override
  void initState() {
    super.initState();
    ref.read(authProvider.notifier).checkAuth();
    ref.listenManual(authProvider, (prev, next) {
      if (!_pushInitialized && next.status == AuthStatus.authenticated) {
        _pushInitialized = true;
        ref.read(pushServiceProvider).init();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'Repa',
      theme: AppTheme.light,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
      builder: (context, child) => OfflineBanner(child: child!),
    );
  }
}
