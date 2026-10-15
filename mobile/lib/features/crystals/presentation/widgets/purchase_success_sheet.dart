import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_tokens.dart';


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
      decoration: BoxDecoration(
        color: context.t.elevation.level2.surface,
        borderRadius: AppTokens.radius.sheet,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: context.t.color.surface,
              borderRadius: BorderRadius.circular(AppTokens.radius.xs),
            ),
          ),
          const SizedBox(height: 24),
          const Text('\u{1F48E}', style: TextStyle(fontSize: 56))
              .animate()
              .scale(
                begin: const Offset(0.3, 0.3),
                end: const Offset(1.0, 1.0),
                duration: context.motion(MotionClass.emphasis),
                curve: Curves.elasticOut,
              )
              .fadeIn(duration: context.motion(MotionClass.surface)),
          const SizedBox(height: 16),
          Text(
            '+$amount кристаллов',
            style: context.ts.heading1.copyWith(color: context.t.color.accent),
          )
              .animate()
              .fadeIn(delay: AppTokens.motion.stagger(3), duration: context.motion(MotionClass.surface))
              .slideY(begin: 0.2),
          const SizedBox(height: 8),
          Text(
            'Баланс: $newBalance',
            style: context.ts.bodySecondary,
          ).animate().fadeIn(delay: AppTokens.motion.stagger(7), duration: context.motion(MotionClass.surface)),
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
