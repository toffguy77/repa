import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers/auth_provider.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_text_styles.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Text('\u{1F351}', style: TextStyle(fontSize: 64)),
            const SizedBox(height: 16),
            Text('Добро пожаловать в Репу!', style: AppTextStyles.headline2),
            const SizedBox(height: 8),
            Text('Экран групп появится в T07', style: AppTextStyles.caption),
            const SizedBox(height: 32),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 48),
              child: OutlinedButton(
                onPressed: () => ref.read(authProvider.notifier).logout(),
                child: const Text(
                  'Выйти',
                  style: TextStyle(color: AppColors.error),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
