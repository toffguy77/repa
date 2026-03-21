import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/providers/auth_provider.dart';
import 'core/router/app_router.dart';
import 'core/theme/app_theme.dart';

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
  @override
  void initState() {
    super.initState();
    ref.read(authProvider.notifier).checkAuth();
  }

  @override
  Widget build(BuildContext context) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'Repa',
      theme: AppTheme.light,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
