import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../core/theme/app_tokens.dart';
import '../../../core/widgets/kit/kit.dart';

class ForceUpdateScreen extends StatelessWidget {
  const ForceUpdateScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Text('🍆', style: TextStyle(fontSize: 64)),
              const SizedBox(height: 24),
              Text(
                'Обновите приложение',
                style: context.ts.heading2,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 12),
              Text(
                'Текущая версия больше не поддерживается. '
                'Обновите Репу, чтобы продолжить.',
                style: context.ts.bodySecondary,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 32),
              AppButton(
                label: 'Обновить',
                onPressed: () {
                  // TODO: replace with actual App Store / Play Store URLs
                  launchUrl(
                    Uri.parse('https://repa.app/download'),
                    mode: LaunchMode.externalApplication,
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}
