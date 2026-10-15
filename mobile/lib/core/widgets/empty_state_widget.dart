import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';

import '../theme/app_tokens.dart';
import 'kit/kit.dart';

class EmptyStateWidget extends StatelessWidget {
  final String emoji;
  final String title;
  final String subtitle;
  final String? buttonText;
  final VoidCallback? onButtonPressed;

  const EmptyStateWidget({
    super.key,
    required this.emoji,
    required this.title,
    required this.subtitle,
    this.buttonText,
    this.onButtonPressed,
  });

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    return Center(
      child: Padding(
        padding: EdgeInsets.symmetric(horizontal: AppTokens.space.xxl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(emoji, style: const TextStyle(fontSize: 64)).animate().scale(
                  begin: const Offset(0.5, 0.5),
                  end: const Offset(1, 1),
                  duration: context.motion(MotionClass.emphasis),
                  curve: context.motionCurve(MotionClass.emphasis),
                ),
            SizedBox(height: AppTokens.space.lg),
            Text(
              title,
              style: AppTokens.text.heading2.copyWith(color: t.color.textPrimary),
              textAlign: TextAlign.center,
            ),
            SizedBox(height: AppTokens.space.sm),
            Text(
              subtitle,
              style: AppTokens.text.body.copyWith(color: t.color.textSecondary),
              textAlign: TextAlign.center,
            ),
            if (buttonText != null && onButtonPressed != null) ...[
              SizedBox(height: AppTokens.space.xl),
              AppButton(label: buttonText!, onPressed: onButtonPressed),
            ],
          ],
        ),
      ),
    );
  }
}
