import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';

class PurchaseSuccessSheet extends StatelessWidget {
  final int amount;
  final int newBalance;
  final VoidCallback onDone;

  const PurchaseSuccessSheet({
    super.key,
    required this.amount,
    required this.newBalance,
    required this.onDone,
  });

  @override
  Widget build(BuildContext context) {
    HapticFeedback.heavyImpact();

    return Container(
      padding: const EdgeInsets.all(24),
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: AppColors.surface,
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          const SizedBox(height: 24),
          const Text('\u{1F48E}', style: TextStyle(fontSize: 56))
              .animate()
              .scale(
                begin: const Offset(0.3, 0.3),
                end: const Offset(1.0, 1.0),
                duration: 400.ms,
                curve: Curves.elasticOut,
              )
              .fadeIn(duration: 200.ms),
          const SizedBox(height: 16),
          Text(
            '+$amount кристаллов',
            style: AppTextStyles.headline1.copyWith(color: AppColors.primary),
          )
              .animate()
              .fadeIn(delay: 200.ms, duration: 300.ms)
              .slideY(begin: 0.2),
          const SizedBox(height: 8),
          Text(
            'Баланс: $newBalance',
            style: AppTextStyles.bodySecondary,
          ).animate().fadeIn(delay: 400.ms, duration: 300.ms),
          const SizedBox(height: 32),
          SizedBox(
            width: double.infinity,
            height: 52,
            child: ElevatedButton(
              onPressed: () {
                HapticFeedback.mediumImpact();
                onDone();
              },
              child: const Text('Отлично'),
            ),
          ),
          const SizedBox(height: 16),
        ],
      ),
    );
  }
}
